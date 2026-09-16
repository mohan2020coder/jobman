package auth

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jobman/backend/internal/business"
	"github.com/jobman/backend/internal/users"
	"github.com/jobman/backend/pkg/httpapi"
)

type Config struct {
	JWTSecret     []byte
	JWTExpiration time.Duration
}

type Service struct {
	pool     *pgxpool.Pool
	users    *users.Repository
	business *business.Repository
	cfg      Config
}

func NewService(pool *pgxpool.Pool, usersRepo *users.Repository, businessRepo *business.Repository, cfg Config) *Service {
	return &Service{pool: pool, users: usersRepo, business: businessRepo, cfg: cfg}
}

// Register creates a business and its owner atomically.
func (s *Service) Register(ctx context.Context, req RegisterRequest) (*AuthResponse, error) {
	req.BusinessName = strings.TrimSpace(req.BusinessName)
	req.OwnerName = strings.TrimSpace(req.OwnerName)
	req.Phone = strings.TrimSpace(req.Phone)

	if req.BusinessName == "" {
		return nil, httpapi.NewAPIError(400, "VALIDATION_ERROR", "business_name is required.")
	}
	if req.OwnerName == "" {
		return nil, httpapi.NewAPIError(400, "VALIDATION_ERROR", "owner_name is required.")
	}
	if req.Phone == "" || len(req.Phone) < 6 || len(req.Phone) > 20 {
		return nil, httpapi.NewAPIError(400, "VALIDATION_ERROR", "A valid phone number is required.")
	}
	if len(req.Password) < 6 {
		return nil, httpapi.NewAPIError(400, "VALIDATION_ERROR", "Password must be at least 6 characters.")
	}

	if _, err := s.users.FindByPhoneAcrossBusinesses(ctx, req.Phone); err == nil {
		return nil, httpapi.NewAPIError(409, "USER_PHONE_EXISTS", "An account with this phone already exists.")
	}

	hash, err := HashPassword(req.Password)
	if err != nil {
		log.Printf("auth: hash password: %v", err)
		return nil, httpapi.NewAPIError(500, "PASSWORD_HASH_FAILED", "Failed to secure password.")
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	txUsers := users.NewRepository(tx)
	txBiz := business.NewRepository(tx)

	b, err := txBiz.Create(ctx, &business.Business{
		Name:  req.BusinessName,
		Phone: stringPtrOrNil(req.Phone),
		Email: stringPtrOrNil(req.Email),
	})
	if err != nil {
		return nil, err
	}

	owner, err := txUsers.Create(ctx, &users.User{
		BusinessID:   b.ID,
		Name:         req.OwnerName,
		Phone:        req.Phone,
		Email:        stringPtrOrNil(req.Email),
		PasswordHash: hash,
		Role:         users.RoleOwner,
		IsActive:     true,
	})
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	token, err := Sign(s.cfg.JWTSecret, s.cfg.JWTExpiration, owner.ID, b.ID, owner.Role)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{
		Token:    token,
		User:     PublicUser{ID: owner.ID, Name: owner.Name, Role: owner.Role, BusinessID: b.ID},
		Business: PublicBusiness{ID: b.ID, Name: b.Name},
	}, nil
}

func (s *Service) Login(ctx context.Context, req LoginRequest) (*AuthResponse, error) {
	req.Phone = strings.TrimSpace(req.Phone)
	if req.Phone == "" || req.Password == "" {
		return nil, httpapi.NewAPIError(400, "VALIDATION_ERROR", "Phone and password are required.")
	}

	u, err := s.users.FindByPhoneAcrossBusinesses(ctx, req.Phone)
	if err != nil {
		return nil, httpapi.NewAPIError(401, "AUTH_INVALID_CREDENTIALS", "Invalid phone or password.")
	}
	if !u.IsActive {
		return nil, httpapi.NewAPIError(403, "USER_INACTIVE", "This account is inactive. Contact the business owner.")
	}
	if !CheckPassword(u.PasswordHash, req.Password) {
		return nil, httpapi.NewAPIError(401, "AUTH_INVALID_CREDENTIALS", "Invalid phone or password.")
	}

	_ = s.users.UpdateLastLogin(ctx, u.ID)

	b, err := s.business.GetByID(ctx, u.BusinessID)
	if err != nil {
		return nil, err
	}

	token, err := Sign(s.cfg.JWTSecret, s.cfg.JWTExpiration, u.ID, b.ID, u.Role)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{
		Token:    token,
		User:     PublicUser{ID: u.ID, Name: u.Name, Role: u.Role, BusinessID: b.ID},
		Business: PublicBusiness{ID: b.ID, Name: b.Name},
	}, nil
}

func (s *Service) Me(ctx context.Context, principal *Principal) (*MeResponse, error) {
	u, err := s.users.GetByID(ctx, principal.UserID)
	if err != nil {
		return nil, err
	}
	b, err := s.business.GetByID(ctx, principal.BusinessID)
	if err != nil {
		return nil, err
	}
	return &MeResponse{
		User:     PublicUser{ID: u.ID, Name: u.Name, Role: u.Role, BusinessID: u.BusinessID},
		Business: PublicBusiness{ID: b.ID, Name: b.Name},
	}, nil
}

func stringPtrOrNil(s string) *string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	return &s
}