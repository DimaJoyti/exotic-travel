package middleware

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/google/uuid"
)

// Middleware represents a middleware function
type Middleware func(http.Handler) http.Handler

// Chain applies multiple middleware functions to a handler
func Chain(h http.Handler, middlewares ...Middleware) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}

// CORS adds CORS headers to responses
func CORS() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Set CORS headers
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-User-ID, X-Session-ID, X-Request-ID")
			w.Header().Set("Access-Control-Expose-Headers", "X-Request-ID")
			w.Header().Set("Access-Control-Max-Age", "86400")
			
			// Handle preflight requests
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusOK)
				return
			}
			
			next.ServeHTTP(w, r)
		})
	}
}

// Logging logs HTTP requests
func Logging() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			
			// Create a response writer wrapper to capture status code
			wrapped := &responseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}
			
			// Get request ID from context
			requestID := getRequestID(r.Context())
			
			// Log request
			log.Printf("[%s] %s %s %s - Started", requestID, r.Method, r.URL.Path, r.RemoteAddr)
			
			// Process request
			next.ServeHTTP(wrapped, r)
			
			// Log response
			duration := time.Since(start)
			log.Printf("[%s] %s %s %s - %d %s (%v)", 
				requestID, r.Method, r.URL.Path, r.RemoteAddr, 
				wrapped.statusCode, http.StatusText(wrapped.statusCode), duration)
		})
	}
}

// Recovery recovers from panics and returns a 500 error
func Recovery() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					requestID := getRequestID(r.Context())
					
					// Log the panic
					log.Printf("[%s] PANIC: %v\n%s", requestID, err, debug.Stack())
					
					// Return 500 error
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusInternalServerError)
					fmt.Fprintf(w, `{"error": "Internal server error", "request_id": "%s"}`, requestID)
				}
			}()
			
			next.ServeHTTP(w, r)
		})
	}
}

// RequestID adds a unique request ID to each request
func RequestID() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check if request ID is already provided
			requestID := r.Header.Get("X-Request-ID")
			if requestID == "" {
				requestID = uuid.New().String()
			}
			
			// Add request ID to context
			ctx := context.WithValue(r.Context(), "request_id", requestID)
			r = r.WithContext(ctx)
			
			// Add request ID to response headers
			w.Header().Set("X-Request-ID", requestID)
			
			next.ServeHTTP(w, r)
		})
	}
}

// Timeout adds a timeout to requests
func Timeout(timeout time.Duration) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Create context with timeout
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer cancel()
			
			// Update request with new context
			r = r.WithContext(ctx)
			
			// Channel to signal completion
			done := make(chan struct{})
			
			// Run the handler in a goroutine
			go func() {
				defer func() {
					if err := recover(); err != nil {
						requestID := getRequestID(r.Context())
						log.Printf("[%s] PANIC in timeout handler: %v", requestID, err)
					}
					close(done)
				}()
				
				next.ServeHTTP(w, r)
			}()
			
			// Wait for completion or timeout
			select {
			case <-done:
				// Request completed normally
				return
			case <-ctx.Done():
				// Request timed out
				requestID := getRequestID(r.Context())
				log.Printf("[%s] Request timed out after %v", requestID, timeout)
				
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusRequestTimeout)
				fmt.Fprintf(w, `{"error": "Request timeout", "request_id": "%s"}`, requestID)
				return
			}
		})
	}
}

// RateLimit implements basic rate limiting (simplified version)
func RateLimit(requestsPerMinute int) Middleware {
	// This is a simplified rate limiter for demonstration
	// In production, use a proper rate limiting library like golang.org/x/time/rate
	
	type client struct {
		requests  int
		lastReset time.Time
	}
	
	clients := make(map[string]*client)
	
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Use IP address as client identifier
			clientIP := getClientIP(r)
			now := time.Now()
			
			// Get or create client record
			c, exists := clients[clientIP]
			if !exists {
				c = &client{
					requests:  0,
					lastReset: now,
				}
				clients[clientIP] = c
			}
			
			// Reset counter if a minute has passed
			if now.Sub(c.lastReset) >= time.Minute {
				c.requests = 0
				c.lastReset = now
			}
			
			// Check rate limit
			if c.requests >= requestsPerMinute {
				requestID := getRequestID(r.Context())
				log.Printf("[%s] Rate limit exceeded for %s", requestID, clientIP)
				
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				fmt.Fprintf(w, `{"error": "Rate limit exceeded", "request_id": "%s"}`, requestID)
				return
			}
			
			// Increment request counter
			c.requests++
			
			next.ServeHTTP(w, r)
		})
	}
}

// Authentication validates API keys or tokens (placeholder implementation)
func Authentication() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip authentication for health checks and root endpoint
			if r.URL.Path == "/health" || r.URL.Path == "/" {
				next.ServeHTTP(w, r)
				return
			}
			
			// Check for API key in header
			apiKey := r.Header.Get("Authorization")
			if apiKey == "" {
				apiKey = r.Header.Get("X-API-Key")
			}
			
			// For demo purposes, we'll skip actual validation
			// In production, validate the API key against a database or service
			
			// Add user information to context if authenticated
			if apiKey != "" {
				// Extract user ID from API key (placeholder)
				userID := "user_123" // This would be extracted from the validated API key
				ctx := context.WithValue(r.Context(), "user_id", userID)
				r = r.WithContext(ctx)
			}
			
			next.ServeHTTP(w, r)
		})
	}
}

// ContentType validates and sets content type for JSON APIs
func ContentType() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// For POST and PUT requests, validate content type
			if r.Method == http.MethodPost || r.Method == http.MethodPut {
				contentType := r.Header.Get("Content-Type")
				if contentType != "application/json" && contentType != "" {
					requestID := getRequestID(r.Context())
					log.Printf("[%s] Invalid content type: %s", requestID, contentType)
					
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusUnsupportedMediaType)
					fmt.Fprintf(w, `{"error": "Content-Type must be application/json", "request_id": "%s"}`, requestID)
					return
				}
			}
			
			next.ServeHTTP(w, r)
		})
	}
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// getRequestID extracts request ID from context
func getRequestID(ctx context.Context) string {
	if requestID, ok := ctx.Value("request_id").(string); ok {
		return requestID
	}
	return "unknown"
}

// getClientIP extracts client IP address from request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}
	
	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	
	// Fall back to RemoteAddr
	return r.RemoteAddr
}
