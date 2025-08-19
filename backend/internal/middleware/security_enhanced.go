package middleware

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/exotic-travel-booking/backend/internal/security"
)

// SecurityConfig holds security middleware configuration
type SecurityConfig struct {
	JWTManager     *security.JWTManager
	Validator      *security.Validator
	AuditLogger    *security.AuditLogger
	MaxRequestSize int64
	RateLimitRPS   float64
	RateLimitBurst int
	EnableCSRF     bool
	EnableCORS     bool
	TrustedProxies []string
	BlockedIPs     []string
	AllowedOrigins []string
	SessionTimeout time.Duration
}

// SecurityMiddleware provides comprehensive security features
type SecurityMiddleware struct {
	config      SecurityConfig
	trustedNets []*net.IPNet
	blockedNets []*net.IPNet
}

// NewSecurityMiddleware creates a new security middleware
func NewSecurityMiddleware(config SecurityConfig) (*SecurityMiddleware, error) {
	sm := &SecurityMiddleware{
		config: config,
	}

	// Parse trusted proxy networks
	for _, proxy := range config.TrustedProxies {
		_, network, err := net.ParseCIDR(proxy)
		if err != nil {
			// Try parsing as single IP
			ip := net.ParseIP(proxy)
			if ip == nil {
				return nil, fmt.Errorf("invalid trusted proxy: %s", proxy)
			}
			if ip.To4() != nil {
				_, network, _ = net.ParseCIDR(proxy + "/32")
			} else {
				_, network, _ = net.ParseCIDR(proxy + "/128")
			}
		}
		sm.trustedNets = append(sm.trustedNets, network)
	}

	// Parse blocked IP networks
	for _, blockedIP := range config.BlockedIPs {
		_, network, err := net.ParseCIDR(blockedIP)
		if err != nil {
			// Try parsing as single IP
			ip := net.ParseIP(blockedIP)
			if ip == nil {
				return nil, fmt.Errorf("invalid blocked IP: %s", blockedIP)
			}
			if ip.To4() != nil {
				_, network, _ = net.ParseCIDR(blockedIP + "/32")
			} else {
				_, network, _ = net.ParseCIDR(blockedIP + "/128")
			}
		}
		sm.blockedNets = append(sm.blockedNets, network)
	}

	return sm, nil
}

// AuthenticationMiddleware handles JWT authentication
func (sm *SecurityMiddleware) AuthenticationMiddleware() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip authentication for public endpoints
			if sm.isPublicEndpoint(r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}

			// Extract token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				sm.logSecurityEvent("AUTHENTICATION", "MISSING_TOKEN", "FAILURE", nil, r)
				http.Error(w, "Authorization header required", http.StatusUnauthorized)
				return
			}

			// Parse Bearer token
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				sm.logSecurityEvent("AUTHENTICATION", "INVALID_TOKEN_FORMAT", "FAILURE", nil, r)
				http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
				return
			}

			token := parts[1]

			// Validate token
			claims, err := sm.config.JWTManager.ValidateToken(token)
			if err != nil {
				sm.logSecurityEvent("AUTHENTICATION", "TOKEN_VALIDATION", "FAILURE", nil, r)
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			// Add claims to request context
			ctx := context.WithValue(r.Context(), "user_claims", claims)
			ctx = context.WithValue(ctx, "user_id", claims.UserID)
			ctx = context.WithValue(ctx, "session_id", claims.SessionID)

			sm.logSecurityEvent("AUTHENTICATION", "TOKEN_VALIDATION", "SUCCESS", &claims.UserID, r)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// AuthorizationMiddleware handles role-based access control
func (sm *SecurityMiddleware) AuthorizationMiddleware(requiredRole string, requiredPermissions ...string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := r.Context().Value("user_claims").(*security.TokenClaims)
			if !ok {
				sm.logSecurityEvent("AUTHORIZATION", "MISSING_CLAIMS", "DENIED", nil, r)
				http.Error(w, "Authentication required", http.StatusUnauthorized)
				return
			}

			// Check role
			if requiredRole != "" && claims.Role != requiredRole && claims.Role != "admin" {
				sm.logSecurityEvent("AUTHORIZATION", "INSUFFICIENT_ROLE", "DENIED", &claims.UserID, r)
				http.Error(w, "Insufficient privileges", http.StatusForbidden)
				return
			}

			// Check permissions
			if len(requiredPermissions) > 0 {
				userPermissions := make(map[string]bool)
				for _, perm := range claims.Permissions {
					userPermissions[perm] = true
				}

				for _, requiredPerm := range requiredPermissions {
					if !userPermissions[requiredPerm] && claims.Role != "admin" {
						sm.logSecurityEvent("AUTHORIZATION", "INSUFFICIENT_PERMISSIONS", "DENIED", &claims.UserID, r)
						http.Error(w, "Insufficient permissions", http.StatusForbidden)
						return
					}
				}
			}

			sm.logSecurityEvent("AUTHORIZATION", "ACCESS_GRANTED", "SUCCESS", &claims.UserID, r)
			next.ServeHTTP(w, r)
		})
	}
}

