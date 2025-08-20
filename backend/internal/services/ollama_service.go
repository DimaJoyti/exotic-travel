package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/exotic-travel-booking/backend/internal/llm/providers"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// OllamaService manages Ollama models and operations
type OllamaService struct {
	client *providers.OllamaClient
	tracer trace.Tracer
}

// NewOllamaService creates a new Ollama service
func NewOllamaService(baseURL string) *OllamaService {
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}

	client := providers.NewOllamaClient(baseURL, 60*time.Second)
	tracer := otel.Tracer("service.ollama")

	return &OllamaService{
		client: client,
		tracer: tracer,
	}
}

// ModelInfo represents information about an Ollama model
type ModelInfo struct {
	Name         string    `json:"name"`
	Size         int64     `json:"size"`
	ModifiedAt   time.Time `json:"modified_at"`
	Family       string    `json:"family"`
	ParameterSize string   `json:"parameter_size"`
}

// ListModels returns all available models
func (s *OllamaService) ListModels(ctx context.Context) ([]ModelInfo, error) {
	ctx, span := s.tracer.Start(ctx, "ollama_service.list_models")
	defer span.End()

	resp, err := s.client.ListModels(ctx)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to list models: %w", err)
	}

	models := make([]ModelInfo, len(resp.Models))
	for i, model := range resp.Models {
		models[i] = ModelInfo{
			Name:         model.Name,
			Size:         model.Size,
			ModifiedAt:   model.ModifiedAt,
			Family:       model.Details.Family,
			ParameterSize: model.Details.ParameterSize,
		}
	}

	span.SetAttributes(attribute.Int("models.count", len(models)))
	return models, nil
}

// PullModel downloads a model from Ollama registry
func (s *OllamaService) PullModel(ctx context.Context, modelName string) error {
	ctx, span := s.tracer.Start(ctx, "ollama_service.pull_model")
	defer span.End()

	span.SetAttributes(attribute.String("model.name", modelName))

	log.Printf("Pulling model: %s", modelName)
	err := s.client.PullModel(ctx, modelName)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to pull model %s: %w", modelName, err)
	}

	log.Printf("Successfully pulled model: %s", modelName)
	return nil
}

// DeleteModel removes a model from Ollama
func (s *OllamaService) DeleteModel(ctx context.Context, modelName string) error {
	ctx, span := s.tracer.Start(ctx, "ollama_service.delete_model")
	defer span.End()

	span.SetAttributes(attribute.String("model.name", modelName))

	log.Printf("Deleting model: %s", modelName)
	err := s.client.DeleteModel(ctx, modelName)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to delete model %s: %w", modelName, err)
	}

	log.Printf("Successfully deleted model: %s", modelName)
	return nil
}

// CheckHealth verifies Ollama is running and accessible
func (s *OllamaService) CheckHealth(ctx context.Context) error {
	ctx, span := s.tracer.Start(ctx, "ollama_service.check_health")
	defer span.End()

	err := s.client.Health(ctx)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("ollama health check failed: %w", err)
	}

	return nil
}

// GenerateResponse generates a response using a specific model
func (s *OllamaService) GenerateResponse(ctx context.Context, model, prompt string) (string, error) {
	ctx, span := s.tracer.Start(ctx, "ollama_service.generate_response")
	defer span.End()

	span.SetAttributes(
		attribute.String("model.name", model),
		attribute.String("prompt.length", fmt.Sprintf("%d", len(prompt))),
	)

	req := &providers.OllamaChatRequest{
		Model: model,
		Messages: []providers.OllamaChatMessage{
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Stream: false,
	}

	resp, err := s.client.Chat(ctx, req)
	if err != nil {
		span.RecordError(err)
		return "", fmt.Errorf("failed to generate response: %w", err)
	}

	span.SetAttributes(attribute.String("response.length", fmt.Sprintf("%d", len(resp.Message.Content))))
	return resp.Message.Content, nil
}

// StreamResponse generates a streaming response using a specific model
func (s *OllamaService) StreamResponse(ctx context.Context, model, prompt string) (<-chan string, error) {
	ctx, span := s.tracer.Start(ctx, "ollama_service.stream_response")
	defer span.End()

	span.SetAttributes(
		attribute.String("model.name", model),
		attribute.String("prompt.length", fmt.Sprintf("%d", len(prompt))),
	)

	req := &providers.OllamaChatRequest{
		Model: model,
		Messages: []providers.OllamaChatMessage{
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Stream: true,
	}

	ollamaStream, err := s.client.ChatStream(ctx, req)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to start streaming response: %w", err)
	}

	// Convert Ollama stream to string stream
	responseStream := make(chan string)
	go func() {
		defer close(responseStream)
		for chunk := range ollamaStream {
			if chunk.Message.Content != "" {
				select {
				case responseStream <- chunk.Message.Content:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return responseStream, nil
}

// EnsureModel ensures a model is available, pulling it if necessary
func (s *OllamaService) EnsureModel(ctx context.Context, modelName string) error {
	ctx, span := s.tracer.Start(ctx, "ollama_service.ensure_model")
	defer span.End()

	span.SetAttributes(attribute.String("model.name", modelName))

	// Check if model exists
	models, err := s.ListModels(ctx)
	if err != nil {
		return fmt.Errorf("failed to list models: %w", err)
	}

	for _, model := range models {
		if model.Name == modelName {
			log.Printf("Model %s already available", modelName)
			return nil
		}
	}

	// Model doesn't exist, pull it
	log.Printf("Model %s not found, pulling...", modelName)
	return s.PullModel(ctx, modelName)
}

// GetRecommendedModels returns a list of recommended models for travel use cases
func (s *OllamaService) GetRecommendedModels() []string {
	return []string{
		"llama3.2",      // Latest Llama model, good for general tasks
		"llama3.2:1b",   // Smaller, faster model
		"llama3.2:3b",   // Medium-sized model
		"mistral",       // Good alternative to Llama
		"codellama",     // For code-related tasks
		"phi3",          // Microsoft's efficient model
		"gemma2",        // Google's Gemma model
	}
}

// ModelStatus represents the status of a model
type ModelStatus struct {
	Name      string `json:"name"`
	Available bool   `json:"available"`
	Size      int64  `json:"size,omitempty"`
}

// GetModelStatus returns the status of recommended models
func (s *OllamaService) GetModelStatus(ctx context.Context) ([]ModelStatus, error) {
	ctx, span := s.tracer.Start(ctx, "ollama_service.get_model_status")
	defer span.End()

	availableModels, err := s.ListModels(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list models: %w", err)
	}

	// Create a map for quick lookup
	modelMap := make(map[string]ModelInfo)
	for _, model := range availableModels {
		modelMap[model.Name] = model
	}

	recommended := s.GetRecommendedModels()
	status := make([]ModelStatus, len(recommended))

	for i, modelName := range recommended {
		if model, exists := modelMap[modelName]; exists {
			status[i] = ModelStatus{
				Name:      modelName,
				Available: true,
				Size:      model.Size,
			}
		} else {
			status[i] = ModelStatus{
				Name:      modelName,
				Available: false,
			}
		}
	}

	return status, nil
}
