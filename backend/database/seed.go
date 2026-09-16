package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

// SeedPasswords replaces placeholder hashes introduced by the seed migration
// with a real bcrypt hash so demo accounts can log in.
func SeedPasswords(ctx context.Context, pool *pgxpool.Pool, plain string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("seed password hash: %w", err)
	}
	tag, err := pool.Exec(ctx,
		`UPDATE users SET password_hash = $1, updated_at = NOW() WHERE password_hash = 'PLACEHOLDER'`, string(hash))
	if err != nil {
		return err
	}
	if tag.RowsAffected() > 0 {
		fmt.Printf("database: set demo password for %d seed users\n", tag.RowsAffected())
	}
	return nil
}