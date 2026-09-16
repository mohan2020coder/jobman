package jobs

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jobman/backend/database"
	"github.com/jobman/backend/internal/customers"
	"github.com/jobman/backend/pkg/httpapi"
)

type Repository struct {
	db database.Querier
}

func NewRepository(db database.Querier) *Repository {
	return &Repository{db: db}
}

const jobColumns = `
	j.id, j.business_id, j.customer_id, j.technician_id, j.job_number, j.service_type,
	j.problem_description, j.address, j.latitude, j.longitude, j.status,
	j.estimated_amount, j.final_amount, j.scheduled_at, j.accepted_at, j.on_the_way_at,
	j.started_at, j.completed_at, j.notes, j.created_by, j.created_at, j.updated_at`

// jobColumnsPlain is the same list without the table alias, for INSERT ... RETURNING.
const jobColumnsPlain = `
	id, business_id, customer_id, technician_id, job_number, service_type,
	problem_description, address, latitude, longitude, status,
	estimated_amount, final_amount, scheduled_at, accepted_at, on_the_way_at,
	started_at, completed_at, notes, created_by, created_at, updated_at`

// scanJob scans a job row (aggregate columns resolved separately).
func scanJob(row pgx.Row) (*Job, error) {
	var j Job
	err := row.Scan(
		&j.ID, &j.BusinessID, &j.CustomerID, &j.TechnicianID, &j.JobNumber, &j.ServiceType,
		&j.ProblemDescription, &j.Address, &j.Latitude, &j.Longitude, &j.Status,
		&j.EstimatedAmount, &j.FinalAmount, &j.ScheduledAt, &j.AcceptedAt, &j.OnTheWayAt,
		&j.StartedAt, &j.CompletedAt, &j.Notes, &j.CreatedBy, &j.CreatedAt, &j.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &j, nil
}

const jobDetailSelect = `
	SELECT ` + jobColumns + `,
		c.name AS customer_name, c.phone AS customer_phone, c.address AS customer_address,
		t.id AS technician_id_out, t.status AS technician_status, u.name AS technician_name, u.phone AS technician_phone
	FROM jobs j
	JOIN customers c ON c.id = j.customer_id
	LEFT JOIN technicians t ON t.id = j.technician_id
	LEFT JOIN users u ON u.id = t.user_id`

func scanJobDetail(row pgx.Row) (*Job, error) {
	var j Job
	var custName, custPhone string
	var custAddr *string
	var techID, techStatus, techName, techPhone *string
	err := row.Scan(
		&j.ID, &j.BusinessID, &j.CustomerID, &j.TechnicianID, &j.JobNumber, &j.ServiceType,
		&j.ProblemDescription, &j.Address, &j.Latitude, &j.Longitude, &j.Status,
		&j.EstimatedAmount, &j.FinalAmount, &j.ScheduledAt, &j.AcceptedAt, &j.OnTheWayAt,
		&j.StartedAt, &j.CompletedAt, &j.Notes, &j.CreatedBy, &j.CreatedAt, &j.UpdatedAt,
		&custName, &custPhone, &custAddr,
		&techID, &techStatus, &techName, &techPhone,
	)
	if err != nil {
		return nil, err
	}
	j.Customer = &customers.Customer{
		ID:      j.CustomerID,
		Name:    custName,
		Phone:   custPhone,
		Address: custAddr,
	}
	if techID != nil {
		j.Technician = &TechnicianLite{ID: *techID, Name: *techName, Phone: stringPtrDefault(techPhone), Status: *techStatus}
	}
	return &j, nil
}

func stringPtrDefault(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

// NextJobNumber allocates the next sequential job number.
func (r *Repository) NextJobNumber(ctx context.Context) (string, error) {
	var n int64
	if err := r.db.QueryRow(ctx, `SELECT nextval('job_number_seq')`).Scan(&n); err != nil {
		return "", err
	}
	return fmt.Sprintf("JB-%06d", n), nil
}

func (r *Repository) Create(ctx context.Context, j *Job) (*Job, error) {
	j.ID = uuid.NewString()
	row := r.db.QueryRow(ctx, `
		INSERT INTO jobs (id, business_id, customer_id, technician_id, job_number, service_type,
			problem_description, address, latitude, longitude, status, estimated_amount, scheduled_at, created_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
		RETURNING `+jobColumnsPlain,
		j.ID, j.BusinessID, j.CustomerID, j.TechnicianID, j.JobNumber, j.ServiceType,
		j.ProblemDescription, j.Address, j.Latitude, j.Longitude, j.Status,
		j.EstimatedAmount, j.ScheduledAt, j.CreatedBy)
	return scanJob(row)
}

func (r *Repository) GetByID(ctx context.Context, businessID, id string) (*Job, error) {
	row := r.db.QueryRow(ctx,
		jobDetailSelect+` WHERE j.id = $1 AND j.business_id = $2`, id, businessID)
	j, err := scanJobDetail(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, httpapi.NewAPIError(404, "RESOURCE_NOT_FOUND", "Job not found.")
	}
	if err != nil {
		return nil, err
	}
	return j, nil
}

// GetForUpdate returns a full job row (no joins) for state transitions.
func (r *Repository) GetForUpdate(ctx context.Context, businessID, id string) (*Job, error) {
	row := r.db.QueryRow(ctx,
		`SELECT `+jobColumns+` FROM jobs j WHERE j.id = $1 AND j.business_id = $2`, id, businessID)
	j, err := scanJob(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, httpapi.NewAPIError(404, "RESOURCE_NOT_FOUND", "Job not found.")
	}
	if err != nil {
		return nil, err
	}
	return j, nil
}

// VerifyCustomer confirms the customer exists within the business.
func (r *Repository) VerifyCustomer(ctx context.Context, businessID, customerID string) error {
	var one int
	err := r.db.QueryRow(ctx,
		`SELECT 1 FROM customers WHERE id=$1 AND business_id=$2`, customerID, businessID).Scan(&one)
	if errors.Is(err, pgx.ErrNoRows) {
		return httpapi.NewAPIError(404, "RESOURCE_NOT_FOUND", "Customer not found.")
	}
	return err
}

func (r *Repository) UpdateDetails(ctx context.Context, j *Job) (*Job, error) {
	row := r.db.QueryRow(ctx, `
		UPDATE jobs
		SET customer_id=$3, technician_id=$4, service_type=$5, problem_description=$6,
		    address=$7, latitude=$8, longitude=$9, estimated_amount=$10, scheduled_at=$11,
		    notes=$12, updated_at=NOW()
		WHERE id=$1 AND business_id=$2
		RETURNING `+jobColumnsPlain,
		j.ID, j.BusinessID, j.CustomerID, j.TechnicianID, j.ServiceType, j.ProblemDescription,
		j.Address, j.Latitude, j.Longitude, j.EstimatedAmount, j.ScheduledAt, j.Notes)
	updated, err := scanJob(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, httpapi.NewAPIError(404, "RESOURCE_NOT_FOUND", "Job not found.")
	}
	return updated, err
}

// Transition sets the new status, stamps the transition timestamp, and inserts
// a status history row. Runs atomically within the passed transaction.
func transitionInTx(ctx context.Context, tx pgx.Tx, jobID, businessID, oldStatus, newStatus, changedBy string) error {
	if ts := transitionTimestamp(newStatus); ts != "" {
		_, err := tx.Exec(ctx,
			`UPDATE jobs SET status=$1, updated_at=NOW(), `+ts+`=NOW() WHERE id=$2 AND business_id=$3`,
			newStatus, jobID, businessID)
		if err != nil {
			return err
		}
	} else {
		_, err := tx.Exec(ctx,
			`UPDATE jobs SET status=$1, updated_at=NOW() WHERE id=$2 AND business_id=$3`,
			newStatus, jobID, businessID)
		if err != nil {
			return err
		}
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO job_status_history (id, business_id, job_id, old_status, new_status, changed_by)
		VALUES ($1,$2,$3,$4,$5,$6)`,
		uuid.NewString(), businessID, jobID, strPtr(oldStatus), newStatus, changedBy); err != nil {
		return err
	}
	return nil
}

func strPtr(s string) *string { return &s }

// SetFinalAmount completes a job with final amount and notes inside a tx.
func setFinalAmountInTx(ctx context.Context, tx pgx.Tx, businessID, jobID string, finalAmount float64, notes *string) error {
	_, err := tx.Exec(ctx,
		`UPDATE jobs SET final_amount=$3, notes=COALESCE($4, notes), updated_at=NOW() WHERE id=$1 AND business_id=$2`,
		jobID, businessID, finalAmount, notes)
	return err
}

func (r *Repository) List(ctx context.Context, businessID string, f ListFilters) (*ListResult, error) {
	where := `WHERE j.business_id = $1`
	args := []any{businessID}
	param := 2

	if f.Status != "" {
		where += ` AND j.status = $` + strconv.Itoa(param)
		args = append(args, f.Status)
		param++
	}
	if f.TechnicianID != "" {
		where += ` AND j.technician_id = $` + strconv.Itoa(param)
		args = append(args, f.TechnicianID)
		param++
	}
	if f.CustomerID != "" {
		where += ` AND j.customer_id = $` + strconv.Itoa(param)
		args = append(args, f.CustomerID)
		param++
	}
	if f.Date != "" {
		if start, end, err := dateRange(f.Date); err == nil {
			where += ` AND j.scheduled_at >= $` + strconv.Itoa(param) + ` AND j.scheduled_at < $` + strconv.Itoa(param+1)
			args = append(args, start, end)
			param += 2
		}
	}
	if s := strings.TrimSpace(f.Search); s != "" {
		where += ` AND (c.name ILIKE $` + strconv.Itoa(param) + ` OR j.service_type ILIKE $` + strconv.Itoa(param) + ` OR j.job_number ILIKE $` + strconv.Itoa(param) + `)`
		args = append(args, "%"+s+"%")
		param++
	}

	var total int
	if err := r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM jobs j JOIN customers c ON c.id=j.customer_id `+where, args...).Scan(&total); err != nil {
		return nil, err
	}

	orderBy := `j.created_at DESC`
	if f.SortBy != "" {
		switch f.SortBy {
		case "scheduled_at":
			orderBy = `j.scheduled_at ASC NULLS LAST`
		case "job_number":
			orderBy = `j.job_number ASC`
		}
	}

	where += ` ORDER BY ` + orderBy + ` LIMIT $` + strconv.Itoa(param) + ` OFFSET $` + strconv.Itoa(param+1)
	args = append(args, f.Limit, httpapi.OffsetFrom(f.Page, f.Limit))

	rows, err := r.db.Query(ctx, jobDetailSelect+` `+where, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*Job
	for rows.Next() {
		j, err := scanJobDetail(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, j)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return &ListResult{Total: total, Data: out}, nil
}

// ListByTechnician returns the technician's own jobs. Filter is one of
// "today", "upcoming", "completed", or empty for all.
func (r *Repository) ListByTechnician(ctx context.Context, businessID, technicianID, filter string) ([]*Job, error) {
	where := `WHERE j.business_id = $1 AND j.technician_id = $2`
	args := []any{businessID, technicianID}
	param := 3

	switch filter {
	case "today":
		start, end, err := dateRange(time.Now().Format("2006-01-02"))
		if err == nil {
			where += ` AND j.scheduled_at >= $` + strconv.Itoa(param) + ` AND j.scheduled_at < $` + strconv.Itoa(param+1)
			args = append(args, start, end)
		}
		where += ` AND j.status NOT IN ('COMPLETED','CANCELLED')`
	case "upcoming":
		where += ` AND j.status NOT IN ('COMPLETED','CANCELLED') AND j.scheduled_at >= NOW()`
	case "completed":
		where += ` AND j.status IN ('COMPLETED','CANCELLED')`
	}

	where += ` ORDER BY j.scheduled_at ASC NULLS LAST`

	rows, err := r.db.Query(ctx, jobDetailSelect+` `+where, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*Job
	for rows.Next() {
		j, err := scanJobDetail(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, j)
	}
	return out, rows.Err()
}

func (r *Repository) ListByCustomer(ctx context.Context, businessID, customerID string) ([]*Job, error) {
	rows, err := r.db.Query(ctx,
		jobDetailSelect+` WHERE j.business_id=$1 AND j.customer_id=$2 ORDER BY j.created_at DESC`,
		businessID, customerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*Job
	for rows.Next() {
		j, err := scanJobDetail(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, j)
	}
	return out, rows.Err()
}

// SumPaidByJobs returns job_id -> total paid for the given jobs.
func (r *Repository) SumPaidByJobs(ctx context.Context, businessID string, jobIDs []string) (map[string]float64, error) {
	out := make(map[string]float64, len(jobIDs))
	if len(jobIDs) == 0 {
		return out, nil
	}
	args := make([]any, 0, len(jobIDs)+1)
	args = append(args, businessID)
	placeholders := make([]string, 0, len(jobIDs))
	for i, id := range jobIDs {
		placeholders = append(placeholders, `$`+strconv.Itoa(i+2))
		args = append(args, id)
	}
	query := `SELECT job_id, COALESCE(SUM(amount),0) FROM payments
		WHERE business_id=$1 AND status='PAID' AND job_id IN (` + strings.Join(placeholders, ",") + `)
		GROUP BY job_id`
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var jobID string
		var sum float64
		if err := rows.Scan(&jobID, &sum); err != nil {
			return nil, err
		}
		out[jobID] = sum
	}
	return out, rows.Err()
}

// GetItems returns the items attached to a job.
func (r *Repository) GetItems(ctx context.Context, businessID, jobID string) ([]*Item, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, business_id, job_id, description, quantity, unit_price, total_price, created_at
		FROM job_items WHERE business_id=$1 AND job_id=$2 ORDER BY created_at ASC`, businessID, jobID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*Item
	for rows.Next() {
		var it Item
		if err := rows.Scan(&it.ID, &it.BusinessID, &it.JobID, &it.Description, &it.Quantity, &it.UnitPrice, &it.TotalPrice, &it.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, &it)
	}
	return out, rows.Err()
}

func (r *Repository) GetStatusHistory(ctx context.Context, businessID, jobID string) ([]*StatusHistory, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, old_status, new_status, changed_by, created_at
		FROM job_status_history WHERE business_id=$1 AND job_id=$2 ORDER BY created_at ASC`, businessID, jobID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*StatusHistory
	for rows.Next() {
		var h StatusHistory
		if err := rows.Scan(&h.ID, &h.OldStatus, &h.NewStatus, &h.ChangedBy, &h.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, &h)
	}
	return out, rows.Err()
}

// dateRange returns the inclusive [start, end) range for a local date (YYYY-MM-DD).
func dateRange(date string) (time.Time, time.Time, error) {
	start, err := time.ParseInLocation("2006-01-02", date, time.Local)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	return start, start.AddDate(0, 0, 1), nil
}

type ListFilters struct {
	Status       string
	TechnicianID string
	CustomerID   string
	Date         string
	Search       string
	SortBy       string
	Page         int
	Limit        int
}

type ListResult struct {
	Data  []*Job
	Total int
}