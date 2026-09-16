package customers

import (
	"context"
	"strings"

	"github.com/jobman/backend/pkg/httpapi"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

type CreateRequest struct {
	Name      string   `json:"name"`
	Phone     string   `json:"phone"`
	Email     *string  `json:"email"`
	Address   *string  `json:"address"`
	Latitude  *float64 `json:"latitude"`
	Longitude *float64 `json:"longitude"`
	Notes     *string  `json:"notes"`
}

type UpdateRequest struct {
	Name      *string  `json:"name"`
	Phone     *string  `json:"phone"`
	Email     *string  `json:"email"`
	Address   *string  `json:"address"`
	Latitude  *float64 `json:"latitude"`
	Longitude *float64 `json:"longitude"`
	Notes     *string  `json:"notes"`
}

func (s *Service) Create(ctx context.Context, businessID string, req CreateRequest) (*Customer, error) {
	req.Name = strings.TrimSpace(req.Name)
	req.Phone = strings.TrimSpace(req.Phone)
	if req.Name == "" {
		return nil, validationError("name is required.")
	}
	if req.Phone == "" {
		return nil, validationError("phone is required.")
	}

	c := &Customer{
		BusinessID: businessID,
		Name:       req.Name,
		Phone:      req.Phone,
		Email:      req.Email,
		Address:    req.Address,
		Latitude:   req.Latitude,
		Longitude:  req.Longitude,
		Notes:      req.Notes,
	}
	return s.repo.Create(ctx, c)
}

func (s *Service) Get(ctx context.Context, businessID, id string) (*Customer, error) {
	return s.repo.GetByID(ctx, businessID, id)
}

func (s *Service) Update(ctx context.Context, businessID, id string, req UpdateRequest) (*Customer, error) {
	existing, err := s.repo.GetByID(ctx, businessID, id)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		if n := strings.TrimSpace(*req.Name); n != "" {
			existing.Name = n
		}
	}
	if req.Phone != nil {
		existing.Phone = strings.TrimSpace(*req.Phone)
	}
	if req.Email != nil {
		existing.Email = req.Email
	}
	if req.Address != nil {
		existing.Address = req.Address
	}
	if req.Latitude != nil {
		existing.Latitude = req.Latitude
	}
	if req.Longitude != nil {
		existing.Longitude = req.Longitude
	}
	if req.Notes != nil {
		existing.Notes = req.Notes
	}

	if existing.Name == "" || existing.Phone == "" {
		return nil, validationError("name and phone are required.")
	}

	return s.repo.Update(ctx, existing)
}

func (s *Service) Delete(ctx context.Context, businessID, id string) error {
	return s.repo.Delete(ctx, businessID, id)
}

func (s *Service) List(ctx context.Context, businessID string, f ListFilters) (*ListResult, error) {
	return s.repo.List(ctx, businessID, f)
}

func validationError(msg string) error {
	return httpapi.NewAPIError(400, "VALIDATION_ERROR", msg)
}