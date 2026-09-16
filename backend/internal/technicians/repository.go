package technicians

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jobman/backend/database"
	"github.com/jobman/backend/pkg/httpapi"
)

const columns = `id, business_id, user_id, status, created_at, updated_at`

type rowScanner interface {
	Scan(dest ...any) error
}

func scanTech(row rowScanner) (*Technician, error) {
	var t Technician
	err := row.Scan(&t.ID, &t.BusinessID, &t.UserID, &t.Status, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// Repository handles technician + user rows together.
type Repository struct {
	db database.Querier
}

func NewRepository(db database.Querier) *Repository {
	return &Repository{db: db}
}

const joinColumns = `
	t.id, t.business_id, t.user_id, t.status, t.created_at, t.updated_at,
	u.name, u.phone, u.email, u.is_active, u.last_login_at`

func scanTechnicianWithUser(row pgx.Row) (*Technician, error) {
	var t Technician
	err := row.Scan(
		&t.ID, &t.BusinessID, &t.UserID, &t.Status, &t.CreatedAt, &t.UpdatedAt,
		&t.Name, &t.Phone, &t.Email, &t.IsActive, &t.LastLogin,
	)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// CreatePermissionTxn is a callback that prepares a user within the same
// transaction as the technician row. For MVP the service performs two steps
// inside one pool transaction instead.

func (r *Repository) Create(ctx context.Context, t *Technician) (*Technician, error) {
	t.ID = uuid.NewString()
	row := r.db.QueryRow(ctx,
		`INSERT INTO technicians (id, business_id, user_id, status)
		 VALUES ($1,$2,$3,$4) RETURNING `+columns,
		t.ID, t.BusinessID, t.UserID, t.Status)
	return scanTech(row)
}

func (r *Repository) GetByUserID(ctx context.Context, businessID, userID string) (*Technician, error) {
	row := r.db.QueryRow(ctx,
		`SELECT `+joinColumns+` FROM technicians t
		 JOIN users u ON u.id = t.user_id
		 WHERE t.user_id = $1 AND t.business_id = $2`, userID, businessID)
	tech, err := scanTechnicianWithUser(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, httpapi.NewAPIError(404, "RESOURCE_NOT_FOUND", "Technician not found.")
	}
	if err != nil {
		return nil, err
	}
	return tech, nil
}

func (r *Repository) GetByID(ctx context.Context, businessID, id string) (*Technician, error) {
	row := r.db.QueryRow(ctx,
		`SELECT `+joinColumns+` FROM technicians t
		 JOIN users u ON u.id = t.user_id
		 WHERE t.id = $1 AND t.business_id = $2`, id, businessID)
	tech, err := scanTechnicianWithUser(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, httpapi.NewAPIError(404, "RESOURCE_NOT_FOUND", "Technician not found.")
	}
	if err != nil {
		return nil, err
	}
	return tech, nil
}

func (r *Repository) UpdateStatus(ctx context.Context, businessID, id, status string) error {
	tag, err := r.db.Exec(ctx,
		`UPDATE technicians SET status = $3, updated_at = NOW() WHERE id = $1 AND business_id = $2`,
		id, businessID, status)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return httpapi.NewAPIError(404, "RESOURCE_NOT_FOUND", "Technician not found.")
	}
	return nil
}

func (r *Repository) List(ctx context.Context, businessID string) ([]*Technician, error) {
	rows, err := r.db.Query(ctx,
		`SELECT `+joinColumns+` FROM technicians t
		 JOIN users u ON u.id = t.user_id
		 WHERE t.business_id = $1
		 ORDER BY u.name ASC`, businessID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*Technician
	for rows.Next() {
		var t Technician
		if err := rows.Scan(
			&t.ID, &t.BusinessID, &t.UserID, &t.Status, &t.CreatedAt, &t.UpdatedAt,
			&t.Name, &t.Phone, &t.Email, &t.IsActive, &t.LastLogin,
		); err != nil {
			return nil, err
		}
		out = append(out, &t)
	}
	return out, rows.Err()
}