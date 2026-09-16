package payments

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jobman/backend/database"
	"github.com/jobman/backend/pkg/httpapi"
)

type Repository struct {
	db database.Querier
}

func NewRepository(db database.Querier) *Repository {
	return &Repository{db: db}
}

const columns = `id, business_id, job_id, amount, method, status, transaction_reference, paid_at, created_by, created_at`

func scanPayment(row pgx.Row) (*Payment, error) {
	var p Payment
	err := row.Scan(
		&p.ID, &p.BusinessID, &p.JobID, &p.Amount, &p.Method, &p.Status,
		&p.TransactionReference, &p.PaidAt, &p.CreatedBy, &p.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func ValidMethod(m string) bool {
	switch m {
	case MethodCash, MethodUPI, MethodCard, MethodBankTransfer, MethodOther:
		return true
	}
	return false
}

func (r *Repository) Create(ctx context.Context, p *Payment) (*Payment, error) {
	p.ID = uuid.NewString()
	row := r.db.QueryRow(ctx, `
		INSERT INTO payments (id, business_id, job_id, amount, method, status, transaction_reference, created_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8) RETURNING `+columns,
		p.ID, p.BusinessID, p.JobID, p.Amount, p.Method, p.Status, p.TransactionReference, p.CreatedBy)
	return scanPayment(row)
}

func (r *Repository) ListByJob(ctx context.Context, businessID, jobID string) ([]*Payment, error) {
	rows, err := r.db.Query(ctx,
		`SELECT `+columns+` FROM payments WHERE business_id=$1 AND job_id=$2 ORDER BY paid_at ASC`,
		businessID, jobID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*Payment
	for rows.Next() {
		p, err := scanPayment(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (r *Repository) SumPaidByJob(ctx context.Context, businessID, jobID string) (float64, error) {
	var sum float64
	err := r.db.QueryRow(ctx,
		`SELECT COALESCE(SUM(amount),0) FROM payments WHERE business_id=$1 AND job_id=$2 AND status='PAID'`,
		businessID, jobID).Scan(&sum)
	if err != nil {
		return 0, err
	}
	return sum, nil
}

// VerifyJobAccess confirms the job belongs to the business (used as a cheap
// tenant check before payment writes).
func (r *Repository) VerifyJobAccess(ctx context.Context, businessID, jobID string) error {
	var one int
	err := r.db.QueryRow(ctx,
		`SELECT 1 FROM jobs WHERE id=$1 AND business_id=$2`, jobID, businessID).Scan(&one)
	if errors.Is(err, pgx.ErrNoRows) {
		return httpapi.NewAPIError(404, "RESOURCE_NOT_FOUND", "Job not found.")
	}
	return err
}

// GetJobAmounts returns (final_amount, estimated_amount) for a job.
func (r *Repository) GetJobAmounts(ctx context.Context, businessID, jobID string) (final, estimated float64, err error) {
	err = r.db.QueryRow(ctx,
		`SELECT final_amount, estimated_amount FROM jobs WHERE id=$1 AND business_id=$2`,
		jobID, businessID).Scan(&final, &estimated)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, 0, httpapi.NewAPIError(404, "RESOURCE_NOT_FOUND", "Job not found.")
	}
	return final, estimated, err
}