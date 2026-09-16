package business

import (
	"context"
	"strings"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

type UpdateRequest struct {
	Name     string  `json:"name"`
	Phone    *string `json:"phone"`
	Email    *string `json:"email"`
	Address  *string `json:"address"`
	Timezone string  `json:"timezone"`
	Currency string  `json:"currency"`
}

func (s *Service) Get(ctx context.Context, id string) (*Business, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) Update(ctx context.Context, id string, req UpdateRequest) (*Business, error) {
	b, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if name := strings.TrimSpace(req.Name); name != "" {
		b.Name = name
	}
	if req.Phone != nil {
		b.Phone = req.Phone
	}
	if req.Email != nil {
		b.Email = req.Email
	}
	if req.Address != nil {
		b.Address = req.Address
	}
	if req.Timezone != "" {
		b.Timezone = req.Timezone
	}
	if req.Currency != "" {
		b.Currency = req.Currency
	}
	return s.repo.Update(ctx, b)
}