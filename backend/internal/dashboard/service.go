package dashboard

import (
	"context"
	"time"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// Today returns today's job counts and collection using the server's local
// timezone. MVP keeps this simple; per-business timezones can be layered on.
func (s *Service) Today(ctx context.Context, businessID string) (*Today, error) {
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	end := start.AddDate(0, 0, 1)
	dateStr := now.Format("2006-01-02")

	// Use whichever is later for completed-based visibility: jobs created
	// today keep being counted even when completed.
	counts, err := s.repo.JobsCounts(ctx, businessID, start, end)
	if err != nil {
		return nil, err
	}

	// For today's collection we want paid_at in [today, tomorrow).
	collection, err := s.repo.CollectionTotals(ctx, businessID, start, end)
	if err != nil {
		return nil, err
	}

	return &Today{
		Date:       dateStr,
		Jobs:       *counts,
		Collection: *collection,
	}, nil
}