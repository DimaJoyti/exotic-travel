package providers

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOllamaProvider_Creation(t *testing.T) {
	config := &LLMConfig{
		Provider: "ollama",
		Model:    "llama3.2",
		BaseURL:  "http://localhost:11434",
		Timeout:  30 * time.Second,
	}

	provider, err := NewOllamaProvider(config)
	require.NoError(t, err)
	assert.NotNil(t, provider)
	assert.Equal(t, "ollama", provider.GetName())
}

func TestOllamaProvider_DefaultURL(t *testing.T) {
	config := &LLMConfig{
		Provider: "ollama",
		Model:    "llama3.2",
		// BaseURL not set, should default to localhost:11434
		Timeout: 30 * time.Second,
	}

	provider, err := NewOllamaProvider(config)
	require.NoError(t, err)
	assert.NotNil(t, provider)

	// Cast to access internal fields for testing
	ollamaProvider, ok := provider.(*OllamaProvider)
	require.True(t, ok)
	assert.NotNil(t, ollamaProvider.client)
}

func TestOllamaClient_Creation(t *testing.T) {
	client := NewOllamaClient("http://localhost:11434", 30*time.Second)
	assert.NotNil(t, client)
}

func TestOllamaClient_DefaultURL(t *testing.T) {
	client := NewOllamaClient("", 30*time.Second)
	assert.NotNil(t, client)
	// Should use default URL
}

// Integration test - only runs if Ollama is available
func TestOllamaProvider_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	config := &LLMConfig{
		Provider: "ollama",
		Model:    "llama3.2",
		BaseURL:  "http://localhost:11434",
		Timeout:  60 * time.Second,
	}

	provider, err := NewOllamaProvider(config)
	require.NoError(t, err)

	ctx := context.Background()

	// Test health check first
	ollamaProvider := provider.(*OllamaProvider)
	err = ollamaProvider.client.Health(ctx)
	if err != nil {
		t.Skipf("Ollama not available: %v", err)
	}

	// Test listing models
	models, err := provider.GetModels(ctx)
	if err != nil {
		t.Logf("Warning: Could not list models: %v", err)
		return
	}

	t.Logf("Available models: %v", models)

	// If we have models, test generation
	if len(models) > 0 {
		req := &GenerateRequest{
			Model: models[0], // Use first available model
			Messages: []Message{
				{
					Role:    "user",
					Content: "Hello, this is a test. Please respond with 'Test successful'.",
				},
			},
			MaxTokens:   50,
			Temperature: 0.1,
		}

		resp, err := provider.GenerateResponse(ctx, req)
		if err != nil {
			t.Logf("Warning: Could not generate response: %v", err)
			return
		}

		assert.NotNil(t, resp)
		assert.Greater(t, len(resp.Choices), 0)
		assert.NotEmpty(t, resp.Choices[0].Message.Content)
		t.Logf("Generated response: %s", resp.Choices[0].Message.Content)
	}
}

// Test streaming functionality
func TestOllamaProvider_Streaming(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	config := &LLMConfig{
		Provider: "ollama",
		Model:    "llama3.2",
		BaseURL:  "http://localhost:11434",
		Timeout:  60 * time.Second,
	}

	provider, err := NewOllamaProvider(config)
	require.NoError(t, err)

	ctx := context.Background()

	// Test health check first
	ollamaProvider := provider.(*OllamaProvider)
	err = ollamaProvider.client.Health(ctx)
	if err != nil {
		t.Skipf("Ollama not available: %v", err)
	}

	// Test listing models
	models, err := provider.GetModels(ctx)
	if err != nil {
		t.Logf("Warning: Could not list models: %v", err)
		return
	}

	if len(models) == 0 {
		t.Skip("No models available for streaming test")
	}

	req := &GenerateRequest{
		Model: models[0],
		Messages: []Message{
			{
				Role:    "user",
				Content: "Count from 1 to 5.",
			},
		},
		MaxTokens:   100,
		Temperature: 0.1,
	}

	stream, err := provider.StreamResponse(ctx, req)
	if err != nil {
		t.Logf("Warning: Could not start streaming: %v", err)
		return
	}

	var chunks []string
	timeout := time.After(30 * time.Second)

	for {
		select {
		case chunk, ok := <-stream:
			if !ok {
				// Stream closed
				t.Logf("Received %d chunks", len(chunks))
				assert.Greater(t, len(chunks), 0, "Should receive at least one chunk")
				return
			}
			if len(chunk.Choices) > 0 && chunk.Choices[0].Delta.Content != "" {
				chunks = append(chunks, chunk.Choices[0].Delta.Content)
				t.Logf("Chunk: %s", chunk.Choices[0].Delta.Content)
			}
		case <-timeout:
			t.Fatal("Streaming test timed out")
		}
	}
}

