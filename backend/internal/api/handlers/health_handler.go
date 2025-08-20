package handlers

import (
	"context"
	"time"

	"github.com/exotic-travel-booking/backend/internal/llm/providers"
	"github.com/exotic-travel-booking/backend/internal/services"
	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// HealthHandler handles health and monitoring endpoints
type HealthHandler struct {
	ollamaService *services.OllamaService
	llmProvider   providers.LLMProvider
	tracer        trace.Tracer
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(ollamaService *services.OllamaService, llmProvider providers.LLMProvider) *HealthHandler {
	return &HealthHandler{
		ollamaService: ollamaService,
		llmProvider:   llmProvider,
		tracer:        otel.Tracer("api.health_handler"),
	}
}

// HealthResponse represents a health check response
type HealthResponse struct {
	Status    string                 `json:"status"`
	Timestamp time.Time              `json:"timestamp"`
	Services  map[string]interface{} `json:"services"`
	Version   string                 `json:"version"`
	Uptime    time.Duration          `json:"uptime"`
}

// ServiceStatus represents the status of a service
type ServiceStatus struct {
	Status       string                 `json:"status"`
	LastChecked  time.Time              `json:"last_checked"`
	ResponseTime time.Duration          `json:"response_time,omitempty"`
	Error        string                 `json:"error,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

var startTime = time.Now()

// Health performs a comprehensive health check
func (h *HealthHandler) Health(c *fiber.Ctx) error {
	ctx, span := h.tracer.Start(c.Context(), "health_handler.health")
	defer span.End()

	services := make(map[string]interface{})
	overallStatus := "healthy"

	// Check OLAMA service
	ollamaStatus := h.checkOllamaHealth(ctx)
	services["ollama"] = ollamaStatus
	if ollamaStatus.Status != "healthy" {
		overallStatus = "degraded"
	}

	// Check LLM provider
	llmStatus := h.checkLLMProviderHealth(ctx)
	services["llm_provider"] = llmStatus
	if llmStatus.Status != "healthy" {
		overallStatus = "degraded"
	}

	response := &HealthResponse{
		Status:    overallStatus,
		Timestamp: time.Now(),
		Services:  services,
		Version:   "1.0.0",
		Uptime:    time.Since(startTime),
	}

	span.SetAttributes(
		attribute.String("health.status", overallStatus),
		attribute.Int("health.services_count", len(services)),
	)

	statusCode := fiber.StatusOK
	if overallStatus == "degraded" {
		statusCode = fiber.StatusServiceUnavailable
	}

	return c.Status(statusCode).JSON(response)
}

// checkOllamaHealth checks the health of the OLAMA service
func (h *HealthHandler) checkOllamaHealth(ctx context.Context) *ServiceStatus {
	start := time.Now()

	if h.ollamaService == nil {
		return &ServiceStatus{
			Status:      "unavailable",
			LastChecked: time.Now(),
			Error:       "OLAMA service not configured",
		}
	}

	err := h.ollamaService.CheckHealth(ctx)
	responseTime := time.Since(start)

	if err != nil {
		return &ServiceStatus{
			Status:       "unhealthy",
			LastChecked:  time.Now(),
			ResponseTime: responseTime,
			Error:        err.Error(),
		}
	}

	// Get available models
	models, err := h.ollamaService.ListModels(ctx)
	metadata := map[string]interface{}{
		"models_available": len(models),
	}
	if err == nil && len(models) > 0 {
		metadata["models"] = models
	}

	return &ServiceStatus{
		Status:       "healthy",
		LastChecked:  time.Now(),
		ResponseTime: responseTime,
		Metadata:     metadata,
	}
}

// checkLLMProviderHealth checks the health of the LLM provider
func (h *HealthHandler) checkLLMProviderHealth(ctx context.Context) *ServiceStatus {
	start := time.Now()

	if h.llmProvider == nil {
		return &ServiceStatus{
			Status:      "unavailable",
			LastChecked: time.Now(),
			Error:       "LLM provider not configured",
		}
	}

	// Try to get models as a health check
	_, err := h.llmProvider.GetModels(ctx)
	responseTime := time.Since(start)

	if err != nil {
		return &ServiceStatus{
			Status:       "unhealthy",
			LastChecked:  time.Now(),
			ResponseTime: responseTime,
			Error:        err.Error(),
		}
	}

	metadata := map[string]interface{}{
		"provider_name": h.llmProvider.GetName(),
	}

	// Try to get available models
	if models, err := h.llmProvider.GetModels(ctx); err == nil {
		metadata["models_available"] = len(models)
		metadata["models"] = models
	}

	return &ServiceStatus{
		Status:       "healthy",
		LastChecked:  time.Now(),
		ResponseTime: responseTime,
		Metadata:     metadata,
	}
}

// ReadinessResponse represents a readiness check response
type ReadinessResponse struct {
	Ready     bool                   `json:"ready"`
	Timestamp time.Time              `json:"timestamp"`
	Checks    map[string]interface{} `json:"checks"`
}

// Readiness performs a readiness check
func (h *HealthHandler) Readiness(c *fiber.Ctx) error {
	ctx, span := h.tracer.Start(c.Context(), "health_handler.readiness")
	defer span.End()

	checks := make(map[string]interface{})
	ready := true

	// Check if OLAMA is ready
	ollamaReady := h.isOllamaReady(ctx)
	checks["ollama"] = map[string]interface{}{
		"ready": ollamaReady,
		"name":  "OLAMA Service",
	}
	if !ollamaReady {
		ready = false
	}

	// Check if LLM provider is ready
	llmReady := h.isLLMProviderReady(ctx)
	checks["llm_provider"] = map[string]interface{}{
		"ready": llmReady,
		"name":  "LLM Provider",
	}
	if !llmReady {
		ready = false
	}

	response := &ReadinessResponse{
		Ready:     ready,
		Timestamp: time.Now(),
		Checks:    checks,
	}

	span.SetAttributes(
		attribute.Bool("readiness.ready", ready),
		attribute.Int("readiness.checks_count", len(checks)),
	)

	statusCode := fiber.StatusOK
	if !ready {
		statusCode = fiber.StatusServiceUnavailable
	}

	return c.Status(statusCode).JSON(response)
}

// isOllamaReady checks if OLAMA is ready to serve requests
func (h *HealthHandler) isOllamaReady(ctx context.Context) bool {
	if h.ollamaService == nil {
		return false
	}

	// Quick health check with timeout
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return h.ollamaService.CheckHealth(ctx) == nil
}

// isLLMProviderReady checks if LLM provider is ready
func (h *HealthHandler) isLLMProviderReady(ctx context.Context) bool {
	if h.llmProvider == nil {
		return false
	}

	// Try to get models as a readiness check
	_, err := h.llmProvider.GetModels(ctx)
	return err == nil
}

// LivenessResponse represents a liveness check response
type LivenessResponse struct {
	Alive     bool          `json:"alive"`
	Timestamp time.Time     `json:"timestamp"`
	Uptime    time.Duration `json:"uptime"`
}

// Liveness performs a liveness check
func (h *HealthHandler) Liveness(c *fiber.Ctx) error {
	_, span := h.tracer.Start(c.Context(), "health_handler.liveness")
	defer span.End()

	response := &LivenessResponse{
		Alive:     true,
		Timestamp: time.Now(),
		Uptime:    time.Since(startTime),
	}

	span.SetAttributes(
		attribute.Bool("liveness.alive", true),
		attribute.Int64("liveness.uptime_seconds", int64(response.Uptime.Seconds())),
	)

	return c.JSON(response)
}

// MetricsResponse represents metrics response
type MetricsResponse struct {
	Timestamp time.Time              `json:"timestamp"`
	Uptime    time.Duration          `json:"uptime"`
	Metrics   map[string]interface{} `json:"metrics"`
}

// Metrics returns application metrics
func (h *HealthHandler) Metrics(c *fiber.Ctx) error {
	_, span := h.tracer.Start(c.Context(), "health_handler.metrics")
	defer span.End()

	metrics := map[string]interface{}{
		"uptime_seconds": time.Since(startTime).Seconds(),
		"start_time":     startTime,
		"current_time":   time.Now(),
		"version":        "1.0.0",
		"go_version":     "1.21+",
	}

	// Add service-specific metrics
	if h.ollamaService != nil {
		metrics["ollama_configured"] = true
	} else {
		metrics["ollama_configured"] = false
	}

	if h.llmProvider != nil {
		metrics["llm_provider_configured"] = true
		metrics["llm_provider_name"] = h.llmProvider.GetName()
	} else {
		metrics["llm_provider_configured"] = false
	}

	response := &MetricsResponse{
		Timestamp: time.Now(),
		Uptime:    time.Since(startTime),
		Metrics:   metrics,
	}

	span.SetAttributes(
		attribute.Int("metrics.count", len(metrics)),
	)

	return c.JSON(response)
}
