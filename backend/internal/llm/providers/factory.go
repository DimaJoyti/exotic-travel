package providers

import (
	"fmt"
)

// ProviderFactory creates LLM providers
type ProviderFactory struct{}

// NewProviderFactory creates a new provider factory
func NewProviderFactory() *ProviderFactory {
	return &ProviderFactory{}
}

// CreateProvider creates a provider based on configuration
func (f *ProviderFactory) CreateProvider(config *LLMConfig) (LLMProvider, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}
	if config.Provider == "" {
		return nil, fmt.Errorf("provider type cannot be empty")
	}

	switch config.Provider {
	case "openai":
		return NewOpenAIProvider(config)
	case "anthropic":
		return NewAnthropicProvider(config)
	case "local":
		return NewLocalProvider(config)
	case "ollama":
		return NewOllamaProvider(config)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", config.Provider)
	}
}
