package payments

import "time"

const (
	MethodCash        = "CASH"
	MethodUPI         = "UPI"
	MethodCard        = "CARD"
	MethodBankTransfer = "BANK_TRANSFER"
	MethodOther       = "OTHER"

	StatusPaid     = "PAID"
	StatusRefunded = "REFUNDED"
)

type Payment struct {
	ID                   string     `json:"id"`
	BusinessID           string     `json:"business_id"`
	JobID                string     `json:"job_id"`
	Amount               float64    `json:"amount"`
	Method               string     `json:"method"`
	Status               string     `json:"status"`
	TransactionReference *string    `json:"transaction_reference"`
	PaidAt               time.Time  `json:"paid_at"`
	CreatedBy            string     `json:"created_by"`
	CreatedAt            time.Time  `json:"created_at"`
}

// PaymentStatusText derives the display status from paid vs final amount.
func PaymentStatusText(paid, final float64) string {
	_ = StatusRefunded
	switch {
	case paid <= 0:
		return "UNPAID"
	case paid < final:
		return "PARTIAL"
	default:
		return "PAID"
	}
}