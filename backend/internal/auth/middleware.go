package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/exotic-travel-booking/backend/internal/models"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// ContextKey represents keys used in request context
type ContextKey string

const (
	// UserContextKey is the key for user in context
	UserContextKey ContextKey = "user"
	// ClaimsContextKey is the key for JWT claims in context
	ClaimsContextKey ContextKey = "claims"
	// SessionContextKey is the key for session ID in context
	SessionContextKey ContextKey = "session_id"
)

// AuthMiddleware provides authentication middleware
type AuthMiddleware struct {
	jwtManager    *JWTManager
	blacklist     TokenBlacklist
	tracer        trace.Tracer
	skipPaths     map[string]bool
	userService   UserService
}

// UserService interface for user operations
type UserService interface {
	GetUserByID(ctx context.Context, id int) (*models.User, error)
	UpdateLastLogin(ctx context.Context, userID int) error
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(jwtManager *JWTManager, blacklist TokenBlacklist, userService UserService) *AuthMiddleware {
	skipPaths := map[string]bool{
		"/api/v1/auth/login":    true,
		"/api/v1/auth/register": true,
		"/api/v1/auth/refresh":  true,
		"/api/v1/health":        true,
		"/metrics":              true,
	}

	return &AuthMiddleware{
		jwtManager:  jwtManager,
		blacklist:   blacklist,
		tracer:      otel.Tracer("auth.middleware"),
		skipPaths:   skipPaths,
		userService: userService,
	}
}

// RequireAuth middleware that requires authentication
func (am *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, span := am.tracer.Start(r.Context(), "auth_middleware.require_auth")
		defer span.End()

		// Skip authentication for certain paths
		if am.skipPaths[r.URL.Path] {
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		// Extract token from Authorization header
		token, err := am.extractToken(r)
		if err != nil {
			span.RecordError(err)
			am.writeUnauthorized(w, "Missing or invalid authorization header")
			return
		}

		// Validate token
		claims, err := am.jwtManager.ValidateToken(ctx, token)
		if err != nil {
			span.RecordError(err)
			am.writeUnauthorized(w, "Invalid token")
			return
		}

		// Check if token is blacklisted
		if am.blacklist != nil {
			blacklisted, err := am.blacklist.IsBlacklisted(ctx, claims.ID)
			if err != nil {
				span.RecordError(err)
				am.writeError(w, "Authentication service error", http.StatusInternalServerError)
				return
			}
			if blacklisted {
				am.writeUnauthorized(w, "Token has been revoked")
				return
			}
		}

		// Ensure it's an access token
		if claims.TokenType != "access" {
			am.writeUnauthorized(w, "Invalid token type")
			return
		}

		// Get user details
		user, err := am.userService.GetUserByID(ctx, claims.UserID)
		if err != nil {
			span.RecordError(err)
			am.writeUnauthorized(w, "User not found")
			return
		}

		// Check if user is active
		if user.Status != models.UserStatusActive {
			am.writeUnauthorized(w, "User account is not active")
			return
		}

		// Add user and claims to context
		ctx = context.WithValue(ctx, UserContextKey, user)
		ctx = context.WithValue(ctx, ClaimsContextKey, claims)
		ctx = context.WithValue(ctx, SessionContextKey, claims.SessionID)

		span.SetAttributes(
			attribute.Int("user.id", user.ID),
			attribute.String("user.email", user.Email),
			attribute.String("user.role", string(user.Role)),
			attribute.String("session.id", claims.SessionID),
		)

		// Update last login time (async)
		go func() {
			am.userService.UpdateLastLogin(context.Background(), user.ID)
		}()

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireRole middleware that requires specific role
func (am *AuthMiddleware) RequireRole(roles ...models.UserRole) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, span := am.tracer.Start(r.Context(), "auth_middleware.require_role")
			defer span.End()

			user := am.GetUserFromContext(r.Context())
			if user == nil {
				am.writeUnauthorized(w, "Authentication required")
				return
			}

			// Check if user has required role
			hasRole := false
			for _, role := range roles {
				if user.Role == role {
					hasRole = true
					break
				}
			}

			if !hasRole {
				span.SetAttributes(
					attribute.String("required.roles", strings.Join(am.rolesToStrings(roles), ",")),
					attribute.String("user.role", string(user.Role)),
				)
				am.writeForbidden(w, "Insufficient permissions")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequirePermission middleware that requires specific permission
func (am *AuthMiddleware) RequirePermission(permission string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, span := am.tracer.Start(r.Context(), "auth_middleware.require_permission")
			defer span.End()

			claims := am.GetClaimsFromContext(r.Context())
			if claims == nil {
				am.writeUnauthorized(w, "Authentication required")
				return
			}

			// Check if user has required permission
			hasPermission := false
			for _, perm := range claims.Permissions {
				if perm == permission {
					hasPermission = true
					break
				}
			}

			if !hasPermission {
				span.SetAttributes(
					attribute.String("required.permission", permission),
					attribute.StringSlice("user.permissions", claims.Permissions),
				)
				am.writeForbidden(w, "Insufficient permissions")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireCompanyAccess middleware that ensures user belongs to the company
func (am *AuthMiddleware) RequireCompanyAccess(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, span := am.tracer.Start(r.Context(), "auth_middleware.require_company_access")
		defer span.End()

		user := am.GetUserFromContext(r.Context())
		if user == nil {
			am.writeUnauthorized(w, "Authentication required")
			return
		}

		// Extract company ID from URL or request
		// This is a simplified example - in practice, you'd extract from URL path
		companyID := am.extractCompanyIDFromRequest(r)
		if companyID == 0 {
			next.ServeHTTP(w, r)
			return
		}

		// Check if user belongs to the company
		if user.CompanyID != companyID {
			span.SetAttributes(
				attribute.Int("required.company_id", companyID),
				attribute.Int("user.company_id", user.CompanyID),
			)
			am.writeForbidden(w, "Access denied to company resources")
			return
		}

		next.ServeHTTP(w, r)
	})
}

// RateLimitMiddleware provides rate limiting
func (am *AuthMiddleware) RateLimitMiddleware(requestsPerMinute int) func(http.Handler) http.Handler {
	// This would typically use Redis or in-memory store for rate limiting
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, span := am.tracer.Start(r.Context(), "auth_middleware.rate_limit")
			defer span.End()

			// Get user ID for rate limiting
			userID := "anonymous"
			if user := am.GetUserFromContext(r.Context()); user != nil {
				userID = string(rune(user.ID))
			}

			// Check rate limit (implementation would use Redis)
			// For now, we'll just pass through
			span.SetAttributes(
				attribute.String("rate_limit.user_id", userID),
				attribute.Int("rate_limit.requests_per_minute", requestsPerMinute),
			)

			next.ServeHTTP(w, r)
		})
	}
}

// CORS middleware for handling cross-origin requests
func (am *AuthMiddleware) CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*") // Configure appropriately for production
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		w.Header().Set("Access-Control-Max-Age", "86400")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// SecurityHeaders middleware adds security headers
func (am *AuthMiddleware) SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Security headers
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		w.Header().Set("Content-Security-Policy", "default-src 'self'")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

		next.ServeHTTP(w, r)
	})
}

