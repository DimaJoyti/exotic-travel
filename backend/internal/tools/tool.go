package tools

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Tool represents an external tool that can be called
type Tool interface {
	// Execute executes the tool with the given input
	Execute(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error)

	// GetName returns the tool name
	GetName() string

	// GetDescription returns the tool description
	GetDescription() string

	// GetSchema returns the JSON schema for the tool's input parameters
	GetSchema() map[string]interface{}

	// Validate validates the tool configuration
	Validate() error
}

// ToolConfig represents configuration for a tool
type ToolConfig struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	APIKey      string                 `json:"api_key,omitempty"`
	BaseURL     string                 `json:"base_url,omitempty"`
	Timeout     time.Duration          `json:"timeout,omitempty"`
	RateLimit   *RateLimitConfig       `json:"rate_limit,omitempty"`
	Cache       *CacheConfig           `json:"cache,omitempty"`
	Retry       *RetryConfig           `json:"retry,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// RateLimitConfig represents rate limiting configuration
type RateLimitConfig struct {
	RequestsPerSecond int           `json:"requests_per_second"`
	BurstSize         int           `json:"burst_size"`
	WindowSize        time.Duration `json:"window_size"`
}

// CacheConfig represents caching configuration
type CacheConfig struct {
	Enabled bool          `json:"enabled"`
	TTL     time.Duration `json:"ttl"`
	MaxSize int           `json:"max_size"`
}

// RetryConfig represents retry configuration
type RetryConfig struct {
	MaxRetries    int           `json:"max_retries"`
	InitialDelay  time.Duration `json:"initial_delay"`
	MaxDelay      time.Duration `json:"max_delay"`
	BackoffFactor float64       `json:"backoff_factor"`
}

// BaseTool provides common functionality for all tools
type BaseTool struct {
	config *ToolConfig
	tracer trace.Tracer
}

// NewBaseTool creates a new base tool
func NewBaseTool(config *ToolConfig) *BaseTool {
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	return &BaseTool{
		config: config,
		tracer: otel.Tracer("tools.base"),
	}
}

// GetName returns the tool name
func (t *BaseTool) GetName() string {
	return t.config.Name
}

// GetDescription returns the tool description
func (t *BaseTool) GetDescription() string {
	return t.config.Description
}

// GetConfig returns the tool configuration
func (t *BaseTool) GetConfig() *ToolConfig {
	return t.config
}

// Validate validates the tool configuration
func (t *BaseTool) Validate() error {
	if t.config.Name == "" {
		return fmt.Errorf("tool name cannot be empty")
	}

	if t.config.Description == "" {
		return fmt.Errorf("tool description cannot be empty")
	}

	return nil
}

// WithRetry executes a function with retry logic
func (t *BaseTool) WithRetry(ctx context.Context, operation func() error) error {
	if t.config.Retry == nil {
		return operation()
	}

	var lastErr error
	delay := t.config.Retry.InitialDelay

	for attempt := 0; attempt <= t.config.Retry.MaxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
				// Continue with retry
			}

			// Calculate next delay with exponential backoff
			delay = time.Duration(float64(delay) * t.config.Retry.BackoffFactor)
			if delay > t.config.Retry.MaxDelay {
				delay = t.config.Retry.MaxDelay
			}
		}

		err := operation()
		if err == nil {
			return nil
		}

		lastErr = err

		// Check if error is retryable (could be enhanced with specific error types)
		if !t.isRetryableError(err) {
			return err
		}
	}

	return fmt.Errorf("operation failed after %d retries: %w", t.config.Retry.MaxRetries, lastErr)
}

// isRetryableError checks if an error should trigger a retry
func (t *BaseTool) isRetryableError(err error) bool {
	// Simple implementation - in production, check for specific error types
	// like network timeouts, 5xx HTTP errors, etc.
	return true
}

// ToolError represents an error from tool execution
type ToolError struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Tool    string                 `json:"tool"`
	Details map[string]interface{} `json:"details,omitempty"`
}

func (e *ToolError) Error() string {
	return fmt.Sprintf("tool %s error [%s]: %s", e.Tool, e.Code, e.Message)
}

// NewToolError creates a new tool error
func NewToolError(code, message, tool string, details map[string]interface{}) *ToolError {
	return &ToolError{
		Code:    code,
		Message: message,
		Tool:    tool,
		Details: details,
	}
}

// ToolRegistry manages available tools
type ToolRegistry struct {
	tools  map[string]Tool
	tracer trace.Tracer
}

// NewToolRegistry creates a new tool registry
func NewToolRegistry() *ToolRegistry {
	return &ToolRegistry{
		tools:  make(map[string]Tool),
		tracer: otel.Tracer("tools.registry"),
	}
}

// RegisterTool registers a tool
func (r *ToolRegistry) RegisterTool(tool Tool) error {
	if err := tool.Validate(); err != nil {
		return fmt.Errorf("invalid tool: %w", err)
	}

	name := tool.GetName()
	if _, exists := r.tools[name]; exists {
		return fmt.Errorf("tool already registered: %s", name)
	}

	r.tools[name] = tool
	return nil
}

// GetTool retrieves a tool by name
func (r *ToolRegistry) GetTool(name string) (Tool, error) {
	tool, exists := r.tools[name]
	if !exists {
		return nil, fmt.Errorf("tool not found: %s", name)
	}

	return tool, nil
}

// ListTools returns all registered tool names
func (r *ToolRegistry) ListTools() []string {
	names := make([]string, 0, len(r.tools))
	for name := range r.tools {
		names = append(names, name)
	}
	return names
}

// ExecuteTool executes a tool by name
func (r *ToolRegistry) ExecuteTool(ctx context.Context, name string, input map[string]interface{}) (map[string]interface{}, error) {
	ctx, span := r.tracer.Start(ctx, "tool_registry.execute_tool")
	defer span.End()

	span.SetAttributes(
		attribute.String("tool.name", name),
	)

	tool, err := r.GetTool(name)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	return tool.Execute(ctx, input)
}

// GetToolSchema returns the schema for a tool
func (r *ToolRegistry) GetToolSchema(name string) (map[string]interface{}, error) {
	tool, err := r.GetTool(name)
	if err != nil {
		return nil, err
	}

	return tool.GetSchema(), nil
}

// ToolInfo represents information about a tool
type ToolInfo struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Schema      map[string]interface{} `json:"schema"`
}

// GetToolInfo returns information about a tool
func (r *ToolRegistry) GetToolInfo(name string) (*ToolInfo, error) {
	tool, err := r.GetTool(name)
	if err != nil {
		return nil, err
	}

	return &ToolInfo{
		Name:        tool.GetName(),
		Description: tool.GetDescription(),
		Schema:      tool.GetSchema(),
	}, nil
}

// GetAllToolsInfo returns information about all registered tools
func (r *ToolRegistry) GetAllToolsInfo() []*ToolInfo {
	infos := make([]*ToolInfo, 0, len(r.tools))

	for _, tool := range r.tools {
		info := &ToolInfo{
			Name:        tool.GetName(),
			Description: tool.GetDescription(),
			Schema:      tool.GetSchema(),
		}
		infos = append(infos, info)
	}

	return infos
}

// MockTool represents a mock tool for testing
type MockTool struct {
	*BaseTool
	mockResponse map[string]interface{}
	mockError    error
}

// NewMockTool creates a new mock tool
func NewMockTool(name, description string, response map[string]interface{}, err error) *MockTool {
	config := &ToolConfig{
		Name:        name,
		Description: description,
	}

	return &MockTool{
		BaseTool:     NewBaseTool(config),
		mockResponse: response,
		mockError:    err,
	}
}

// Execute executes the mock tool
func (t *MockTool) Execute(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	ctx, span := t.tracer.Start(ctx, "mock_tool.execute")
	defer span.End()

	span.SetAttributes(
		attribute.String("tool.name", t.GetName()),
		attribute.String("tool.type", "mock"),
	)

	if t.mockError != nil {
		span.RecordError(t.mockError)
		return nil, t.mockError
	}

	// Add input to response for testing
	response := make(map[string]interface{})
	for k, v := range t.mockResponse {
		response[k] = v
	}
	response["input"] = input
	response["timestamp"] = time.Now().Unix()

	return response, nil
}

// GetSchema returns the mock tool schema
func (t *MockTool) GetSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"message": map[string]interface{}{
				"type":        "string",
				"description": "Test message",
			},
		},
	}
}

// JSONSchemaTool represents a tool with JSON schema validation
type JSONSchemaTool struct {
	*BaseTool
	schema   map[string]interface{}
	executor func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error)
}

// NewJSONSchemaTool creates a new JSON schema tool
func NewJSONSchemaTool(config *ToolConfig, schema map[string]interface{}, executor func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error)) *JSONSchemaTool {
	return &JSONSchemaTool{
		BaseTool: NewBaseTool(config),
		schema:   schema,
		executor: executor,
	}
}

// Execute executes the JSON schema tool
func (t *JSONSchemaTool) Execute(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	ctx, span := t.tracer.Start(ctx, "json_schema_tool.execute")
	defer span.End()

	span.SetAttributes(
		attribute.String("tool.name", t.GetName()),
		attribute.String("tool.type", "json_schema"),
	)

	// Validate input against schema (basic validation)
	if err := t.validateInput(input); err != nil {
		span.RecordError(err)
		return nil, NewToolError("validation_error", err.Error(), t.GetName(), nil)
	}

	// Execute the tool with retry
	var result map[string]interface{}
	err := t.WithRetry(ctx, func() error {
		var execErr error
		result, execErr = t.executor(ctx, input)
		return execErr
	})

	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	return result, nil
}

// validateInput performs basic input validation
func (t *JSONSchemaTool) validateInput(input map[string]interface{}) error {
	// Basic validation - in production, use a proper JSON schema validator
	if t.schema == nil {
		return nil
	}

	properties, ok := t.schema["properties"].(map[string]interface{})
	if !ok {
		return nil
	}

	required, _ := t.schema["required"].([]interface{})

	// Check required fields
	for _, req := range required {
		if reqStr, ok := req.(string); ok {
			if _, exists := input[reqStr]; !exists {
				return fmt.Errorf("required field missing: %s", reqStr)
			}
		}
	}

	// Basic type checking
	for field, value := range input {
		if propSchema, exists := properties[field]; exists {
			if propMap, ok := propSchema.(map[string]interface{}); ok {
				if expectedType, ok := propMap["type"].(string); ok {
					if !t.validateType(value, expectedType) {
						return fmt.Errorf("field %s has invalid type, expected %s", field, expectedType)
					}
				}
			}
		}
	}

	return nil
}

// validateType performs basic type validation
func (t *JSONSchemaTool) validateType(value interface{}, expectedType string) bool {
	switch expectedType {
	case "string":
		_, ok := value.(string)
		return ok
	case "number":
		_, ok1 := value.(float64)
		_, ok2 := value.(int)
		return ok1 || ok2
	case "integer":
		_, ok := value.(int)
		return ok
	case "boolean":
		_, ok := value.(bool)
		return ok
	case "array":
		_, ok := value.([]interface{})
		return ok
	case "object":
		_, ok := value.(map[string]interface{})
		return ok
	default:
		return true // Unknown type, allow it
	}
}

// GetSchema returns the tool schema
func (t *JSONSchemaTool) GetSchema() map[string]interface{} {
	return t.schema
}
