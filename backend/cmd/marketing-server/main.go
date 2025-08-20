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

	"github.com/exotic-travel-booking/backend/internal/api/handlers"
	"github.com/exotic-travel-booking/backend/internal/api/middleware"
	"github.com/exotic-travel-booking/backend/internal/database"
	"github.com/exotic-travel-booking/backend/internal/llm/providers"
	"github.com/exotic-travel-booking/backend/internal/marketing/agents"
	"github.com/exotic-travel-booking/backend/internal/marketing/content"
	"github.com/exotic-travel-booking/backend/internal/marketing/repository"
	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

func main() {
	// Initialize OpenTelemetry
	if err := initTracing(); err != nil {
		log.Fatalf("Failed to initialize tracing: %v", err)
	}

	// Initialize database
	db, err := initDatabase()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Initialize LLM provider
	llmProvider, err := initLLMProvider()
	if err != nil {
		log.Fatalf("Failed to initialize LLM provider: %v", err)
	}

	// Initialize repositories
	marketingRepo := repository.NewMarketingRepository(db)

	// Initialize agents
	contentAgent := agents.NewContentAgent(llmProvider)

	// Initialize services
	contentGenerator := content.NewGenerator(contentAgent, marketingRepo)

	// Initialize handlers
	marketingHandler := handlers.NewMarketingHandler(contentGenerator)

	// Setup HTTP server
	server := setupServer(marketingHandler)

	// Start server
	go func() {
		log.Printf("🚀 Marketing AI Server starting on port %s", getPort())
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 Shutting down Marketing AI Server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("✅ Marketing AI Server exited")
}

func initTracing() error {
	// Create Jaeger exporter
	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint("http://localhost:14268/api/traces")))
	if err != nil {
		return fmt.Errorf("failed to create Jaeger exporter: %w", err)
	}

	// Create trace provider
	tp := trace.NewTracerProvider(
		trace.WithBatcher(exp),
		trace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("marketing-ai-server"),
			semconv.ServiceVersionKey.String("1.0.0"),
			attribute.String("environment", getEnv("ENVIRONMENT", "development")),
		)),
	)

	// Set global trace provider
	otel.SetTracerProvider(tp)

	return nil
}

func initDatabase() (*sqlx.DB, error) {
	dbConfig := database.Config{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnv("DB_PORT", "5432"),
		User:     getEnv("DB_USER", "postgres"),
		Password: getEnv("DB_PASSWORD", "password"),
		DBName:   getEnv("DB_NAME", "exotic_travel"),
		SSLMode:  getEnv("DB_SSLMODE", "disable"),
	}

	db, err := database.Connect(dbConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("✅ Database connected successfully")
	return db, nil
}

func initLLMProvider() (providers.LLMProvider, error) {
	providerType := getEnv("LLM_PROVIDER", "openai")
	
	switch providerType {
	case "openai":
		apiKey := getEnv("OPENAI_API_KEY", "")
		if apiKey == "" {
			return nil, fmt.Errorf("OPENAI_API_KEY is required")
		}
		
		config := providers.OpenAIConfig{
			APIKey:      apiKey,
			Model:       getEnv("OPENAI_MODEL", "gpt-4"),
			BaseURL:     getEnv("OPENAI_BASE_URL", "https://api.openai.com/v1"),
			MaxTokens:   2000,
			Temperature: 0.7,
		}
		
		provider, err := providers.NewOpenAIProvider(config)
		if err != nil {
			return nil, fmt.Errorf("failed to create OpenAI provider: %w", err)
		}
		
		log.Println("✅ OpenAI LLM provider initialized")
		return provider, nil
		
	case "anthropic":
		apiKey := getEnv("ANTHROPIC_API_KEY", "")
		if apiKey == "" {
			return nil, fmt.Errorf("ANTHROPIC_API_KEY is required")
		}
		
		config := providers.AnthropicConfig{
			APIKey:      apiKey,
			Model:       getEnv("ANTHROPIC_MODEL", "claude-3-sonnet-20240229"),
			BaseURL:     getEnv("ANTHROPIC_BASE_URL", "https://api.anthropic.com"),
			MaxTokens:   2000,
			Temperature: 0.7,
		}
		
		provider, err := providers.NewAnthropicProvider(config)
		if err != nil {
			return nil, fmt.Errorf("failed to create Anthropic provider: %w", err)
		}
		
		log.Println("✅ Anthropic LLM provider initialized")
		return provider, nil
		
	default:
		return nil, fmt.Errorf("unsupported LLM provider: %s", providerType)
	}
}

func setupServer(marketingHandler *handlers.MarketingHandler) *http.Server {
	mux := http.NewServeMux()

	// Apply middleware
	handler := middleware.Chain(
		mux,
		middleware.CORS(),
		middleware.Logging(),
		middleware.Recovery(),
		middleware.RequestID(),
		middleware.Timeout(30*time.Second),
	)

	// Marketing API routes
	mux.HandleFunc("/api/v1/marketing/health", marketingHandler.HealthCheck)
	mux.HandleFunc("/api/v1/marketing/content/generate", marketingHandler.GenerateContent)
	mux.HandleFunc("/api/v1/marketing/content/history/", marketingHandler.GetContentHistory)
	mux.HandleFunc("/api/v1/marketing/content/", func(w http.ResponseWriter, r *http.Request) {
		// Handle regenerate content endpoint
		if r.Method == http.MethodPost && r.URL.Path[len("/api/v1/marketing/content/"):] != "" {
			marketingHandler.RegenerateContent(w, r)
		} else {
			http.NotFound(w, r)
		}
	})

	// Root endpoint
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		response := map[string]interface{}{
			"service":   "Marketing AI API",
			"version":   "1.0.0",
			"status":    "operational",
			"timestamp": time.Now(),
			"endpoints": map[string]string{
				"health":           "/api/v1/marketing/health",
				"generate_content": "/api/v1/marketing/content/generate",
				"content_history":  "/api/v1/marketing/content/history/{campaignId}",
				"regenerate":       "/api/v1/marketing/content/{contentId}/regenerate",
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		
		// Simple JSON encoding
		fmt.Fprintf(w, `{
			"service": "%s",
			"version": "%s",
			"status": "%s",
			"timestamp": "%s"
		}`, response["service"], response["version"], response["status"], response["timestamp"])
	})

	return &http.Server{
		Addr:         ":" + getPort(),
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getPort() string {
	return getEnv("PORT", "8081")
}
