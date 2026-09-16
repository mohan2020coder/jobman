package technicians

import (
	"github.com/jobman/backend/internal/auth"
	"github.com/jobman/backend/pkg/httpapi"
)

// passwordHash delegates to the auth package's bcrypt helper.
var usersPasswordHash = auth.HashPassword

// passwordHashError maps hashing failures to an API error.
func hashPassword(plain string) (string, error) {
	h, err := usersPasswordHash(plain)
	if err != nil {
		return "", httpapi.NewAPIError(500, "PASSWORD_HASH_FAILED", "Failed to secure password.")
	}
	return h, nil
}