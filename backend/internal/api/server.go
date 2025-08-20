package api

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/exotic-travel-booking/backend/internal/api/handlers"
	"github.com/exotic-travel-booking/backend/internal/api/middleware"
	"github.com/exotic-travel-booking/backend/internal/llm"
	"github.com/exotic-travel-booking/backend/internal/llm/providers"
	"github.com/exotic-travel-booking/backend/internal/services"
	"github.com/exotic-travel-booking/backend/internal/tools"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// Server represents the API server
type Server struct {
	httpServer   *http.Server
	llmManager   *llm.LLMManager
	toolRegistry *tools.ToolRegistry
	tracer       trace.Tracer
}

// Config represents server configuration
type Config struct {
	Port         string        `json:"port"`
	Host         string        `json:"host"`
	ReadTimeout  time.Duration `json:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout"`
	IdleTimeout  time.Duration `json:"idle_timeout"`
}

// NewServer creates a new API server
func NewServer(config *Config) (*Server, error) {
	// Initialize LLM manager
	llmManager := llm.NewLLMManager()

	// Initialize tool registry
	toolRegistry := tools.NewToolRegistry()

	// Initialize tools
	if err := initializeTools(toolRegistry); err != nil {
		return nil, fmt.Errorf("failed to initialize tools: %w", err)
	}

	// Initialize LLM providers
	if err := initializeLLMProviders(llmManager); err != nil {
		return nil, fmt.Errorf("failed to initialize LLM providers: %w", err)
	}

	// Create router
	router := setupRouter(llmManager, toolRegistry)

	// Create HTTP server
	httpServer := &http.Server{
		Addr:         fmt.Sprintf("%s:%s", config.Host, config.Port),
		Handler:      router,
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
		IdleTimeout:  config.IdleTimeout,
	}

	return &Server{
		httpServer:   httpServer,
		llmManager:   llmManager,
		toolRegistry: toolRegistry,
		tracer:       otel.Tracer("api.server"),
	}, nil
}

// Start starts the server
func (s *Server) Start() error {
	log.Printf("Starting server on %s", s.httpServer.Addr)
	return s.httpServer.ListenAndServe()
}

// Stop stops the server gracefully
func (s *Server) Stop(ctx context.Context) error {
	log.Println("Stopping server...")

	// Close LLM manager
	if err := s.llmManager.Close(); err != nil {
		log.Printf("Error closing LLM manager: %v", err)
	}

	// Shutdown HTTP server
	return s.httpServer.Shutdown(ctx)
}

// setupRouter sets up the HTTP router with all routes and middleware
func setupRouter(llmManager *llm.LLMManager, toolRegistry *tools.ToolRegistry) http.Handler {
	mux := http.NewServeMux()

	// Create handlers
	travelHandler := handlers.NewTravelHandler(llmManager, toolRegistry)

	// Create Ollama service and handlers
	ollamaService := services.NewOllamaService("http://localhost:11434")
	ollamaHandler := handlers.NewOllamaHandlers(ollamaService)

	// Apply middleware
	handler := middleware.Chain(
		mux,
		middleware.CORS(),
		middleware.Logging(),
		middleware.Recovery(),
		middleware.RequestID(),
		middleware.Timeout(30*time.Second),
	)

	// Travel routes
	mux.HandleFunc("/api/v1/travel/plan", travelHandler.PlanTrip)
	mux.HandleFunc("/api/v1/travel/flights/search", travelHandler.SearchFlights)
	mux.HandleFunc("/api/v1/travel/hotels/search", travelHandler.SearchHotels)
	mux.HandleFunc("/api/v1/travel/weather", travelHandler.GetWeather)
	mux.HandleFunc("/api/v1/travel/locations/search", travelHandler.SearchLocations)
	mux.HandleFunc("/api/v1/travel/tools", travelHandler.GetTools)

	// Ollama routes
	mux.HandleFunc("/api/v1/ollama/health", ollamaHandler.HealthCheck)
	mux.HandleFunc("/api/v1/ollama/models", ollamaHandler.ListModels)
	mux.HandleFunc("/api/v1/ollama/models/status", ollamaHandler.GetModelStatus)
	mux.HandleFunc("/api/v1/ollama/models/pull", ollamaHandler.PullModel)
	mux.HandleFunc("/api/v1/ollama/models/ensure", ollamaHandler.EnsureModel)
	mux.HandleFunc("/api/v1/ollama/generate", ollamaHandler.Generate)
	// Note: DELETE /api/v1/ollama/models/{model} is handled by DeleteModel method

	// Health check
	mux.HandleFunc("/health", travelHandler.HealthCheck)
	mux.HandleFunc("/api/v1/health", travelHandler.HealthCheck)

	// Root endpoint
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		response := map[string]interface{}{
			"service": "Exotic Travel Booking API",
			"version": "1.0.0",
			"status":  "operational",
			"endpoints": map[string]string{
				"health":           "/health",
				"plan_trip":        "/api/v1/travel/plan",
				"search_flights":   "/api/v1/travel/flights/search",
				"search_hotels":    "/api/v1/travel/hotels/search",
				"get_weather":      "/api/v1/travel/weather",
				"search_locations": "/api/v1/travel/locations/search",
				"get_tools":        "/api/v1/travel/tools",
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		// Simple JSON encoding without external dependencies
		fmt.Fprintf(w, `{
			"service": "%s",
			"version": "%s",
			"status": "%s",
			"endpoints": {
				"health": "%s",
				"plan_trip": "%s",
				"search_flights": "%s",
				"search_hotels": "%s",
				"get_weather": "%s",
				"search_locations": "%s",
				"get_tools": "%s"
			}
		}`,
			response["service"],
			response["version"],
			response["status"],
			response["endpoints"].(map[string]string)["health"],
			response["endpoints"].(map[string]string)["plan_trip"],
			response["endpoints"].(map[string]string)["search_flights"],
			response["endpoints"].(map[string]string)["search_hotels"],
			response["endpoints"].(map[string]string)["get_weather"],
			response["endpoints"].(map[string]string)["search_locations"],
			response["endpoints"].(map[string]string)["get_tools"],
		)
	})

	return handler
}

