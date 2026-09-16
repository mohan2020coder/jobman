package users

import (
	"context"
	"errors"
	"time"

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

const columns = `id, business_id, name, phone, email, password_hash, role, is_active, last_login_at, created_at, updated_at`

func scanUser(row pgx.Row) (*User, error) {
	var u User
	err := row.Scan(
		&u.ID, &u.BusinessID, &u.Name, &u.Phone, &u.Email, &u.PasswordHash,
		&u.Role, &u.IsActive, &u.LastLoginAt, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *Repository) Create(ctx context.Context, u *User) (*User, error) {
	u.ID = uuid.NewString()
	query := `INSERT INTO users (id, business_id, name, phone, email, password_hash, role, is_active)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		RETURNING ` + columns
	row := r.db.QueryRow(ctx, query,
		u.ID, u.BusinessID, u.Name, u.Phone, u.Email, u.PasswordHash, u.Role, u.IsActive)
	created, err := scanUser(row)
	if err != nil {
		if database.IsUniqueViolation(err) {
			return nil, httpapi.NewAPIError(409, "USER_PHONE_EXISTS", "A user with this phone already exists.")
		}
		return nil, err
	}
	return created, nil
}

func (r *Repository) GetByID(ctx context.Context, id string) (*User, error) {
	row := r.db.QueryRow(ctx, `SELECT `+columns+` FROM users WHERE id = $1`, id)
	u, err := scanUser(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, httpapi.NewAPIError(404, "RESOURCE_NOT_FOUND", "User not found.")
	}
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (r *Repository) GetByBusinessPhone(ctx context.Context, businessID, phone string) (*User, error) {
	row := r.db.QueryRow(ctx,
		`SELECT `+columns+` FROM users WHERE business_id = $1 AND phone = $2`,
		businessID, phone)
	u, err := scanUser(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, httpapi.NewAPIError(404, "RESOURCE_NOT_FOUND", "User not found.")
	}
	if err != nil {
		return nil, err
	}
	return u, nil
}

// FindByPhoneAcrossBusinesses locates a user by phone across all businesses,
// used only by login before tenant identity is known.
func (r *Repository) FindByPhoneAcrossBusinesses(ctx context.Context, phone string) (*User, error) {
	row := r.db.QueryRow(ctx, `SELECT `+columns+` FROM users WHERE phone = $1 LIMIT 1`, phone)
	u, err := scanUser(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, httpapi.NewAPIError(401, "AUTH_INVALID_CREDENTIALS", "Invalid phone or password.")
	}
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (r *Repository) UpdateLastLogin(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, `UPDATE users SET last_login_at = $1, updated_at = $1 WHERE id = $2`, time.Now(), id)
	return err
}

func (r *Repository) UpdateDetails(ctx context.Context, u *User) (*User, error) {
	query := `UPDATE users SET name = $1, email = $2, is_active = $3, updated_at = NOW()
		WHERE id = $4 AND business_id = $5
		RETURNING ` + columns
	row := r.db.QueryRow(ctx, query, u.Name, u.Email, u.IsActive, u.ID, u.BusinessID)
	updated, err := scanUser(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, httpapi.NewAPIError(404, "RESOURCE_NOT_FOUND", "User not found.")
	}
	if err != nil {
		return nil, err
	}
	return updated, nil
}

func (r *Repository) SetPassword(ctx context.Context, id, passwordHash string) error {
	_, err := r.db.Exec(ctx, `UPDATE users SET password_hash = $1, updated_at = NOW() WHERE id = $2`, passwordHash, id)
	return err
}

func (r *Repository) SetActive(ctx context.Context, id string, active bool) error {
	_, err := r.db.Exec(ctx, `UPDATE users SET is_active = $1, updated_at = NOW() WHERE id = $2`, active, id)
	return err
}

func (r *Repository) ListByBusiness(ctx context.Context, businessID string) ([]*User, error) {
	rows, err := r.db.Query(ctx,
		`SELECT `+columns+` FROM users WHERE business_id = $1 ORDER BY created_at DESC`, businessID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*User
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	return out, rows.Err()
}