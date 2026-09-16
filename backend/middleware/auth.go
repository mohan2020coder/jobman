package middleware

import (
	"net/http"
	"strings"

	"github.com/jobman/backend/internal/auth"
	"github.com/jobman/backend/pkg/httpapi"
)

// RequireAuth validates the Bearer token and injects the principal into
// the request context.
func RequireAuth(secret []byte) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if header == "" || !strings.HasPrefix(header, "Bearer ") {
				httpapi.WriteError(w, http.StatusUnauthorized, "AUTH_UNAUTHORIZED", "Missing or invalid authorization header.")
				return
			}
			tokenString := strings.TrimPrefix(header, "Bearer ")
			claims, err := auth.Parse(secret, tokenString)
			if err != nil {
				httpapi.WriteError(w, http.StatusUnauthorized, "AUTH_INVALID_TOKEN", "Session expired, please log in again.")
				return
			}
			if claims.BusinessID == "" || claims.UserID == "" || claims.Role == "" {
				httpapi.WriteError(w, http.StatusUnauthorized, "AUTH_INVALID_TOKEN", "Token is missing required claims.")
				return
			}

			principal := &auth.Principal{
				UserID:     claims.UserID,
				BusinessID: claims.BusinessID,
				Role:       claims.Role,
			}
			next.ServeHTTP(w, r.WithContext(auth.WithPrincipal(r.Context(), principal)))
		})
	}
}

// RequireRole restricts access to one of the given roles.
func RequireRole(roles ...string) func(http.Handler) http.Handler {
	allowed := make(map[string]bool, len(roles))
	for _, r := range roles {
		allowed[r] = true
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			p, ok := auth.PrincipalFrom(r.Context())
			if !ok {
				httpapi.WriteError(w, http.StatusUnauthorized, "AUTH_UNAUTHORIZED", "Authentication required.")
				return
			}
			if !allowed[p.Role] {
				httpapi.WriteError(w, http.StatusForbidden, "FORBIDDEN", "You do not have permission to perform this action.")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}