package jobs

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jobman/backend/database"
	"github.com/jobman/backend/internal/notifications"
	"github.com/jobman/backend/internal/payments"
	"github.com/jobman/backend/internal/technicians"
	"github.com/jobman/backend/pkg/httpapi"
)

// ReceiptIssuer creates a receipt inside the complete transaction.
// Implemented in main.go by an adapter over the receipts package.
type ReceiptIssuer interface {
	IssueTx(ctx context.Context, tx database.Querier, businessID, jobID string) (*IssuedReceipt, error)
}

type IssuedReceipt struct {
	ID            string
	ReceiptNumber string
	PublicToken   string
	IssuedAt      time.Time
}

type Service struct {
	pool     *pgxpool.Pool
	repo     *Repository
	payments *payments.Repository
	techSvc  *technicians.Service
	notifs   notifications.Service
	issuer   ReceiptIssuer
}

func NewService(pool *pgxpool.Pool, repo *Repository, paymentsRepo *payments.Repository, techSvc *technicians.Service, notifs notifications.Service) *Service {
	return &Service{pool: pool, repo: repo, payments: paymentsRepo, techSvc: techSvc, notifs: notifs}
}

func (s *Service) SetReceiptIssuer(issuer ReceiptIssuer) { s.issuer = issuer }

type CreateRequest struct {
	CustomerID         string   `json:"customer_id"`
	TechnicianID       *string  `json:"technician_id"`
	ServiceType        string   `json:"service_type"`
	ProblemDescription *string  `json:"problem_description"`
	Address            *string  `json:"address"`
	Latitude           *float64 `json:"latitude"`
	Longitude          *float64 `json:"longitude"`
	EstimatedAmount    float64  `json:"estimated_amount"`
	ScheduledAt        *string  `json:"scheduled_at"` // RFC3339
}

type UpdateRequest struct {
	CustomerID         *string  `json:"customer_id"`
	TechnicianID       *string  `json:"technician_id"`
	ServiceType        *string  `json:"service_type"`
	ProblemDescription *string  `json:"problem_description"`
	Address            *string  `json:"address"`
	EstimatedAmount    *float64 `json:"estimated_amount"`
	ScheduledAt        *string  `json:"scheduled_at"`
	Notes              *string  `json:"notes"`
}

func (s *Service) Create(ctx context.Context, businessID, createdBy string, req CreateRequest) (*Job, error) {
	req.ServiceType = strings.TrimSpace(req.ServiceType)
	if req.CustomerID == "" {
		return nil, httpapi.NewAPIError(400, "VALIDATION_ERROR", "customer_id is required.")
	}
	if req.ServiceType == "" {
		return nil, httpapi.NewAPIError(400, "VALIDATION_ERROR", "service_type is required.")
	}

	if err := s.repo.VerifyCustomer(ctx, businessID, req.CustomerID); err != nil {
		return nil, err
	}
	var techID *string
	if req.TechnicianID != nil && *req.TechnicianID != "" {
		tech, err := s.techSvc.Get(ctx, businessID, *req.TechnicianID)
		if err != nil {
			return nil, err
		}
		if tech.Status == "INACTIVE" {
			return nil, httpapi.NewAPIError(422, "JOB_INVALID_ASSIGNMENT", "Cannot assign an inactive technician.")
		}
		techID = &tech.ID
	}

	scheduled, err := parseOptionalTime(req.ScheduledAt)
	if err != nil {
		return nil, httpapi.NewAPIError(400, "VALIDATION_ERROR", "scheduled_at must be a valid RFC3339 timestamp.")
	}

	number, err := s.repo.NextJobNumber(ctx)
	if err != nil {
		return nil, err
	}

	created, err := s.repo.Create(ctx, &Job{
		BusinessID:         businessID,
		CustomerID:         req.CustomerID,
		TechnicianID:       techID,
		JobNumber:          number,
		ServiceType:        req.ServiceType,
		ProblemDescription: req.ProblemDescription,
		Address:            req.Address,
		Latitude:           req.Latitude,
		Longitude:          req.Longitude,
		Status:             StatusPending,
		EstimatedAmount:    req.EstimatedAmount,
		ScheduledAt:        scheduled,
		CreatedBy:          createdBy,
	})
	if err != nil {
		return nil, err
	}
	if techID != nil {
		s.publishAssigned(ctx, businessID, created.ID, *techID)
	}
	return created, nil
}

// publishAssigned re-reads the new job (with customer/technician joined) and
// pushes it to the assigned technician's live connections.
func (s *Service) publishAssigned(ctx context.Context, businessID, jobID, techID string) {
	full, err := s.repo.GetByID(ctx, businessID, jobID)
	if err != nil {
		return
	}
	ev := notifications.JobEvent{
		BusinessID: businessID, JobID: full.ID, JobNumber: full.JobNumber,
		TechnicianID: techID, Status: full.Status, ServiceType: full.ServiceType,
		EstimatedAmount: full.EstimatedAmount, ScheduledAt: full.ScheduledAt, Address: full.Address,
	}
	if full.Customer != nil {
		ev.CustomerName = full.Customer.Name
	}
	s.notifs.NotifyJobAssigned(ctx, ev)
}

