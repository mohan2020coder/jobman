package receipts

import "time"

type Receipt struct {
	ID            string    `json:"id"`
	BusinessID    string    `json:"business_id"`
	JobID         string    `json:"job_id"`
	ReceiptNumber string    `json:"receipt_number"`
	PublicToken   string    `json:"public_token"`
	IssuedAt      time.Time `json:"issued_at"`
	CreatedAt     time.Time `json:"created_at"`
}