package receipts

import (
	"context"
	"time"

	"github.com/jobman/backend/database"
	"github.com/jobman/backend/pkg/httpapi"
)

// Service exposes receipt issuing and the safe public receipt payload.
type Service struct {
	repo     *Repository
	jobData  JobDataProvider
	bizInfo  BusinessInfoFn
}

// JobSnapshot is the job data needed to render a receipt. It is provided by
// an implementation injected in main.go to keep this package dependency-free.
type JobSnapshot struct {
	CustomerName       string
	ServiceType        string
	ProblemDescription *string
	FinalAmount        float64
	CompletedAt        *time.Time
	Items              []JobSnapshotItem
	Payments           []PaymentLine
}

type JobSnapshotItem struct {
	Description string
	Quantity    float64
	TotalPrice  float64
}

type PaymentLine struct {
	Amount float64
	Method string
	Status string
}

type JobDataProvider interface {
	PublicJobData(ctx context.Context, businessID, jobID string) (*JobSnapshot, error)
}

// BusinessInfoFn resolves a business name/phone from its id.
type BusinessInfoFn func(ctx context.Context, id string) (name, phone string, err error)

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) SetJobDataProvider(p JobDataProvider) { s.jobData = p }
func (s *Service) SetBusinessInfo(fn BusinessInfoFn)    { s.bizInfo = fn }

// Issuer creates receipts inside the caller's transaction. Jobs depends on
// this implementation via a small interface.
type Issuer struct{ repo *Repository }

func NewIssuer(repo *Repository) *Issuer { return &Issuer{repo: repo} }

func (i *Issuer) IssueTx(ctx context.Context, tx database.Querier, businessID, jobID string) (*Receipt, error) {
	return i.repo.CreateTx(ctx, tx, businessID, jobID)
}

// GetByJob returns the receipt (and token) for an authenticated job view.
func (s *Service) GetByJob(ctx context.Context, businessID, jobID string) (*Receipt, error) {
	return s.repo.GetByJob(ctx, businessID, jobID)
}

// PublicReceipt builds the safe public receipt payload.
type PublicReceipt struct {
	Business struct {
		Name  string `json:"name"`
		Phone string `json:"phone"`
	} `json:"business"`
	ReceiptNumber string       `json:"receipt_number"`
	CustomerName  string       `json:"customer_name"`
	Service       string       `json:"service"`
	Problem       string       `json:"problem_description,omitempty"`
	Items         []PublicItem `json:"items"`
	Total         float64      `json:"total"`
	Paid          float64      `json:"paid"`
	PaymentStatus string       `json:"payment_status"`
	PaymentMethods []string    `json:"payment_methods"`
	CompletedAt   *time.Time   `json:"completed_at"`
	IssuedAt      time.Time    `json:"issued_at"`
}

type PublicItem struct {
	Description string  `json:"description"`
	Quantity    float64 `json:"quantity"`
	Amount      float64 `json:"amount"`
}

func (s *Service) PublicReceipt(ctx context.Context, token string) (*PublicReceipt, error) {
	rec, err := s.repo.GetByToken(ctx, token)
	if err != nil {
		return nil, err
	}
	if s.jobData == nil {
		return nil, httpapi.NewAPIError(500, "INTERNAL_ERROR", "Receipt data unavailable.")
	}

	snap, err := s.jobData.PublicJobData(ctx, rec.BusinessID, rec.JobID)
	if err != nil {
		return nil, err
	}

	paid := 0.0
	methods := map[string]bool{}
	for _, p := range snap.Payments {
		if p.Status != "PAID" {
			continue
		}
		paid += p.Amount
		methods[p.Method] = true
	}

	out := &PublicReceipt{
		ReceiptNumber: rec.ReceiptNumber,
		CustomerName:  snap.CustomerName,
		Service:       snap.ServiceType,
		Problem:       strOrEmpty(snap.ProblemDescription),
		Total:         snap.FinalAmount,
		Paid:          paid,
		PaymentStatus: "UNPAID",
		CompletedAt:   snap.CompletedAt,
		IssuedAt:      rec.IssuedAt,
	}
	if out.Total > 0 && paid >= out.Total {
		out.PaymentStatus = "PAID"
	} else if paid > 0 {
		out.PaymentStatus = "PARTIAL"
	}
	for m := range methods {
		out.PaymentMethods = append(out.PaymentMethods, m)
	}
	for _, it := range snap.Items {
		out.Items = append(out.Items, PublicItem{
			Description: it.Description,
			Quantity:    it.Quantity,
			Amount:      it.TotalPrice,
		})
	}

	if s.bizInfo != nil {
		if name, phone, err := s.bizInfo(ctx, rec.BusinessID); err == nil {
			out.Business.Name = name
			out.Business.Phone = phone
		}
	}

	return out, nil
}

func strOrEmpty(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}