package dashboard

import (
	"context"
	"time"

	"github.com/jobman/backend/database"
)

type Repository struct {
	db database.Querier
}

func NewRepository(db database.Querier) *Repository {
	return &Repository{db: db}
}

type JobCounts struct {
	Total     int `json:"total"`
	Pending   int `json:"pending"`
	Accepted  int `json:"accepted"`
	OnTheWay  int `json:"on_the_way"`
	Started   int `json:"started"`
	Completed int `json:"completed"`
	Cancelled int `json:"cancelled"`
}

type Collection struct {
	Total        float64 `json:"total"`
	Cash         float64 `json:"cash"`
	UPI          float64 `json:"upi"`
	Card         float64 `json:"card"`
	BankTransfer float64 `json:"bank_transfer"`
	Other        float64 `json:"other"`
}

type Today struct {
	Date       string     `json:"date"`
	Jobs       JobCounts  `json:"jobs"`
	Collection Collection `json:"collection"`
}

func (r *Repository) JobsCounts(ctx context.Context, businessID string, start, end time.Time) (*JobCounts, error) {
	var c JobCounts
	err := r.db.QueryRow(ctx, `
		SELECT
			COUNT(*) AS total,
			COUNT(*) FILTER (WHERE status = 'PENDING') AS pending,
			COUNT(*) FILTER (WHERE status = 'ACCEPTED') AS accepted,
			COUNT(*) FILTER (WHERE status = 'ON_THE_WAY') AS on_the_way,
			COUNT(*) FILTER (WHERE status = 'STARTED') AS started,
			COUNT(*) FILTER (WHERE status = 'COMPLETED') AS completed,
			COUNT(*) FILTER (WHERE status = 'CANCELLED') AS cancelled
		FROM jobs
		WHERE business_id = $1 AND created_at >= $2 AND created_at < $3`,
		businessID, start, end).Scan(
		&c.Total, &c.Pending, &c.Accepted, &c.OnTheWay, &c.Started, &c.Completed, &c.Cancelled)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *Repository) CollectionTotals(ctx context.Context, businessID string, start, end time.Time) (*Collection, error) {
	rows, err := r.db.Query(ctx, `
		SELECT method, COALESCE(SUM(amount),0)
		FROM payments
		WHERE business_id = $1 AND status = 'PAID' AND paid_at >= $2 AND paid_at < $3
		GROUP BY method`, businessID, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var c Collection
	for rows.Next() {
		var method string
		var sum float64
		if err := rows.Scan(&method, &sum); err != nil {
			return nil, err
		}
		c.Total += sum
		switch method {
		case "CASH":
			c.Cash = sum
		case "UPI":
			c.UPI = sum
		case "CARD":
			c.Card = sum
		case "BANK_TRANSFER":
			c.BankTransfer = sum
		default:
			c.Other = sum
		}
	}
	return &c, rows.Err()
}

// DayRange returns the local [start, end) pair for a YYYY-MM-DD date.
func DayRange(date string) (time.Time, time.Time, error) {
	if date == "" {
		return time.Time{}, time.Time{}, nil
	}
	start, err := time.ParseInLocation("2006-01-02", date, time.Local)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	return start, start.AddDate(0, 0, 1), nil
}