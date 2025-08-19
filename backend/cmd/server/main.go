package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/exotic-travel-booking/backend/internal/config"
	"github.com/exotic-travel-booking/backend/internal/handlers"
	"github.com/exotic-travel-booking/backend/internal/metrics"
	"github.com/exotic-travel-booking/backend/internal/middleware"
	"github.com/exotic-travel-booking/backend/internal/repositories"
	"github.com/exotic-travel-booking/backend/internal/services"
	"github.com/exotic-travel-booking/backend/pkg/database"
	"github.com/exotic-travel-booking/backend/pkg/observability"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize tracing
	cleanup, err := observability.InitTracing("exotic-travel-booking", cfg.Environment)
	if err != nil {
		log.Fatalf("Failed to initialize tracing: %v", err)
	}
	defer cleanup()

	// Connect to database
	db, err := database.NewConnection(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize repositories
	userRepo := repositories.NewUserRepository(db)
	destinationRepo := repositories.NewDestinationRepository(db)
	bookingRepo := repositories.NewBookingRepository(db)

	// Initialize services
	authService := services.NewAuthService(userRepo, cfg.JWTSecret)
	destinationService := services.NewDestinationService(destinationRepo)
	bookingService := services.NewBookingService(bookingRepo, destinationRepo)

	// Initialize handlers
	authHandlers := handlers.NewAuthHandlers(authService)
	destinationHandlers := handlers.NewDestinationHandlers(destinationService)
	bookingHandlers := handlers.NewBookingHandlers(bookingService)
	imageHandler := handlers.NewImageHandler("./uploads", "http://localhost:8080")

	// Create HTTP server with new ServeMux
	mux := http.NewServeMux()

	// Initialize performance monitoring
	metrics.InitGlobalCollector()
	defer metrics.StopGlobalCollector()

	// Setup rate limiter (10 requests per second, burst of 20)
	rateLimiter := middleware.NewRateLimiter(10.0, 20)

	// Setup circuit breaker
	circuitBreaker := middleware.NewCircuitBreaker(5, 30*time.Second)

	// Setup middleware
	handler := middleware.Chain(
		mux,
		middleware.SecurityHeaders(),
		middleware.RequestID(),
		middleware.PerformanceMiddleware(),
		middleware.PerformanceHeaders(),
		rateLimiter.Middleware,
		circuitBreaker.CircuitBreakerMiddleware(),
		middleware.InputValidation(),
		middleware.TimeoutMiddleware(30*time.Second),
		middleware.RequestSizeLimit(10<<20), // 10MB
		middleware.CompressionMiddleware(),
		middleware.CachingMiddleware(3600), // 1 hour cache for static content
		middleware.HealthCheck("/health"),
		middleware.Tracing(),
		middleware.CORS(),
		middleware.Logging(),
		middleware.Recovery(),
	)

	// Setup routes
	setupRoutes(mux, authHandlers, destinationHandlers, bookingHandlers, imageHandler, authService)

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Server starting on port %d", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Create a deadline for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

func setupRoutes(mux *http.ServeMux, authHandlers *handlers.AuthHandlers, destinationHandlers *handlers.DestinationHandlers, bookingHandlers *handlers.BookingHandlers, imageHandler *handlers.ImageHandler, authService *services.AuthService) {
	// Health check endpoint
	mux.HandleFunc("GET /health", handlers.HealthCheck)

	// Auth routes (public)
	mux.HandleFunc("POST /api/auth/register", authHandlers.Register)
	mux.HandleFunc("POST /api/auth/login", authHandlers.Login)
	mux.HandleFunc("POST /api/auth/refresh", authHandlers.RefreshToken)

	// Protected auth routes
	authMiddleware := middleware.AuthMiddleware(authService)
	mux.Handle("GET /api/auth/me", authMiddleware(http.HandlerFunc(authHandlers.Me)))

	// Destinations routes (public read, admin write)
	mux.HandleFunc("GET /api/destinations", destinationHandlers.List)
	mux.HandleFunc("GET /api/destinations/search", destinationHandlers.Search)
	mux.HandleFunc("GET /api/destinations/{id}", destinationHandlers.GetByID)

	// Admin-only destination routes
	adminMiddleware := middleware.Chain(
		http.HandlerFunc(destinationHandlers.Create),
		authMiddleware,
		middleware.AdminMiddleware(),
	)
	mux.Handle("POST /api/destinations", adminMiddleware)

	adminUpdateMiddleware := middleware.Chain(
		http.HandlerFunc(destinationHandlers.Update),
		authMiddleware,
		middleware.AdminMiddleware(),
	)
	mux.Handle("PUT /api/destinations/{id}", adminUpdateMiddleware)

	adminDeleteMiddleware := middleware.Chain(
		http.HandlerFunc(destinationHandlers.Delete),
		authMiddleware,
		middleware.AdminMiddleware(),
	)
	mux.Handle("DELETE /api/destinations/{id}", adminDeleteMiddleware)

	// Bookings routes (protected)
	mux.Handle("GET /api/bookings", authMiddleware(http.HandlerFunc(bookingHandlers.List)))
	mux.Handle("POST /api/bookings", authMiddleware(http.HandlerFunc(bookingHandlers.Create)))
	mux.Handle("GET /api/bookings/{id}", authMiddleware(http.HandlerFunc(bookingHandlers.GetByID)))
	mux.Handle("PUT /api/bookings/{id}", authMiddleware(http.HandlerFunc(bookingHandlers.Update)))
	mux.Handle("DELETE /api/bookings/{id}", authMiddleware(http.HandlerFunc(bookingHandlers.Delete)))

	// Image routes
	mux.HandleFunc("POST /api/images/upload", imageHandler.UploadImage)
	mux.HandleFunc("GET /api/images", imageHandler.GetImages)
	mux.HandleFunc("GET /api/images/{id}", imageHandler.GetImage)
	mux.Handle("PATCH /api/images/{id}", authMiddleware(http.HandlerFunc(imageHandler.UpdateImageMetadata)))
	mux.Handle("DELETE /api/images/{id}", authMiddleware(http.HandlerFunc(imageHandler.DeleteImage)))

	// Static file serving for uploads
	mux.Handle("GET /uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir("./uploads/"))))
}