// initializeTools initializes all available tools
func initializeTools(registry *tools.ToolRegistry) error {
	// Flight search tool
	flightConfig := &tools.ToolConfig{
		Name:        "flight_search",
		Description: "Search for flights between destinations",
		Timeout:     30 * time.Second,
	}
	flightTool := tools.NewFlightSearchTool(flightConfig)
	if err := registry.RegisterTool(flightTool); err != nil {
		return fmt.Errorf("failed to register flight search tool: %w", err)
	}

	// Hotel search tool
	hotelConfig := &tools.ToolConfig{
		Name:        "hotel_search",
		Description: "Search for hotels and accommodations",
		Timeout:     30 * time.Second,
	}
	hotelTool := tools.NewHotelSearchTool(hotelConfig)
	if err := registry.RegisterTool(hotelTool); err != nil {
		return fmt.Errorf("failed to register hotel search tool: %w", err)
	}

	// Weather tool
	weatherConfig := &tools.ToolConfig{
		Name:        "weather",
		Description: "Get weather information and forecasts",
		Timeout:     15 * time.Second,
	}
	weatherTool := tools.NewWeatherTool(weatherConfig)
	if err := registry.RegisterTool(weatherTool); err != nil {
		return fmt.Errorf("failed to register weather tool: %w", err)
	}

	// Location tool
	locationConfig := &tools.ToolConfig{
		Name:        "location",
		Description: "Search for locations and places",
		Timeout:     15 * time.Second,
	}
	locationTool := tools.NewLocationTool(locationConfig)
	if err := registry.RegisterTool(locationTool); err != nil {
		return fmt.Errorf("failed to register location tool: %w", err)
	}

	log.Printf("Initialized %d tools", len(registry.ListTools()))
	return nil
}

// initializeLLMProviders initializes LLM providers
func initializeLLMProviders(manager *llm.LLMManager) error {
	// OpenAI provider (if API key is available)
	openaiConfig := &providers.LLMConfig{
		Provider: "openai",
		Model:    "gpt-4",
		APIKey:   "", // Will be loaded from environment or config
		BaseURL:  "https://api.openai.com/v1",
		Timeout:  30 * time.Second,
	}

	if err := manager.AddProvider("openai", openaiConfig); err != nil {
		log.Printf("Warning: Failed to initialize OpenAI provider: %v", err)
	}

	// Anthropic provider (if API key is available)
	anthropicConfig := &providers.LLMConfig{
		Provider: "anthropic",
		Model:    "claude-3-5-sonnet-20241022",
		APIKey:   "", // Will be loaded from environment or config
		BaseURL:  "https://api.anthropic.com/v1",
		Timeout:  30 * time.Second,
	}

	if err := manager.AddProvider("anthropic", anthropicConfig); err != nil {
		log.Printf("Warning: Failed to initialize Anthropic provider: %v", err)
	}

	// Ollama provider (for local inference)
	ollamaConfig := &providers.LLMConfig{
		Provider: "ollama",
		Model:    "llama3.2", // Default model, can be overridden
		BaseURL:  "http://localhost:11434",
		Timeout:  60 * time.Second,
	}

	if err := manager.AddProvider("ollama", ollamaConfig); err != nil {
		log.Printf("Warning: Failed to initialize Ollama provider: %v", err)
	}

	// Local provider (for development/testing)
	localConfig := &providers.LLMConfig{
		Provider: "local",
		Model:    "mock-model",
		Timeout:  10 * time.Second,
	}

	if err := manager.AddProvider("local", localConfig); err != nil {
		log.Printf("Warning: Failed to initialize local provider: %v", err)
	}

	// Set default provider
	providers := manager.ListProviders()
	if len(providers) > 0 {
		if err := manager.SetDefaultProvider(providers[0]); err != nil {
			return fmt.Errorf("failed to set default provider: %w", err)
		}
		log.Printf("Set default LLM provider to: %s", providers[0])
	} else {
		log.Println("Warning: No LLM providers available")
	}

	log.Printf("Initialized %d LLM providers", len(providers))
	return nil
}

// GetDefaultConfig returns default server configuration
func GetDefaultConfig() *Config {
	return &Config{
		Port:         "8080",
		Host:         "0.0.0.0",
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
}
