package technicians

import (
	"time"

	"github.com/jobman/backend/internal/users"
)

const (
	StatusAvailable = "AVAILABLE"
	StatusBusy      = "BUSY"
	StatusInactive  = "INACTIVE"
)

// Technician is a user plus technician-specific information.
type Technician struct {
	ID         string     `json:"id"`
	BusinessID string     `json:"business_id"`
	UserID     string     `json:"user_id"`
	Name       string     `json:"name"`
	Phone      string     `json:"phone"`
	Email      *string    `json:"email"`
	Status     string     `json:"status"`
	IsActive   bool       `json:"is_active"`
	LastLogin  *time.Time `json:"last_login_at"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

// WithUser decorates a tech with user fields. Kept close to the model for clarity.
func (t *Technician) WithUser(u *users.User) *Technician {
	t.UserID = u.ID
	t.Name = u.Name
	t.Phone = u.Phone
	t.Email = u.Email
	t.IsActive = u.IsActive
	t.LastLogin = u.LastLoginAt
	return t
}