// jobEvent builds a notification event carrying the assigned technician plus
// display fields for the assigned technician's devices.
func (s *Service) jobEvent(ctx context.Context, businessID, jobID, status string) notifications.JobEvent {
	full, err := s.repo.GetByID(ctx, businessID, jobID)
	if err != nil {
		return notifications.JobEvent{BusinessID: businessID, JobID: jobID, Status: status}
	}
	ev := notifications.JobEvent{
		BusinessID: businessID, JobID: full.ID, JobNumber: full.JobNumber,
		Status: status, ServiceType: full.ServiceType,
		EstimatedAmount: full.EstimatedAmount, ScheduledAt: full.ScheduledAt, Address: full.Address,
	}
	if full.TechnicianID != nil {
		ev.TechnicianID = *full.TechnicianID
	}
	if full.Customer != nil {
		ev.CustomerName = full.Customer.Name
	}
	return ev
}

func (s *Service) Update(ctx context.Context, businessID, id string, req UpdateRequest) (*Job, error) {
	existing, err := s.repo.GetForUpdate(ctx, businessID, id)
	if err != nil {
		return nil, err
	}
	if existing.Status != StatusPending {
		return nil, httpapi.NewAPIError(422, "JOB_INVALID_STATUS", "Jobs can only be edited while PENDING.")
	}

	if req.CustomerID != nil && *req.CustomerID != "" {
		if err := s.repo.VerifyCustomer(ctx, businessID, *req.CustomerID); err != nil {
			return nil, err
		}
		existing.CustomerID = *req.CustomerID
	}
	if req.TechnicianID != nil {
		if *req.TechnicianID == "" {
			existing.TechnicianID = nil
		} else {
			tech, err := s.techSvc.Get(ctx, businessID, *req.TechnicianID)
			if err != nil {
				return nil, err
			}
			existing.TechnicianID = &tech.ID
		}
	}
	if req.ServiceType != nil {
		existing.ServiceType = strings.TrimSpace(*req.ServiceType)
	}
	if req.ProblemDescription != nil {
		existing.ProblemDescription = req.ProblemDescription
	}
	if req.Address != nil {
		existing.Address = req.Address
	}
	if req.EstimatedAmount != nil {
		existing.EstimatedAmount = *req.EstimatedAmount
	}
	if req.ScheduledAt != nil {
		t, err := parseOptionalTime(req.ScheduledAt)
		if err != nil {
			return nil, httpapi.NewAPIError(400, "VALIDATION_ERROR", "scheduled_at must be a valid RFC3339 timestamp.")
		}
		existing.ScheduledAt = t
	}
	if req.Notes != nil {
		existing.Notes = req.Notes
	}

	return s.repo.UpdateDetails(ctx, existing)
}

func (s *Service) Get(ctx context.Context, businessID, id string) (*Job, error) {
	job, err := s.repo.GetByID(ctx, businessID, id)
	if err != nil {
		return nil, err
	}
	return s.decorate(ctx, businessID, job)
}

func (s *Service) decorate(ctx context.Context, businessID string, job *Job) (*Job, error) {
	items, err := s.repo.GetItems(ctx, businessID, job.ID)
	if err != nil {
		return nil, err
	}
	paymentsList, err := s.payments.ListByJob(ctx, businessID, job.ID)
	if err != nil {
		return nil, err
	}
	history, err := s.repo.GetStatusHistory(ctx, businessID, job.ID)
	if err != nil {
		return nil, err
	}
	totalPaid, err := s.payments.SumPaidByJob(ctx, businessID, job.ID)
	if err != nil {
		return nil, err
	}

	job.Items = items
	job.Payments = paymentsList
	job.StatusLog = history
	job.TotalPaid = totalPaid
	job.PaymentText = payments.PaymentStatusText(totalPaid, job.FinalAmount)
	return job, nil
}

func (s *Service) List(ctx context.Context, businessID string, f ListFilters) (*ListResult, error) {
	res, err := s.repo.List(ctx, businessID, f)
	if err != nil {
		return nil, err
	}
	if len(res.Data) == 0 {
		return res, nil
	}
	ids := make([]string, 0, len(res.Data))
	for _, j := range res.Data {
		ids = append(ids, j.ID)
	}
	sums, err := s.repo.SumPaidByJobs(ctx, businessID, ids)
	if err != nil {
		return nil, err
	}
	for _, j := range res.Data {
		paid := sums[j.ID]
		j.TotalPaid = paid
		j.PaymentText = payments.PaymentStatusText(paid, j.FinalAmount)
	}
	return res, nil
}