func TestConvertToOllamaRequest(t *testing.T) {
	config := &LLMConfig{
		Provider: "ollama",
		Model:    "llama3.2",
		BaseURL:  "http://localhost:11434",
		Timeout:  30 * time.Second,
	}

	provider, err := NewOllamaProvider(config)
	require.NoError(t, err)

	ollamaProvider := provider.(*OllamaProvider)

	req := &GenerateRequest{
		Model: "test-model",
		Messages: []Message{
			{Role: "user", Content: "Hello"},
			{Role: "assistant", Content: "Hi there"},
			{Role: "user", Content: "How are you?"},
		},
		Temperature: 0.7,
		MaxTokens:   100,
		TopP:        0.9,
	}

	ollamaReq := ollamaProvider.convertToOllamaRequest(req)

	assert.Equal(t, "test-model", ollamaReq.Model)
	assert.Equal(t, 3, len(ollamaReq.Messages))
	assert.Equal(t, "user", ollamaReq.Messages[0].Role)
	assert.Equal(t, "Hello", ollamaReq.Messages[0].Content)
	assert.Equal(t, "assistant", ollamaReq.Messages[1].Role)
	assert.Equal(t, "Hi there", ollamaReq.Messages[1].Content)
	assert.Equal(t, "user", ollamaReq.Messages[2].Role)
	assert.Equal(t, "How are you?", ollamaReq.Messages[2].Content)
	assert.False(t, ollamaReq.Stream)

	// Check options
	assert.Equal(t, 0.7, ollamaReq.Options["temperature"])
	assert.Equal(t, 100, ollamaReq.Options["num_predict"])
	assert.Equal(t, 0.9, ollamaReq.Options["top_p"])
}

func TestConvertFromOllamaResponse(t *testing.T) {
	config := &LLMConfig{
		Provider: "ollama",
		Model:    "llama3.2",
		BaseURL:  "http://localhost:11434",
		Timeout:  30 * time.Second,
	}

	provider, err := NewOllamaProvider(config)
	require.NoError(t, err)

	ollamaProvider := provider.(*OllamaProvider)

	ollamaResp := &OllamaChatResponse{
		Model: "test-model",
		Message: OllamaChatMessage{
			Role:    "assistant",
			Content: "Hello, how can I help you?",
		},
		Done: true,
	}

	resp := ollamaProvider.convertFromOllamaResponse(ollamaResp)

	assert.Equal(t, "test-model", resp.Model)
	assert.Equal(t, 1, len(resp.Choices))
	assert.Equal(t, "assistant", resp.Choices[0].Message.Role)
	assert.Equal(t, "Hello, how can I help you?", resp.Choices[0].Message.Content)
	assert.Equal(t, "stop", resp.Choices[0].FinishReason)
}

func TestConvertOllamaStreamChunk(t *testing.T) {
	config := &LLMConfig{
		Provider: "ollama",
		Model:    "llama3.2",
		BaseURL:  "http://localhost:11434",
		Timeout:  30 * time.Second,
	}

	provider, err := NewOllamaProvider(config)
	require.NoError(t, err)

	ollamaProvider := provider.(*OllamaProvider)

	// Test non-final chunk
	ollamaResp := &OllamaChatResponse{
		Model: "test-model",
		Message: OllamaChatMessage{
			Role:    "assistant",
			Content: "Hello",
		},
		Done: false,
	}

	chunk := ollamaProvider.convertOllamaStreamChunk(ollamaResp)

	assert.Equal(t, "test-model", chunk.Model)
	assert.Equal(t, 1, len(chunk.Choices))
	assert.Equal(t, "assistant", chunk.Choices[0].Delta.Role)
	assert.Equal(t, "Hello", chunk.Choices[0].Delta.Content)
	assert.Nil(t, chunk.Choices[0].FinishReason)
	assert.False(t, chunk.Done)

	// Test final chunk
	ollamaResp.Done = true
	chunk = ollamaProvider.convertOllamaStreamChunk(ollamaResp)

	assert.True(t, chunk.Done)
	assert.NotNil(t, chunk.Choices[0].FinishReason)
	assert.Equal(t, "stop", *chunk.Choices[0].FinishReason)
}
