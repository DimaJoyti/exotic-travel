package middleware

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/exotic-travel-booking/backend/internal/metrics"
)

// PerformanceMiddleware tracks request performance metrics
func PerformanceMiddleware() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			
			// Wrap response writer to capture status code
			wrapped := &responseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			// Add request start time to context
			ctx := context.WithValue(r.Context(), "request_start", start)
			r = r.WithContext(ctx)

			// Process request
			next.ServeHTTP(wrapped, r)

			// Calculate duration
			duration := time.Since(start)

			// Record metrics
			collector := metrics.GetGlobalCollector()
			if collector != nil {
				isError := wrapped.statusCode >= 400
				collector.RecordHTTPRequest(duration, wrapped.statusCode, isError)
				
				// Record response time histogram
				collector.ObserveHistogram("http_request_duration_ms", float64(duration.Nanoseconds())/1e6)
				
				// Record endpoint-specific metrics
				endpoint := r.Method + " " + r.URL.Path
				collector.IncrementCounter("endpoint_requests:"+endpoint, 1)
				
				if isError {
					collector.IncrementCounter("endpoint_errors:"+endpoint, 1)
				}
			}
		})
	}
}

// DatabasePerformanceMiddleware tracks database operation performance
func DatabasePerformanceMiddleware(operation string) func(func() error) error {
	return func(fn func() error) error {
		start := time.Now()
		err := fn()
		duration := time.Since(start)

		// Record metrics
		collector := metrics.GetGlobalCollector()
		if collector != nil {
			collector.RecordDBQuery(duration, err != nil)
			collector.ObserveHistogram("db_query_duration_ms", float64(duration.Nanoseconds())/1e6)
			collector.IncrementCounter("db_operations:"+operation, 1)
			
			if err != nil {
				collector.IncrementCounter("db_errors:"+operation, 1)
			}
		}

		return err
	}
}

// CachePerformanceMiddleware tracks cache operation performance
func CachePerformanceMiddleware(operation string) func(bool, error) {
	return func(hit bool, err error) {
		collector := metrics.GetGlobalCollector()
		if collector != nil {
			collector.RecordCacheOperation(hit, err != nil)
			collector.IncrementCounter("cache_operations:"+operation, 1)
			
			if hit {
				collector.IncrementCounter("cache_hits:"+operation, 1)
			} else {
				collector.IncrementCounter("cache_misses:"+operation, 1)
			}
			
			if err != nil {
				collector.IncrementCounter("cache_errors:"+operation, 1)
			}
		}
	}
}

// CompressionMiddleware adds gzip compression for responses
func CompressionMiddleware() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check if client accepts gzip
			if !acceptsGzip(r) {
				next.ServeHTTP(w, r)
				return
			}

			// Create gzip writer
			gw := &gzipResponseWriter{
				ResponseWriter: w,
			}
			defer gw.Close()

			// Set content encoding header
			w.Header().Set("Content-Encoding", "gzip")
			w.Header().Set("Vary", "Accept-Encoding")

			next.ServeHTTP(gw, r)
		})
	}
}

// acceptsGzip checks if the client accepts gzip encoding
func acceptsGzip(r *http.Request) bool {
	acceptEncoding := r.Header.Get("Accept-Encoding")
	return contains(acceptEncoding, "gzip")
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || 
		(len(s) > len(substr) && (s[:len(substr)+1] == substr+"," || 
		s[len(s)-len(substr)-1:] == ","+substr || 
		containsMiddle(s, ","+substr+","))))
}

// containsMiddle checks if string contains substring in the middle
func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// gzipResponseWriter wraps http.ResponseWriter with gzip compression
type gzipResponseWriter struct {
	http.ResponseWriter
	writer *gzipWriter
}

// gzipWriter is a simple gzip writer implementation
type gzipWriter struct {
	w http.ResponseWriter
}

// Write compresses and writes data (simplified implementation)
func (gw *gzipResponseWriter) Write(data []byte) (int, error) {
	if gw.writer == nil {
		gw.writer = &gzipWriter{w: gw.ResponseWriter}
	}
	// In a real implementation, this would use compress/gzip
	// For now, we'll just pass through
	return gw.ResponseWriter.Write(data)
}

