package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/exotic-travel-booking/backend/internal/services"
)

// AuthMiddleware creates authentication middleware
func AuthMiddleware(authService *services.AuthService) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Authorization header required", http.StatusUnauthorized)
				return
			}

			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			if tokenString == authHeader {
				http.Error(w, "Bearer token required", http.StatusUnauthorized)
				return
			}

			// Validate token
			claims, err := authService.ValidateToken(tokenString)
			if err != nil {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			// Get user details
			user, err := authService.GetUserByID(r.Context(), claims.UserID)
			if err != nil {
				http.Error(w, "User not found", http.StatusUnauthorized)
				return
			}

			// Add user to context
			ctx := context.WithValue(r.Context(), "user", user)
			ctx = context.WithValue(ctx, "userID", user.ID)
			ctx = context.WithValue(ctx, "userRole", user.Role)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// OptionalAuthMiddleware creates optional authentication middleware
// This middleware will add user to context if token is present, but won't fail if it's not
func OptionalAuthMiddleware(authService *services.AuthService) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader != "" {
				tokenString := strings.TrimPrefix(authHeader, "Bearer ")
				if tokenString != authHeader {
					// Validate token
					claims, err := authService.ValidateToken(tokenString)
					if err == nil {
						// Get user details
						user, err := authService.GetUserByID(r.Context(), claims.UserID)
						if err == nil {
							// Add user to context
							ctx := context.WithValue(r.Context(), "user", user)
							ctx = context.WithValue(ctx, "userID", user.ID)
							ctx = context.WithValue(ctx, "userRole", user.Role)
							r = r.WithContext(ctx)
						}
					}
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

// AdminMiddleware ensures the user has admin role
func AdminMiddleware() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userRole, ok := r.Context().Value("userRole").(string)
			if !ok || userRole != "admin" {
				http.Error(w, "Admin access required", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
