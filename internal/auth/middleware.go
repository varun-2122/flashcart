package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/varun-2122/flashcart/internal/domain"
	"github.com/varun-2122/flashcart/internal/response"
)

type contextKey string

const (
	UserIDKey    contextKey = "auth_user_id"
	UserRoleKey  contextKey = "auth_user_role"
	UserEmailKey contextKey = "auth_user_email"
)

// AuthMiddleware verifies JWT Bearer token and attaches claims to context.
func AuthMiddleware(jwtManager *JWTManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				response.ErrorJSON(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing Authorization header", nil)
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				response.ErrorJSON(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid Authorization header format", nil)
				return
			}

			claims, err := jwtManager.VerifyToken(parts[1])
			if err != nil {
				response.ErrorJSON(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid or expired JWT token", nil)
				return
			}

			ctx := r.Context()
			ctx = context.WithValue(ctx, UserIDKey, claims.UserID)
			ctx = context.WithValue(ctx, UserRoleKey, claims.Role)
			ctx = context.WithValue(ctx, UserEmailKey, claims.Email)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireRole enforces Role-Based Access Control (RBAC) on routes.
func RequireRole(allowedRoles ...domain.Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userRole, ok := r.Context().Value(UserRoleKey).(domain.Role)
			if !ok {
				response.ErrorJSON(w, http.StatusForbidden, "FORBIDDEN", "User role not found in context", nil)
				return
			}

			allowed := false
			for _, role := range allowedRoles {
				if userRole == role {
					allowed = true
					break
				}
			}

			if !allowed {
				response.ErrorJSON(w, http.StatusForbidden, "FORBIDDEN", "Insufficient permissions for operation", nil)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// GetUserIDFromContext retrieves authenticated User ID from request context.
func GetUserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(UserIDKey).(uuid.UUID)
	return id, ok
}

// GetUserRoleFromContext retrieves authenticated User Role from request context.
func GetUserRoleFromContext(ctx context.Context) (domain.Role, bool) {
	role, ok := ctx.Value(UserRoleKey).(domain.Role)
	return role, ok
}