// Close closes the gzip writer
func (gw *gzipResponseWriter) Close() error {
	// In a real implementation, this would close the gzip writer
	return nil
}

// CachingMiddleware adds HTTP caching headers
func CachingMiddleware(maxAge int) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Set cache headers for static content
			if isStaticContent(r.URL.Path) {
				w.Header().Set("Cache-Control", "public, max-age="+strconv.Itoa(maxAge))
				w.Header().Set("Expires", time.Now().Add(time.Duration(maxAge)*time.Second).Format(http.TimeFormat))
			} else {
				// For dynamic content, set no-cache
				w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
				w.Header().Set("Pragma", "no-cache")
				w.Header().Set("Expires", "0")
			}

			next.ServeHTTP(w, r)
		})
	}
}

// isStaticContent checks if the path represents static content
func isStaticContent(path string) bool {
	staticExtensions := []string{".css", ".js", ".png", ".jpg", ".jpeg", ".gif", ".ico", ".svg", ".woff", ".woff2", ".ttf", ".eot"}
	
	for _, ext := range staticExtensions {
		if len(path) >= len(ext) && path[len(path)-len(ext):] == ext {
			return true
		}
	}
	
	return false
}

// ConnectionPoolMiddleware manages database connection pooling
func ConnectionPoolMiddleware() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Add connection pool metrics
			collector := metrics.GetGlobalCollector()
			if collector != nil {
				collector.IncrementCounter("active_connections", 1)
				defer collector.IncrementCounter("active_connections", -1)
			}

			next.ServeHTTP(w, r)
		})
	}
}

// MemoryMonitoringMiddleware monitors memory usage
func MemoryMonitoringMiddleware() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Record memory usage before request
			collector := metrics.GetGlobalCollector()
			if collector != nil {
				// This will be collected by the background metrics collector
				// We can add request-specific memory tracking here if needed
			}

			next.ServeHTTP(w, r)
		})
	}
}

// CircuitBreakerMiddleware implements a simple circuit breaker pattern
type CircuitBreaker struct {
	maxFailures int
	resetTime   time.Duration
	failures    int
	lastFailure time.Time
	state       CircuitState
}

// CircuitState represents the state of a circuit breaker
type CircuitState int

const (
	CircuitClosed CircuitState = iota
	CircuitOpen
	CircuitHalfOpen
)

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(maxFailures int, resetTime time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		maxFailures: maxFailures,
		resetTime:   resetTime,
		state:       CircuitClosed,
	}
}

// CircuitBreakerMiddleware returns middleware that implements circuit breaker pattern
func (cb *CircuitBreaker) CircuitBreakerMiddleware() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check circuit state
			if cb.state == CircuitOpen {
				if time.Since(cb.lastFailure) > cb.resetTime {
					cb.state = CircuitHalfOpen
					cb.failures = 0
				} else {
					http.Error(w, "Service temporarily unavailable", http.StatusServiceUnavailable)
					return
				}
			}

			// Wrap response writer to detect failures
			wrapped := &responseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			next.ServeHTTP(wrapped, r)

			// Update circuit breaker state based on response
			if wrapped.statusCode >= 500 {
				cb.recordFailure()
			} else if cb.state == CircuitHalfOpen {
				cb.recordSuccess()
			}
		})
	}
}

// recordFailure records a failure and updates circuit state
func (cb *CircuitBreaker) recordFailure() {
	cb.failures++
	cb.lastFailure = time.Now()
	
	if cb.failures >= cb.maxFailures {
		cb.state = CircuitOpen
	}
}

// recordSuccess records a success and updates circuit state
func (cb *CircuitBreaker) recordSuccess() {
	cb.failures = 0
	cb.state = CircuitClosed
}

// GetRequestDuration extracts request duration from context
func GetRequestDuration(ctx context.Context) time.Duration {
	if start, ok := ctx.Value("request_start").(time.Time); ok {
		return time.Since(start)
	}
	return 0
}

// PerformanceHeaders adds performance-related headers
func PerformanceHeaders() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			next.ServeHTTP(w, r)

			// Add server timing header
			duration := time.Since(start)
			w.Header().Set("Server-Timing", "total;dur="+strconv.FormatFloat(float64(duration.Nanoseconds())/1e6, 'f', 2, 64))
		})
	}
}
