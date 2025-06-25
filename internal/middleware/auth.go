package middleware

import (
	"context"
	"net/http"
	"strings"

	"todo/internal/models"
	"todo/internal/services"
)

type contextKey string

const userContextKey contextKey = "userInfo"

func AuthMiddleware(authService *services.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Authorization header is required", http.StatusUnauthorized)
				return
			}

			// Check for Bearer prefix
			if !strings.HasPrefix(authHeader, "Bearer ") {
				http.Error(w, "Invalid authorization header format. Use 'Bearer <token>'", http.StatusUnauthorized)
				return
			}

			token := strings.TrimPrefix(authHeader, "Bearer ")
			if strings.TrimSpace(token) == "" {
				http.Error(w, "Token is required", http.StatusUnauthorized)
				return
			}

			user, err := authService.ValidateToken(token)
			if err != nil {
				http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
				return
			}

			// Attach user info to context
			ctx := context.WithValue(r.Context(), userContextKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// Helper to get user info from context
func GetUserFromContext(ctx context.Context) (*models.UserInfo, bool) {
	user, ok := ctx.Value(userContextKey).(*models.UserInfo)
	return user, ok
}