// Helper methods

// GetUserFromContext extracts user from request context
func (am *AuthMiddleware) GetUserFromContext(ctx context.Context) *models.User {
	if user, ok := ctx.Value(UserContextKey).(*models.User); ok {
		return user
	}
	return nil
}

// GetClaimsFromContext extracts JWT claims from request context
func (am *AuthMiddleware) GetClaimsFromContext(ctx context.Context) *Claims {
	if claims, ok := ctx.Value(ClaimsContextKey).(*Claims); ok {
		return claims
	}
	return nil
}

// GetSessionIDFromContext extracts session ID from request context
func (am *AuthMiddleware) GetSessionIDFromContext(ctx context.Context) string {
	if sessionID, ok := ctx.Value(SessionContextKey).(string); ok {
		return sessionID
	}
	return ""
}

func (am *AuthMiddleware) extractToken(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf("missing authorization header")
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return "", fmt.Errorf("invalid authorization header format")
	}

	return parts[1], nil
}

func (am *AuthMiddleware) extractCompanyIDFromRequest(r *http.Request) int {
	// This would extract company ID from URL path or headers
	// For example: /api/v1/companies/123/campaigns
	// Implementation depends on your URL structure
	return 0
}

func (am *AuthMiddleware) rolesToStrings(roles []models.UserRole) []string {
	result := make([]string, len(roles))
	for i, role := range roles {
		result[i] = string(role)
	}
	return result
}

func (am *AuthMiddleware) writeUnauthorized(w http.ResponseWriter, message string) {
	am.writeError(w, message, http.StatusUnauthorized)
}

func (am *AuthMiddleware) writeForbidden(w http.ResponseWriter, message string) {
	am.writeError(w, message, http.StatusForbidden)
}

func (am *AuthMiddleware) writeError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := map[string]interface{}{
		"success": false,
		"error":   message,
		"code":    statusCode,
		"timestamp": time.Now().Unix(),
	}

	json.NewEncoder(w).Encode(response)
}