// MyJobs returns the authenticated technician's own jobs.
func (s *Service) MyJobs(ctx context.Context, businessID, userID, filter string) ([]*Job, error) {
	tech, err := s.techSvc.GetByUserID(ctx, businessID, userID)
	if err != nil {
		return nil, err
	}
	jobs, err := s.repo.ListByTechnician(ctx, businessID, tech.ID, filter)
	if err != nil {
		return nil, err
	}
	if len(jobs) == 0 {
		return jobs, nil
	}
	ids := make([]string, 0, len(jobs))
	for _, j := range jobs {
		ids = append(ids, j.ID)
	}
	sums, err := s.repo.SumPaidByJobs(ctx, businessID, ids)
	if err != nil {
		return nil, err
	}
	for _, j := range jobs {
		paid := sums[j.ID]
		j.TotalPaid = paid
		j.PaymentText = payments.PaymentStatusText(paid, j.FinalAmount)
	}
	return jobs, nil
}

func (s *Service) ListByCustomer(ctx context.Context, businessID, customerID string) ([]*Job, error) {
	if err := s.repo.VerifyCustomer(ctx, businessID, customerID); err != nil {
		return nil, err
	}
	return s.repo.ListByCustomer(ctx, businessID, customerID)
}

func (s *Service) ListByTechnician(ctx context.Context, businessID, technicianID string) ([]*Job, error) {
	if _, err := s.techSvc.Get(ctx, businessID, technicianID); err != nil {
		return nil, err
	}
	return s.repo.ListByTechnician(ctx, businessID, technicianID, "")
}

func (s *Service) assertAssignedTechnician(ctx context.Context, businessID, userID, jobID string) (*Job, error) {
	tech, err := s.techSvc.GetByUserID(ctx, businessID, userID)
	if err != nil {
		return nil, err
	}
	job, err := s.repo.GetForUpdate(ctx, businessID, jobID)
	if err != nil {
		return nil, err
	}
	if job.TechnicianID == nil || *job.TechnicianID != tech.ID {
		return nil, httpapi.NewAPIError(403, "JOB_NOT_ASSIGNED", "This job is not assigned to you.")
	}
	return job, nil
}

// EnsureAssigned verifies the job is assigned to the given (technician) user.
// Provided to the payments module for read/write guards.
func (s *Service) EnsureAssigned(ctx context.Context, businessID, userID, jobID string) error {
	_, err := s.assertAssignedTechnician(ctx, businessID, userID, jobID)
	return err
}

