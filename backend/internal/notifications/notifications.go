package notifications

import (
	"context"
	"time"
)

// JobEvent carries the context for a job notification.
type JobEvent struct {
	BusinessID     string     `json:"-"`
	JobID          string     `json:"id"`
	JobNumber      string     `json:"job_number"`
	TechnicianID   string     `json:"-"`
	Status         string     `json:"status"`
	ServiceType    string     `json:"service_type"`
	CustomerName   string     `json:"customer_name"`
	EstimatedAmount float64   `json:"estimated_amount"`
	ScheduledAt    *time.Time `json:"scheduled_at,omitempty"`
	Address        *string    `json:"address,omitempty"`
}

// Service is the notification abstraction. The job service depends on this
// interface, not on any concrete push provider. Implementations can be
// wired later (FCM, WhatsApp, SMS) without changing job logic.
type Service interface {
	NotifyJobAssigned(ctx context.Context, e JobEvent)
	NotifyJobStatusChanged(ctx context.Context, e JobEvent)
	NotifyJobCompleted(ctx context.Context, e JobEvent)
}

// Noop ignores all notifications. Used as the default implementation.
type Noop struct{}

func (Noop) NotifyJobAssigned(_ context.Context, _ JobEvent)     {}
func (Noop) NotifyJobStatusChanged(_ context.Context, _ JobEvent) {}
func (Noop) NotifyJobCompleted(_ context.Context, _ JobEvent)     {}

var _ Service = Noop{}