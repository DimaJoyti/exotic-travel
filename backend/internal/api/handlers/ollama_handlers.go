package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/exotic-travel-booking/backend/internal/services"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// OllamaHandlers handles Ollama-related HTTP requests
type OllamaHandlers struct {
	ollamaService *services.OllamaService
	tracer        trace.Tracer
}

// NewOllamaHandlers creates new Ollama handlers
func NewOllamaHandlers(ollamaService *services.OllamaService) *OllamaHandlers {
	return &OllamaHandlers{
		ollamaService: ollamaService,
		tracer:        otel.Tracer("handlers.ollama"),
	}
}

// ListModels handles GET /api/ollama/models
func (h *OllamaHandlers) ListModels(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "ollama_handlers.list_models")
	defer span.End()

	models, err := h.ollamaService.ListModels(ctx)
	if err != nil {
		span.RecordError(err)
		http.Error(w, fmt.Sprintf("Failed to list models: %v", err), http.StatusInternalServerError)
		return
	}

	span.SetAttributes(attribute.Int("models.count", len(models)))

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"models": models,
		"count":  len(models),
	}); err != nil {
		span.RecordError(err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// GetModelStatus handles GET /api/ollama/models/status
func (h *OllamaHandlers) GetModelStatus(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "ollama_handlers.get_model_status")
	defer span.End()

	status, err := h.ollamaService.GetModelStatus(ctx)
	if err != nil {
		span.RecordError(err)
		http.Error(w, fmt.Sprintf("Failed to get model status: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"models":      status,
		"recommended": h.ollamaService.GetRecommendedModels(),
	}); err != nil {
		span.RecordError(err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// PullModelRequest represents a request to pull a model
type PullModelRequest struct {
	ModelName string `json:"model_name"`
}

// PullModel handles POST /api/ollama/models/pull
func (h *OllamaHandlers) PullModel(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "ollama_handlers.pull_model")
	defer span.End()

	var req PullModelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		span.RecordError(err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.ModelName == "" {
		http.Error(w, "model_name is required", http.StatusBadRequest)
		return
	}

	span.SetAttributes(attribute.String("model.name", req.ModelName))

	// Start pulling model in background
	go func() {
		if err := h.ollamaService.PullModel(ctx, req.ModelName); err != nil {
			// Log error but don't fail the request since it's async
			fmt.Printf("Failed to pull model %s: %v\n", req.ModelName, err)
		}
	}()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": fmt.Sprintf("Started pulling model: %s", req.ModelName),
		"model":   req.ModelName,
		"status":  "pulling",
	})
}

// DeleteModel handles DELETE /api/ollama/models/{model}
func (h *OllamaHandlers) DeleteModel(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "ollama_handlers.delete_model")
	defer span.End()

	// Extract model name from URL path
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 4 {
		http.Error(w, "Model name is required", http.StatusBadRequest)
		return
	}
	modelName := pathParts[3] // /api/ollama/models/{model}

	span.SetAttributes(attribute.String("model.name", modelName))

	if err := h.ollamaService.DeleteModel(ctx, modelName); err != nil {
		span.RecordError(err)
		http.Error(w, fmt.Sprintf("Failed to delete model: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": fmt.Sprintf("Successfully deleted model: %s", modelName),
		"model":   modelName,
	})
}

// HealthCheck handles GET /api/ollama/health
func (h *OllamaHandlers) HealthCheck(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "ollama_handlers.health_check")
	defer span.End()

	if err := h.ollamaService.CheckHealth(ctx); err != nil {
		span.RecordError(err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "unhealthy",
			"error":  err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "healthy",
	})
}

// GenerateRequest represents a request to generate text
type GenerateRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream,omitempty"`
}

// Generate handles POST /api/ollama/generate
func (h *OllamaHandlers) Generate(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "ollama_handlers.generate")
	defer span.End()

	var req GenerateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		span.RecordError(err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Model == "" {
		http.Error(w, "model is required", http.StatusBadRequest)
		return
	}

	if req.Prompt == "" {
		http.Error(w, "prompt is required", http.StatusBadRequest)
		return
	}

	span.SetAttributes(
		attribute.String("model.name", req.Model),
		attribute.Bool("request.stream", req.Stream),
	)

	if req.Stream {
		h.handleStreamGenerate(w, r, req)
		return
	}

	// Non-streaming response
	response, err := h.ollamaService.GenerateResponse(ctx, req.Model, req.Prompt)
	if err != nil {
		span.RecordError(err)
		http.Error(w, fmt.Sprintf("Failed to generate response: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"model":    req.Model,
		"response": response,
		"done":     true,
	})
}

// handleStreamGenerate handles streaming text generation
func (h *OllamaHandlers) handleStreamGenerate(w http.ResponseWriter, r *http.Request, req GenerateRequest) {
	ctx, span := h.tracer.Start(r.Context(), "ollama_handlers.stream_generate")
	defer span.End()

	// Set headers for streaming
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	stream, err := h.ollamaService.StreamResponse(ctx, req.Model, req.Prompt)
	if err != nil {
		span.RecordError(err)
		http.Error(w, fmt.Sprintf("Failed to start streaming: %v", err), http.StatusInternalServerError)
		return
	}

	for {
		select {
		case chunk, ok := <-stream:
			if !ok {
				// Stream closed
				fmt.Fprintf(w, "data: {\"done\": true}\n\n")
				flusher.Flush()
				return
			}

			// Send chunk as SSE
			data := map[string]interface{}{
				"model":    req.Model,
				"response": chunk,
				"done":     false,
			}

			jsonData, _ := json.Marshal(data)
			fmt.Fprintf(w, "data: %s\n\n", jsonData)
			flusher.Flush()

		case <-ctx.Done():
			return
		}
	}
}

// EnsureModel handles POST /api/ollama/models/ensure
func (h *OllamaHandlers) EnsureModel(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "ollama_handlers.ensure_model")
	defer span.End()

	var req PullModelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		span.RecordError(err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.ModelName == "" {
		http.Error(w, "model_name is required", http.StatusBadRequest)
		return
	}

	span.SetAttributes(attribute.String("model.name", req.ModelName))

	if err := h.ollamaService.EnsureModel(ctx, req.ModelName); err != nil {
		span.RecordError(err)
		http.Error(w, fmt.Sprintf("Failed to ensure model: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": fmt.Sprintf("Model %s is ready", req.ModelName),
		"model":   req.ModelName,
		"status":  "ready",
	})
}
