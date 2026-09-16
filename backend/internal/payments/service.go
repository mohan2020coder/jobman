package payments

import (
	"context"
	"strings"

	"github.com/jobman/backend/pkg/httpapi"
)

type Service struct {
	repo *Repository
	// assignedCheck, when set, verifies a self-serve technician owns the job.
	assignedCheck func(ctx context.Context, businessID, userID, jobID string) error
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// CheckAssigned keeps technician self-service safe (owner/admin paths in
// main.go simply skip the check).
type AssignmentChecker func(ctx context.Context, businessID, userID, jobID string) error

func (s *Service) SetAssignmentChecker(c AssignmentChecker) { s.assignedCheck = c }

func (s *Service) selfServeAllowed(ctx context.Context, businessID, userID, jobID string) error {
	if s.assignedCheck != nil {
		return s.assignedCheck(ctx, businessID, userID, jobID)
	}
	return nil
}

type CreateRequest struct {
	Amount               float64 `json:"amount"`
	Method               string  `json:"method"`
	TransactionReference string  `json:"transaction_reference"`
}

func (s *Service) Create(ctx context.Context, businessID, jobID, createdBy string, req CreateRequest) (*Payment, error) {
	if req.Amount <= 0 {
		return nil, httpapi.NewAPIError(400, "VALIDATION_ERROR", "Payment amount must be greater than zero.")
	}
	method := strings.ToUpper(strings.TrimSpace(req.Method))
	if !ValidMethod(method) {
		return nil, httpapi.NewAPIError(400, "VALIDATION_ERROR", "Invalid payment method.")
	}

	final, estimated, err := s.repo.GetJobAmounts(ctx, businessID, jobID)
	if err != nil {
		return nil, err
	}

	basis := estimated
	if final > 0 {
		basis = final
	}

	paid, err := s.repo.SumPaidByJob(ctx, businessID, jobID)
	if err != nil {
		return nil, err
	}

	remaining := basis - paid
	if req.Amount > remaining {
		return nil, httpapi.NewAPIError(422, "PAYMENT_EXCEEDS_BALANCE",
			"Payment exceeds the remaining balance of "+formatRupee(remaining)+". Amount already paid: "+formatRupee(paid)+".")
	}

	var txnRef *string
	if tr := strings.TrimSpace(req.TransactionReference); tr != "" {
		txnRef = &tr
	}

	return s.repo.Create(ctx, &Payment{
		BusinessID:           businessID,
		JobID:                jobID,
		Amount:               req.Amount,
		Method:               method,
		Status:               StatusPaid,
		TransactionReference: txnRef,
		CreatedBy:            createdBy,
	})
}

func (s *Service) ListByJob(ctx context.Context, businessID, jobID string) ([]*Payment, error) {
	if err := s.repo.VerifyJobAccess(ctx, businessID, jobID); err != nil {
		return nil, err
	}
	return s.repo.ListByJob(ctx, businessID, jobID)
}

func (s *Service) SumPaidByJob(ctx context.Context, businessID, jobID string) (float64, error) {
	return s.repo.SumPaidByJob(ctx, businessID, jobID)
}

func formatRupee(v float64) string {
	// simple two-decimal formatting without currency symbol
	whole := int64(v)
	frac := int64((v-float64(whole))*100 + 0.5)
	return itoa(whole) + "." + two(frac)
}

func itoa(n int64) string {
	if n == 0 {
		return "0"
	}
	var b []byte
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	return string(b)
}

func two(n int64) string {
	if n > 99 {
		n = 99
	}
	if n < 10 {
		return "0" + itoa(n)
	}
	return itoa(n)
}