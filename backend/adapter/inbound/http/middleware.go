package http

import (
	"net/http"
	"strings"

	"splitwise/domain/port/inbound"
)

// AuthMiddleware creates middleware to validate JWT tokens
func AuthMiddleware(authService inbound.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Authorization header required", http.StatusUnauthorized)
				return
			}

			// Extract token from "Bearer <token>"
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
				return
			}

			token := parts[1]
			userID, err := authService.ValidateToken(token)
			if err != nil {
				http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
				return
			}

			// Add user ID to request context for use in handlers
			ctx := r.Context()
			r = r.WithContext(ctx)
			r.Header.Set("X-User-ID", userID)

			next.ServeHTTP(w, r)
		})
	}
}
