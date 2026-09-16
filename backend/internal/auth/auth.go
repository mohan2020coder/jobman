package auth

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jobman/backend/internal/identity"
	"github.com/jobman/backend/pkg/httpapi"
)

// Aliases so existing callers keep using auth package helpers.
type Principal = identity.Principal

func WithPrincipal(ctx context.Context, p *Principal) context.Context {
	return identity.WithPrincipal(ctx, p)
}

func PrincipalFrom(ctx context.Context) (*Principal, bool) {
	return identity.PrincipalFrom(ctx)
}

func MustPrincipal(ctx context.Context) *Principal {
	return identity.MustPrincipal(ctx)
}

// Claims is the JWT payload for an authenticated user.
type Claims struct {
	jwt.RegisteredClaims
	UserID     string `json:"uid"`
	BusinessID string `json:"bid"`
	Role       string `json:"role"`
}

var ErrInvalidToken = errors.New("invalid or expired token")

// Sign issues a new JWT for the given identity.
func Sign(secret []byte, expiration time.Duration, userID, businessID, role string) (string, error) {
	now := time.Now()
	claims := &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(expiration)),
		},
		UserID:     userID,
		BusinessID: businessID,
		Role:       role,
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(secret)
	if err != nil {
		return "", httpapi.NewAPIError(500, "TOKEN_SIGNING_FAILED", "Failed to issue token.")
	}
	return token, nil
}

// Parse verifies a JWT and returns its claims.
func Parse(secret []byte, tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return secret, nil
	})
	if err != nil || !token.Valid {
		return nil, ErrInvalidToken
	}
	return claims, nil
}