package receipts

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"

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

const columns = `id, business_id, job_id, receipt_number, public_token, issued_at, created_at`

func scanReceipt(row pgx.Row) (*Receipt, error) {
	var r Receipt
	err := row.Scan(&r.ID, &r.BusinessID, &r.JobID, &r.ReceiptNumber, &r.PublicToken, &r.IssuedAt, &r.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

// GeneratePublicToken returns a URL-safe random token.
func GeneratePublicToken() string {
	b := make([]byte, 18)
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

// Create allocates a receipt number and inserts the record inside tx.
func (r *Repository) CreateTx(ctx context.Context, tx database.Querier, businessID, jobID string) (*Receipt, error) {
	var seq int64
	if err := tx.QueryRow(ctx, `SELECT nextval('receipt_number_seq')`).Scan(&seq); err != nil {
		return nil, err
	}
	rec := &Receipt{
		ID:            uuid.NewString(),
		BusinessID:    businessID,
		JobID:         jobID,
		ReceiptNumber: fmt.Sprintf("RCP-%09d", seq),
		PublicToken:   GeneratePublicToken(),
	}
	row := tx.QueryRow(ctx, `
		INSERT INTO receipts (id, business_id, job_id, receipt_number, public_token)
		VALUES ($1,$2,$3,$4,$5) RETURNING `+columns,
		rec.ID, rec.BusinessID, rec.JobID, rec.ReceiptNumber, rec.PublicToken)
	return scanReceipt(row)
}

func (r *Repository) GetByJob(ctx context.Context, businessID, jobID string) (*Receipt, error) {
	row := r.db.QueryRow(ctx,
		`SELECT `+columns+` FROM receipts WHERE business_id=$1 AND job_id=$2 ORDER BY created_at DESC LIMIT 1`,
		businessID, jobID)
	rec, err := scanReceipt(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, httpapi.NewAPIError(404, "RESOURCE_NOT_FOUND", "No receipt for this job yet.")
	}
	return rec, err
}

func (r *Repository) GetByToken(ctx context.Context, token string) (*Receipt, error) {
	row := r.db.QueryRow(ctx, `SELECT `+columns+` FROM receipts WHERE public_token = $1`, token)
	rec, err := scanReceipt(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, httpapi.NewAPIError(404, "RESOURCE_NOT_FOUND", "Receipt not found.")
	}
	return rec, err
}