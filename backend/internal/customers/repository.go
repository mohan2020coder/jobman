package customers

import (
	"context"
	"errors"
	"strconv"
	"strings"

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

const columns = `id, business_id, name, phone, email, address, latitude, longitude, notes, created_at, updated_at`

func scanCustomer(row pgx.Row) (*Customer, error) {
	var c Customer
	err := row.Scan(
		&c.ID, &c.BusinessID, &c.Name, &c.Phone, &c.Email, &c.Address,
		&c.Latitude, &c.Longitude, &c.Notes, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *Repository) Create(ctx context.Context, c *Customer) (*Customer, error) {
	c.ID = uuid.NewString()
	row := r.db.QueryRow(ctx,
		`INSERT INTO customers (id, business_id, name, phone, email, address, latitude, longitude, notes)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING `+columns,
		c.ID, c.BusinessID, c.Name, c.Phone, c.Email, c.Address, c.Latitude, c.Longitude, c.Notes)
	created, err := scanCustomer(row)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func (r *Repository) GetByID(ctx context.Context, businessID, id string) (*Customer, error) {
	row := r.db.QueryRow(ctx,
		`SELECT `+columns+` FROM customers WHERE id = $1 AND business_id = $2`, id, businessID)
	c, err := scanCustomer(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, httpapi.NewAPIError(404, "RESOURCE_NOT_FOUND", "Customer not found.")
	}
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (r *Repository) Update(ctx context.Context, c *Customer) (*Customer, error) {
	row := r.db.QueryRow(ctx,
		`UPDATE customers
		 SET name=$3, phone=$4, email=$5, address=$6, latitude=$7, longitude=$8, notes=$9, updated_at=NOW()
		 WHERE id=$1 AND business_id=$2
		 RETURNING `+columns,
		c.ID, c.BusinessID, c.Name, c.Phone, c.Email, c.Address, c.Latitude, c.Longitude, c.Notes)
	updated, err := scanCustomer(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, httpapi.NewAPIError(404, "RESOURCE_NOT_FOUND", "Customer not found.")
	}
	if err != nil {
		return nil, err
	}
	return updated, nil
}

func (r *Repository) Delete(ctx context.Context, businessID, id string) error {
	tag, err := r.db.Exec(ctx,
		`DELETE FROM customers WHERE id = $1 AND business_id = $2`, id, businessID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return httpapi.NewAPIError(404, "RESOURCE_NOT_FOUND", "Customer not found.")
	}
	return nil
}

type ListFilters struct {
	Search string
	Page   int
	Limit  int
}

type ListResult struct {
	Data  []*Customer
	Total int
}

func (r *Repository) List(ctx context.Context, businessID string, f ListFilters) (*ListResult, error) {
	where := `WHERE business_id = $1`
	args := []any{businessID}
	param := 2

	if s := strings.TrimSpace(f.Search); s != "" {
		where += ` AND (name ILIKE $` + strconv.Itoa(param) + ` OR phone ILIKE $` + strconv.Itoa(param) + `)`
		args = append(args, "%"+s+"%")
		param++
	}

	var total int
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM customers `+where, args...).Scan(&total); err != nil {
		return nil, err
	}

	where += ` ORDER BY created_at DESC LIMIT $` + strconv.Itoa(param) + ` OFFSET $` + strconv.Itoa(param+1)
	args = append(args, f.Limit, httpapi.OffsetFrom(f.Page, f.Limit))

	rows, err := r.db.Query(ctx, `SELECT `+columns+` FROM customers `+where, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*Customer
	for rows.Next() {
		c, err := scanCustomer(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &ListResult{Data: out, Total: total}, nil
}