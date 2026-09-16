package business

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

const columns = `id, name, phone, email, address, timezone, currency, created_at, updated_at`

func scanBusiness(row pgx.Row) (*Business, error) {
	var b Business
	err := row.Scan(
		&b.ID, &b.Name, &b.Phone, &b.Email, &b.Address,
		&b.Timezone, &b.Currency, &b.CreatedAt, &b.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *Repository) Create(ctx context.Context, b *Business) (*Business, error) {
	b.ID = uuid.NewString()
	row := r.db.QueryRow(ctx,
		`INSERT INTO businesses (id, name, phone, email, address, timezone, currency)
		 VALUES ($1,$2,$3,$4,$5,'Asia/Kolkata','INR') RETURNING `+columns,
		b.ID, b.Name, b.Phone, b.Email, b.Address)
	return scanBusiness(row)
}

func (r *Repository) GetByID(ctx context.Context, id string) (*Business, error) {
	row := r.db.QueryRow(ctx, `SELECT `+columns+` FROM businesses WHERE id = $1`, id)
	b, err := scanBusiness(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, httpapi.NewAPIError(404, "RESOURCE_NOT_FOUND", "Business not found.")
	}
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (r *Repository) Update(ctx context.Context, b *Business) (*Business, error) {
	row := r.db.QueryRow(ctx,
		`UPDATE businesses SET name=$1, phone=$2, email=$3, address=$4, timezone=$5, currency=$6, updated_at=NOW()
		 WHERE id=$7 RETURNING `+columns,
		b.Name, b.Phone, b.Email, b.Address, b.Timezone, b.Currency, b.ID)
	updated, err := scanBusiness(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, httpapi.NewAPIError(404, "RESOURCE_NOT_FOUND", "Business not found.")
	}
	return updated, err
}