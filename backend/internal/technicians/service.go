package technicians

import (
	"context"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jobman/backend/internal/users"
	"github.com/jobman/backend/pkg/httpapi"
)

type Service struct {
	pool  *pgxpool.Pool
	repo  *Repository
	users *users.Repository
}

func NewService(pool *pgxpool.Pool, repo *Repository, usersRepo *users.Repository) *Service {
	return &Service{pool: pool, repo: repo, users: usersRepo}
}

type CreateRequest struct {
	Name     string `json:"name"`
	Phone    string `json:"phone"`
	Password string `json:"password"`
}

type UpdateRequest struct {
	Name     string `json:"name"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`
	Status   string `json:"status"`
	IsActive *bool  `json:"is_active"`
}

func (s *Service) Create(ctx context.Context, businessID string, req CreateRequest) (*Technician, error) {
	req.Name = strings.TrimSpace(req.Name)
	req.Phone = strings.TrimSpace(req.Phone)
	if req.Name == "" || req.Phone == "" {
		return nil, httpapi.NewAPIError(400, "VALIDATION_ERROR", "Name and phone are required.")
	}
	if len(req.Password) < 6 {
		return nil, httpapi.NewAPIError(400, "VALIDATION_ERROR", "Password must be at least 6 characters.")
	}

	hash, err := hashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	txUsers := users.NewRepository(tx)
	txRepo := NewRepository(tx)

	u, err := txUsers.Create(ctx, &users.User{
		BusinessID:   businessID,
		Name:         req.Name,
		Phone:        req.Phone,
		PasswordHash: hash,
		Role:         users.RoleTechnician,
		IsActive:     true,
	})
	if err != nil {
		return nil, err
	}

	_, err = txRepo.Create(ctx, &Technician{
		BusinessID: businessID,
		UserID:     u.ID,
		Status:     StatusAvailable,
	})
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return s.repo.GetByUserID(ctx, businessID, u.ID)
}

func (s *Service) Get(ctx context.Context, businessID, id string) (*Technician, error) {
	return s.repo.GetByID(ctx, businessID, id)
}

// GetByUserID resolves the technician row for the authenticated technician
// user (used to enforce "only assigned technician" rules in jobs).
func (s *Service) GetByUserID(ctx context.Context, businessID, userID string) (*Technician, error) {
	return s.repo.GetByUserID(ctx, businessID, userID)
}

func (s *Service) Update(ctx context.Context, businessID, id string, req UpdateRequest) (*Technician, error) {
	tech, err := s.repo.GetByID(ctx, businessID, id)
	if err != nil {
		return nil, err
	}
	user, err := s.users.GetByID(ctx, tech.UserID)
	if err != nil {
		return nil, err
	}

	if strings.TrimSpace(req.Name) != "" {
		user.Name = strings.TrimSpace(req.Name)
	}
	if strings.TrimSpace(req.Phone) != "" {
		user.Phone = strings.TrimSpace(req.Phone)
	}
	if req.Status != "" {
		if !validStatus(req.Status) {
			return nil, httpapi.NewAPIError(400, "VALIDATION_ERROR", "Invalid technician status.")
		}
		tech.Status = req.Status
	}
	if req.IsActive != nil {
		user.IsActive = *req.IsActive
	}
	// Empty-status technicians are marked inactive via IS_ACTIVE false.
	if req.Status == StatusInactive || (req.IsActive != nil && !*req.IsActive) {
		user.IsActive = false
	}

	if _, err := s.users.UpdateDetails(ctx, user); err != nil {
		return nil, err
	}
	if err := s.repo.UpdateStatus(ctx, businessID, id, tech.Status); err != nil {
		return nil, err
	}
	return s.repo.GetByID(ctx, businessID, id)
}

func (s *Service) Delete(ctx context.Context, businessID, id string) error {
	tech, err := s.repo.GetByID(ctx, businessID, id)
	if err != nil {
		return err
	}
	// Soft-deactivate rather than hard delete: jobs keep their FK references.
	if err := s.repo.UpdateStatus(ctx, businessID, id, StatusInactive); err != nil {
		return err
	}
	return s.users.SetActive(ctx, tech.UserID, false)
}

func (s *Service) List(ctx context.Context, businessID string) ([]*Technician, error) {
	return s.repo.List(ctx, businessID)
}

func validStatus(s string) bool {
	return s == StatusAvailable || s == StatusBusy || s == StatusInactive
}