// InputValidationMiddleware validates and sanitizes input
func (sm *SecurityMiddleware) InputValidationMiddleware() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Validate request size
			if r.ContentLength > sm.config.MaxRequestSize {
				sm.logSecurityEvent("INPUT_VALIDATION", "REQUEST_TOO_LARGE", "BLOCKED", nil, r)
				http.Error(w, "Request too large", http.StatusRequestEntityTooLarge)
				return
			}

			// Validate Content-Type for POST/PUT/PATCH
			if r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodPatch {
				contentType := r.Header.Get("Content-Type")
				if !sm.isValidContentType(contentType) {
					sm.logSecurityEvent("INPUT_VALIDATION", "INVALID_CONTENT_TYPE", "BLOCKED", nil, r)
					http.Error(w, "Invalid content type", http.StatusUnsupportedMediaType)
					return
				}
			}

			// Validate query parameters
			for key, values := range r.URL.Query() {
				for _, value := range values {
					if err := sm.config.Validator.ValidateString(key, value, 0, 1000, false); err != nil {
						sm.logSecurityEvent("INPUT_VALIDATION", "MALICIOUS_QUERY_PARAM", "BLOCKED", nil, r)
						http.Error(w, "Invalid query parameter", http.StatusBadRequest)
						return
					}
				}
			}

			// Validate headers
			for _, values := range r.Header {
				for _, value := range values {
					if sm.containsMaliciousContent(value) {
						sm.logSecurityEvent("INPUT_VALIDATION", "MALICIOUS_HEADER", "BLOCKED", nil, r)
						http.Error(w, "Invalid header content", http.StatusBadRequest)
						return
					}
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

// IPFilteringMiddleware blocks requests from blacklisted IPs
func (sm *SecurityMiddleware) IPFilteringMiddleware() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientIP := sm.getClientIP(r)
			ip := net.ParseIP(clientIP)
			if ip == nil {
				sm.logSecurityEvent("IP_FILTERING", "INVALID_IP", "BLOCKED", nil, r)
				http.Error(w, "Invalid IP address", http.StatusBadRequest)
				return
			}

			// Check if IP is blocked
			for _, blockedNet := range sm.blockedNets {
				if blockedNet.Contains(ip) {
					sm.logSecurityEvent("IP_FILTERING", "BLOCKED_IP", "BLOCKED", nil, r)
					http.Error(w, "Access denied", http.StatusForbidden)
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

// CSRFProtectionMiddleware provides CSRF protection
func (sm *SecurityMiddleware) CSRFProtectionMiddleware() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !sm.config.EnableCSRF {
				next.ServeHTTP(w, r)
				return
			}

			// Skip CSRF for GET, HEAD, OPTIONS
			if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions {
				next.ServeHTTP(w, r)
				return
			}

			// Check CSRF token
			token := r.Header.Get("X-CSRF-Token")
			if token == "" {
				token = r.FormValue("csrf_token")
			}

			if token == "" {
				sm.logSecurityEvent("CSRF", "MISSING_TOKEN", "BLOCKED", nil, r)
				http.Error(w, "CSRF token required", http.StatusForbidden)
				return
			}

			// Validate CSRF token (simplified - in production, use proper CSRF validation)
			if !sm.validateCSRFToken(token, r) {
				sm.logSecurityEvent("CSRF", "INVALID_TOKEN", "BLOCKED", nil, r)
				http.Error(w, "Invalid CSRF token", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// SecurityHeadersEnhanced adds comprehensive security headers
func (sm *SecurityMiddleware) SecurityHeadersEnhanced() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Content Security Policy
			csp := "default-src 'self'; " +
				"script-src 'self' 'unsafe-inline' 'unsafe-eval'; " +
				"style-src 'self' 'unsafe-inline'; " +
				"img-src 'self' data: https:; " +
				"font-src 'self' data:; " +
				"connect-src 'self'; " +
				"frame-ancestors 'none'; " +
				"base-uri 'self'; " +
				"form-action 'self'"

			w.Header().Set("Content-Security-Policy", csp)
			w.Header().Set("X-Content-Security-Policy", csp)
			w.Header().Set("X-WebKit-CSP", csp)

			// Security headers
			w.Header().Set("X-Frame-Options", "DENY")
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("X-XSS-Protection", "1; mode=block")
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
			w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
			w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

			// Remove server information
			w.Header().Set("Server", "")
			w.Header().Del("X-Powered-By")

			next.ServeHTTP(w, r)
		})
	}
}

// Helper methods

func (sm *SecurityMiddleware) isPublicEndpoint(path string) bool {
	publicEndpoints := []string{
		"/health",
		"/api/auth/login",
		"/api/auth/register",
		"/api/destinations", // Public read access
		"/metrics",
	}

	for _, endpoint := range publicEndpoints {
		if strings.HasPrefix(path, endpoint) {
			return true
		}
	}
	return false
}

func (sm *SecurityMiddleware) isValidContentType(contentType string) bool {
	validTypes := []string{
		"application/json",
		"application/x-www-form-urlencoded",
		"multipart/form-data",
		"text/plain",
	}

	for _, validType := range validTypes {
		if strings.HasPrefix(contentType, validType) {
			return true
		}
	}
	return false
}

func (sm *SecurityMiddleware) containsMaliciousContent(content string) bool {
	maliciousPatterns := []string{
		"<script", "javascript:", "vbscript:", "onload=", "onerror=",
		"eval(", "expression(", "data:text/html", "../", "..\\",
	}

	lowerContent := strings.ToLower(content)
	for _, pattern := range maliciousPatterns {
		if strings.Contains(lowerContent, pattern) {
			return true
		}
	}
	return false
}

func (sm *SecurityMiddleware) getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		ips := strings.Split(xff, ",")
		clientIP := strings.TrimSpace(ips[0])

		// Validate that the request came through a trusted proxy
		if sm.isTrustedProxy(r.RemoteAddr) {
			return clientIP
		}
	}

	// Check X-Real-IP header
	xri := r.Header.Get("X-Real-IP")
	if xri != "" && sm.isTrustedProxy(r.RemoteAddr) {
		return xri
	}

	// Fall back to RemoteAddr
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func (sm *SecurityMiddleware) isTrustedProxy(remoteAddr string) bool {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		return false
	}

	ip := net.ParseIP(host)
	if ip == nil {
		return false
	}

	for _, trustedNet := range sm.trustedNets {
		if trustedNet.Contains(ip) {
			return true
		}
	}
	return false
}

func (sm *SecurityMiddleware) validateCSRFToken(token string, r *http.Request) bool {
	// Simplified CSRF validation - in production, implement proper token validation
	// This should validate against a secure token stored in session or cookie
	return len(token) >= 32 // Basic length check
}

func (sm *SecurityMiddleware) logSecurityEvent(eventType, action, result string, userID *int64, r *http.Request) {
	if sm.config.AuditLogger == nil {
		return
	}

	sessionID := ""
	if claims, ok := r.Context().Value("user_claims").(*security.TokenClaims); ok {
		sessionID = claims.SessionID
	}

	details := map[string]interface{}{
		"method":         r.Method,
		"path":           r.URL.Path,
		"query":          r.URL.RawQuery,
		"user_agent":     r.UserAgent(),
		"referer":        r.Referer(),
		"content_length": r.ContentLength,
	}

	sm.config.AuditLogger.LogSecurityEvent(
		eventType,
		action,
		result,
		userID,
		sessionID,
		sm.getClientIP(r),
		r.UserAgent(),
		details,
	)
}
