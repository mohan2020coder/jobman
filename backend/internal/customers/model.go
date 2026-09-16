package customers

import (
	"time"
)

type Customer struct {
	ID         string     `json:"id"`
	BusinessID string     `json:"business_id"`
	Name       string     `json:"name"`
	Phone      string     `json:"phone"`
	Email      *string    `json:"email"`
	Address    *string    `json:"address"`
	Latitude   *float64   `json:"latitude"`
	Longitude  *float64   `json:"longitude"`
	Notes      *string    `json:"notes"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}