func (s *Service) Accept(ctx context.Context, businessID, userID, jobID string) (*Job, error) {
	job, err := s.assertAssignedTechnician(ctx, businessID, userID, jobID)
	if err != nil {
		return nil, err
	}
	if err := ValidateTransition(job.Status, StatusAccepted); err != nil {
		return nil, err
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := transitionInTx(ctx, tx, job.ID, businessID, job.Status, StatusAccepted, userID); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	s.notifs.NotifyJobStatusChanged(ctx, s.jobEvent(ctx, businessID, job.ID, StatusAccepted))
	return s.Get(ctx, businessID, jobID)
}

func (s *Service) OnTheWay(ctx context.Context, businessID, userID, jobID string) (*Job, error) {
	job, err := s.assertAssignedTechnician(ctx, businessID, userID, jobID)
	if err != nil {
		return nil, err
	}
	if err := ValidateTransition(job.Status, StatusOnTheWay); err != nil {
		return nil, err
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := transitionInTx(ctx, tx, job.ID, businessID, job.Status, StatusOnTheWay, userID); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	s.notifs.NotifyJobStatusChanged(ctx, s.jobEvent(ctx, businessID, job.ID, StatusOnTheWay))
	return s.Get(ctx, businessID, jobID)
}

func (s *Service) Start(ctx context.Context, businessID, userID, jobID string) (*Job, error) {
	job, err := s.assertAssignedTechnician(ctx, businessID, userID, jobID)
	if err != nil {
		return nil, err
	}
	if err := ValidateTransition(job.Status, StatusStarted); err != nil {
		return nil, err
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := transitionInTx(ctx, tx, job.ID, businessID, job.Status, StatusStarted, userID); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	s.notifs.NotifyJobStatusChanged(ctx, s.jobEvent(ctx, businessID, job.ID, StatusStarted))
	return s.Get(ctx, businessID, jobID)
}

func (s *Service) Cancel(ctx context.Context, businessID, userID, jobID string) (*Job, error) {
	job, err := s.repo.GetForUpdate(ctx, businessID, jobID)
	if err != nil {
		return nil, err
	}
	if err := ValidateTransition(job.Status, StatusCancelled); err != nil {
		return nil, err
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := transitionInTx(ctx, tx, job.ID, businessID, job.Status, StatusCancelled, userID); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	s.notifs.NotifyJobStatusChanged(ctx, s.jobEvent(ctx, businessID, job.ID, StatusCancelled))
	return s.Get(ctx, businessID, jobID)
}

type CompleteRequest struct {
	Notes   *string          `json:"notes"`
	Items   []CompleteItem   `json:"items"`
	Payment *CompletePayment `json:"payment"`
}

type CompleteItem struct {
	Description string  `json:"description"`
	Quantity    float64 `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
}

type CompletePayment struct {
	Amount               float64 `json:"amount"`
	Method               string  `json:"method"`
	TransactionReference string  `json:"transaction_reference"`
}

func (s *Service) Complete(ctx context.Context, businessID, userID, jobID string, req CompleteRequest) (*Job, error) {
	job, err := s.assertAssignedTechnician(ctx, businessID, userID, jobID)
	if err != nil {
		return nil, err
	}
	if err := ValidateTransition(job.Status, StatusCompleted); err != nil {
		return nil, err
	}
	if len(req.Items) == 0 {
		return nil, httpapi.NewAPIError(400, "VALIDATION_ERROR", "At least one job item is required to complete a job.")
	}

	type itemRow struct {
		id    string
		desc  string
		qty   float64
		unit  float64
		total float64
	}
	var rows []itemRow
	final := 0.0
	for _, it := range req.Items {
		desc := strings.TrimSpace(it.Description)
		if desc == "" {
			return nil, httpapi.NewAPIError(400, "VALIDATION_ERROR", "Job item description is required.")
		}
		if it.UnitPrice < 0 || it.Quantity < 0 {
			return nil, httpapi.NewAPIError(400, "VALIDATION_ERROR", "Item quantity and unit price must not be negative.")
		}
		qty := it.Quantity
		if qty == 0 {
			qty = 1
		}
		total := qty * it.UnitPrice
		final += total
		rows = append(rows, itemRow{id: uuid.NewString(), desc: desc, qty: qty, unit: it.UnitPrice, total: total})
	}

	payMethod := ""
	payAmount := 0.0
	txnRef := ""
	if req.Payment != nil {
		pm := strings.ToUpper(strings.TrimSpace(req.Payment.Method))
		if !payments.ValidMethod(pm) {
			return nil, httpapi.NewAPIError(400, "VALIDATION_ERROR", "Invalid payment method.")
		}
		if req.Payment.Amount <= 0 || req.Payment.Amount > final {
			return nil, httpapi.NewAPIError(422, "PAYMENT_EXCEEDS_BALANCE", "Payment must be between 0 and the job total.")
		}
		payMethod = pm
		payAmount = req.Payment.Amount
		txnRef = strings.TrimSpace(req.Payment.TransactionReference)
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	for _, r := range rows {
		if _, err := tx.Exec(ctx, `
			INSERT INTO job_items (id, business_id, job_id, description, quantity, unit_price, total_price)
			VALUES ($1,$2,$3,$4,$5,$6,$7)`,
			r.id, businessID, jobID, r.desc, r.qty, r.unit, r.total); err != nil {
			return nil, err
		}
	}

	if err := setFinalAmountInTx(ctx, tx, businessID, jobID, final, req.Notes); err != nil {
		return nil, err
	}

	if payMethod != "" {
		var ref any
		if txnRef != "" {
			ref = &txnRef
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO payments (id, business_id, job_id, amount, method, status, transaction_reference, created_by)
			VALUES ($1,$2,$3,$4,$5,'PAID',$6,$7)`,
			uuid.NewString(), businessID, jobID, payAmount, payMethod, ref, userID); err != nil {
			return nil, err
		}
	}

	var issued *IssuedReceipt
	if s.issuer != nil {
		rec, err := s.issuer.IssueTx(ctx, tx, businessID, jobID)
		if err != nil {
			return nil, err
		}
		issued = rec
	}

	if err := transitionInTx(ctx, tx, job.ID, businessID, job.Status, StatusCompleted, userID); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	s.notifs.NotifyJobCompleted(ctx, s.jobEvent(ctx, businessID, job.ID, StatusCompleted))

	job, err = s.Get(ctx, businessID, jobID)
	if err != nil {
		return nil, err
	}
	if issued != nil {
		job.Receipt = &ReceiptInfo{ID: issued.ID, ReceiptNumber: issued.ReceiptNumber, IssuedAt: issued.IssuedAt}
	}
	return job, nil
}

func parseOptionalTime(s *string) (*time.Time, error) {
	if s == nil || strings.TrimSpace(*s) == "" {
		return nil, nil
	}
	t, err := time.Parse(time.RFC3339, strings.TrimSpace(*s))
	if err != nil {
		return nil, err
	}
	return &t, nil
}