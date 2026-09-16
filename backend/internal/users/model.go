package users

import "time"

const (
	RoleOwner      = "OWNER"
	RoleAdmin      = "ADMIN"
	RoleTechnician = "TECHNICIAN"
)

type User struct {
	ID           string     `json:"id"`
	BusinessID   string     `json:"business_id"`
	Name         string     `json:"name"`
	Phone        string     `json:"phone"`
	Email        *string    `json:"email"`
	PasswordHash string     `json:"-"`
	Role         string     `json:"role"`
	IsActive     bool       `json:"is_active"`
	LastLoginAt  *time.Time `json:"last_login_at"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}