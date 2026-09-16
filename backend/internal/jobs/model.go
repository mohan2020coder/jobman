package jobs

import (
	"time"

	"github.com/jobman/backend/internal/customers"
	"github.com/jobman/backend/internal/payments"
)

const (
	StatusPending   = "PENDING"
	StatusAccepted  = "ACCEPTED"
	StatusOnTheWay  = "ON_THE_WAY"
	StatusStarted   = "STARTED"
	StatusCompleted = "COMPLETED"
	StatusCancelled = "CANCELLED"
)

type Job struct {
	ID                 string     `json:"id"`
	BusinessID         string     `json:"business_id"`
	CustomerID         string     `json:"customer_id"`
	TechnicianID       *string    `json:"technician_id"`
	JobNumber          string     `json:"job_number"`
	ServiceType        string     `json:"service_type"`
	ProblemDescription *string    `json:"problem_description"`
	Address            *string    `json:"address"`
	Latitude           *float64   `json:"latitude"`
	Longitude          *float64   `json:"longitude"`
	Status             string     `json:"status"`
	EstimatedAmount    float64    `json:"estimated_amount"`
	FinalAmount        float64    `json:"final_amount"`
	ScheduledAt        *time.Time `json:"scheduled_at"`
	AcceptedAt         *time.Time `json:"accepted_at"`
	OnTheWayAt         *time.Time `json:"on_the_way_at"`
	StartedAt          *time.Time `json:"started_at"`
	CompletedAt        *time.Time `json:"completed_at"`
	Notes              *string    `json:"notes"`
	CreatedBy          string     `json:"created_by"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`

	// Derived presentation fields
	Customer    *customers.Customer     `json:"customer,omitempty"`
	Technician  *TechnicianLite         `json:"technician,omitempty"`
	Items       []*Item                 `json:"items,omitempty"`
	Payments    []*payments.Payment     `json:"payments,omitempty"`
	StatusLog   []*StatusHistory        `json:"status_history,omitempty"`
	TotalPaid   float64                 `json:"total_paid"`
	PaymentText string                  `json:"payment_status"`
	Receipt     *ReceiptInfo            `json:"receipt,omitempty"`
}

type TechnicianLite struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Phone   string `json:"phone"`
	Status  string `json:"status"`
}

type ReceiptInfo struct {
	ID            string    `json:"id"`
	ReceiptNumber string    `json:"receipt_number"`
	IssuedAt      time.Time `json:"issued_at"`
}

type Item struct {
	ID          string  `json:"id"`
	BusinessID  string  `json:"business_id"`
	JobID       string  `json:"job_id"`
	Description string  `json:"description"`
	Quantity    float64 `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
	TotalPrice  float64 `json:"total_price"`
	CreatedAt   time.Time `json:"created_at"`
}

type StatusHistory struct {
	ID        string     `json:"id"`
	OldStatus *string    `json:"old_status"`
	NewStatus string     `json:"new_status"`
	ChangedBy string     `json:"changed_by"`
	CreatedAt time.Time  `json:"created_at"`
}