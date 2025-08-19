package memory

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Local type definitions to avoid import cycles

// Message represents a message in a conversation
type Message struct {
	Role      string     `json:"role"`
	Content   string     `json:"content"`
	Name      string     `json:"name,omitempty"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
}

// ToolCall represents a tool call in a message
type ToolCall struct {
	ID       string   `json:"id"`
	Type     string   `json:"type"`
	Function Function `json:"function"`
}

// Function represents a function call
type Function struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// LLMProvider interface for LLM providers (local definition to avoid import cycle)
type LLMProvider interface {
	GenerateResponse(ctx context.Context, req *GenerateRequest) (*GenerateResponse, error)
}

// GenerateRequest represents a request to generate text
type GenerateRequest struct {
	Messages    []Message `json:"messages"`
	Model       string    `json:"model,omitempty"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Temperature float64   `json:"temperature,omitempty"`
}

// Choice represents a choice in an LLM response
type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

// Usage represents token usage information
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// GenerateResponse represents a response from text generation
type GenerateResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

// Memory interface for storing and retrieving conversation context
type Memory interface {
	Store(ctx context.Context, key string, value interface{}) error
	Retrieve(ctx context.Context, key string) (interface{}, error)
	Clear(ctx context.Context, key string) error
	GetHistory(ctx context.Context, key string, limit int) ([]Message, error)
	AddMessage(ctx context.Context, key string, message Message) error
	GetMessages(ctx context.Context, key string, limit int) ([]Message, error)
	SetTTL(ctx context.Context, key string, ttl time.Duration) error
	Close() error
}

// MemoryEntry represents a stored memory entry
type MemoryEntry struct {
	Key       string      `json:"key"`
	Value     interface{} `json:"value"`
	Timestamp time.Time   `json:"timestamp"`
	TTL       *time.Time  `json:"ttl,omitempty"`
}

// InMemoryStorage provides in-memory storage for development/testing
type InMemoryStorage struct {
	data     map[string]*MemoryEntry
	messages map[string][]Message
	mutex    sync.RWMutex
	tracer   trace.Tracer
}

// NewInMemoryStorage creates a new in-memory storage
func NewInMemoryStorage() *InMemoryStorage {
	storage := &InMemoryStorage{
		data:     make(map[string]*MemoryEntry),
		messages: make(map[string][]Message),
		tracer:   otel.Tracer("llm.memory.inmemory"),
	}

	// Start cleanup goroutine for TTL
	go storage.cleanupExpired()

	return storage
}

// Store stores a value with the given key
func (m *InMemoryStorage) Store(ctx context.Context, key string, value interface{}) error {
	ctx, span := m.tracer.Start(ctx, "inmemory.store")
	defer span.End()

	span.SetAttributes(
		attribute.String("memory.key", key),
		attribute.String("memory.type", "inmemory"),
	)

	m.mutex.Lock()
	defer m.mutex.Unlock()

	entry := &MemoryEntry{
		Key:       key,
		Value:     value,
		Timestamp: time.Now(),
	}

	m.data[key] = entry

	return nil
}

// Retrieve retrieves a value by key
func (m *InMemoryStorage) Retrieve(ctx context.Context, key string) (interface{}, error) {
	ctx, span := m.tracer.Start(ctx, "inmemory.retrieve")
	defer span.End()

	span.SetAttributes(
		attribute.String("memory.key", key),
		attribute.String("memory.type", "inmemory"),
	)

	m.mutex.RLock()
	defer m.mutex.RUnlock()

	entry, exists := m.data[key]
	if !exists {
		return nil, fmt.Errorf("key not found: %s", key)
	}

	// Check TTL
	if entry.TTL != nil && time.Now().After(*entry.TTL) {
		delete(m.data, key)
		return nil, fmt.Errorf("key expired: %s", key)
	}

	return entry.Value, nil
}

// Clear removes a key from storage
func (m *InMemoryStorage) Clear(ctx context.Context, key string) error {
	ctx, span := m.tracer.Start(ctx, "inmemory.clear")
	defer span.End()

	span.SetAttributes(
		attribute.String("memory.key", key),
		attribute.String("memory.type", "inmemory"),
	)

	m.mutex.Lock()
	defer m.mutex.Unlock()

	delete(m.data, key)
	delete(m.messages, key)

	return nil
}

// GetHistory returns conversation history (alias for GetMessages)
func (m *InMemoryStorage) GetHistory(ctx context.Context, key string, limit int) ([]Message, error) {
	return m.GetMessages(ctx, key, limit)
}

// AddMessage adds a message to the conversation history
func (m *InMemoryStorage) AddMessage(ctx context.Context, key string, message Message) error {
	ctx, span := m.tracer.Start(ctx, "inmemory.add_message")
	defer span.End()

	span.SetAttributes(
		attribute.String("memory.key", key),
		attribute.String("message.role", message.Role),
		attribute.String("memory.type", "inmemory"),
	)

	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.messages[key] == nil {
		m.messages[key] = make([]Message, 0)
	}

	m.messages[key] = append(m.messages[key], message)

	return nil
}

// GetMessages retrieves conversation messages
func (m *InMemoryStorage) GetMessages(ctx context.Context, key string, limit int) ([]Message, error) {
	ctx, span := m.tracer.Start(ctx, "inmemory.get_messages")
	defer span.End()

	span.SetAttributes(
		attribute.String("memory.key", key),
		attribute.Int("memory.limit", limit),
		attribute.String("memory.type", "inmemory"),
	)

	m.mutex.RLock()
	defer m.mutex.RUnlock()

	messages, exists := m.messages[key]
	if !exists {
		return []Message{}, nil
	}

	// Apply limit
	if limit > 0 && len(messages) > limit {
		// Return the most recent messages
		start := len(messages) - limit
		messages = messages[start:]
	}

	// Return a copy to prevent external modification
	result := make([]Message, len(messages))
	copy(result, messages)

	return result, nil
}

// SetTTL sets a time-to-live for a key
func (m *InMemoryStorage) SetTTL(ctx context.Context, key string, ttl time.Duration) error {
	ctx, span := m.tracer.Start(ctx, "inmemory.set_ttl")
	defer span.End()

	span.SetAttributes(
		attribute.String("memory.key", key),
		attribute.String("memory.ttl", ttl.String()),
		attribute.String("memory.type", "inmemory"),
	)

	m.mutex.Lock()
	defer m.mutex.Unlock()

	entry, exists := m.data[key]
	if !exists {
		return fmt.Errorf("key not found: %s", key)
	}

	expiryTime := time.Now().Add(ttl)
	entry.TTL = &expiryTime

	return nil
}

// Close cleans up resources
func (m *InMemoryStorage) Close() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.data = make(map[string]*MemoryEntry)
	m.messages = make(map[string][]Message)

	return nil
}

// cleanupExpired removes expired entries
func (m *InMemoryStorage) cleanupExpired() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		m.mutex.Lock()
		now := time.Now()

		for key, entry := range m.data {
			if entry.TTL != nil && now.After(*entry.TTL) {
				delete(m.data, key)
				delete(m.messages, key)
			}
		}

		m.mutex.Unlock()
	}
}

// ConversationMemory provides conversation-specific memory management
type ConversationMemory struct {
	storage        Memory
	conversationID string
	maxMessages    int
	tracer         trace.Tracer
}

// NewConversationMemory creates a new conversation memory
func NewConversationMemory(storage Memory, conversationID string, maxMessages int) *ConversationMemory {
	if maxMessages <= 0 {
		maxMessages = 100 // Default max messages
	}

	return &ConversationMemory{
		storage:        storage,
		conversationID: conversationID,
		maxMessages:    maxMessages,
		tracer:         otel.Tracer("llm.memory.conversation"),
	}
}

// Store stores conversation data
func (c *ConversationMemory) Store(ctx context.Context, key string, value interface{}) error {
	fullKey := c.getFullKey(key)
	return c.storage.Store(ctx, fullKey, value)
}

// Retrieve retrieves conversation data
func (c *ConversationMemory) Retrieve(ctx context.Context, key string) (interface{}, error) {
	fullKey := c.getFullKey(key)
	return c.storage.Retrieve(ctx, fullKey)
}

// Clear clears conversation data
func (c *ConversationMemory) Clear(ctx context.Context, key string) error {
	fullKey := c.getFullKey(key)
	return c.storage.Clear(ctx, fullKey)
}

// GetHistory returns conversation message history
func (c *ConversationMemory) GetHistory(ctx context.Context, key string, limit int) ([]Message, error) {
	fullKey := c.getFullKey(key)
	return c.storage.GetHistory(ctx, fullKey, limit)
}

// AddMessage adds a message to conversation history
func (c *ConversationMemory) AddMessage(ctx context.Context, key string, message Message) error {
	ctx, span := c.tracer.Start(ctx, "conversation.add_message")
	defer span.End()

	span.SetAttributes(
		attribute.String("conversation.id", c.conversationID),
		attribute.String("message.role", message.Role),
	)

	fullKey := c.getFullKey(key)

	// Add the message
	err := c.storage.AddMessage(ctx, fullKey, message)
	if err != nil {
		span.RecordError(err)
		return err
	}

	// Trim messages if we exceed the limit
	messages, err := c.storage.GetMessages(ctx, fullKey, 0) // Get all messages
	if err != nil {
		return err
	}

	if len(messages) > c.maxMessages {
		// Keep only the most recent messages
		trimmedMessages := messages[len(messages)-c.maxMessages:]

		// Clear and re-add trimmed messages
		if err := c.storage.Clear(ctx, fullKey); err != nil {
			return err
		}

		for _, msg := range trimmedMessages {
			if err := c.storage.AddMessage(ctx, fullKey, msg); err != nil {
				return err
			}
		}
	}

	return nil
}

// GetMessages retrieves conversation messages
func (c *ConversationMemory) GetMessages(ctx context.Context, key string, limit int) ([]Message, error) {
	fullKey := c.getFullKey(key)
	return c.storage.GetMessages(ctx, fullKey, limit)
}

// SetTTL sets TTL for conversation data
func (c *ConversationMemory) SetTTL(ctx context.Context, key string, ttl time.Duration) error {
	fullKey := c.getFullKey(key)
	return c.storage.SetTTL(ctx, fullKey, ttl)
}

// Close cleans up conversation memory
func (c *ConversationMemory) Close() error {
	// Don't close the underlying storage as it might be shared
	return nil
}

// getFullKey creates a full key with conversation prefix
func (c *ConversationMemory) getFullKey(key string) string {
	return fmt.Sprintf("conversation:%s:%s", c.conversationID, key)
}

// SummaryMemory provides memory with automatic summarization
type SummaryMemory struct {
	storage       Memory
	llmProvider   LLMProvider
	maxMessages   int
	summaryPrompt string
	tracer        trace.Tracer
}

// NewSummaryMemory creates a new summary memory
func NewSummaryMemory(storage Memory, llmProvider LLMProvider, maxMessages int) *SummaryMemory {
	if maxMessages <= 0 {
		maxMessages = 50
	}

	defaultPrompt := `Please summarize the following conversation history in a concise way that preserves the key context and important details:

{{conversation_history}}

Summary:`

	return &SummaryMemory{
		storage:       storage,
		llmProvider:   llmProvider,
		maxMessages:   maxMessages,
		summaryPrompt: defaultPrompt,
		tracer:        otel.Tracer("llm.memory.summary"),
	}
}

// AddMessage adds a message and triggers summarization if needed
func (s *SummaryMemory) AddMessage(ctx context.Context, key string, message Message) error {
	ctx, span := s.tracer.Start(ctx, "summary.add_message")
	defer span.End()

	// Add the message
	err := s.storage.AddMessage(ctx, key, message)
	if err != nil {
		span.RecordError(err)
		return err
	}

	// Check if we need to summarize
	messages, err := s.storage.GetMessages(ctx, key, 0)
	if err != nil {
		return err
	}

	if len(messages) > s.maxMessages {
		// Trigger summarization
		if err := s.summarizeAndTrim(ctx, key, messages); err != nil {
			span.RecordError(err)
			return fmt.Errorf("failed to summarize conversation: %w", err)
		}
	}

	return nil
}

// summarizeAndTrim summarizes old messages and keeps recent ones
func (s *SummaryMemory) summarizeAndTrim(ctx context.Context, key string, messages []Message) error {
	// Split messages: older half for summary, newer half to keep
	splitPoint := len(messages) / 2
	oldMessages := messages[:splitPoint]
	recentMessages := messages[splitPoint:]

	// Create conversation history text
	var historyText string
	for _, msg := range oldMessages {
		historyText += fmt.Sprintf("%s: %s\n", msg.Role, msg.Content)
	}

	// Generate summary
	summaryReq := &GenerateRequest{
		Messages: []Message{
			{
				Role:    "user",
				Content: fmt.Sprintf(s.summaryPrompt, historyText),
			},
		},
		MaxTokens:   500,
		Temperature: 0.3,
	}

	response, err := s.llmProvider.GenerateResponse(ctx, summaryReq)
	if err != nil {
		return fmt.Errorf("failed to generate summary: %w", err)
	}

	if len(response.Choices) == 0 {
		return fmt.Errorf("no summary generated")
	}

	summary := response.Choices[0].Message.Content

	// Clear old messages and add summary + recent messages
	if err := s.storage.Clear(ctx, key); err != nil {
		return err
	}

	// Add summary as a system message
	summaryMessage := Message{
		Role:    "system",
		Content: fmt.Sprintf("Previous conversation summary: %s", summary),
	}

	if err := s.storage.AddMessage(ctx, key, summaryMessage); err != nil {
		return err
	}

	// Add recent messages
	for _, msg := range recentMessages {
		if err := s.storage.AddMessage(ctx, key, msg); err != nil {
			return err
		}
	}

	return nil
}

// Delegate other methods to underlying storage
func (s *SummaryMemory) Store(ctx context.Context, key string, value interface{}) error {
	return s.storage.Store(ctx, key, value)
}

func (s *SummaryMemory) Retrieve(ctx context.Context, key string) (interface{}, error) {
	return s.storage.Retrieve(ctx, key)
}

func (s *SummaryMemory) Clear(ctx context.Context, key string) error {
	return s.storage.Clear(ctx, key)
}

func (s *SummaryMemory) GetHistory(ctx context.Context, key string, limit int) ([]Message, error) {
	return s.storage.GetHistory(ctx, key, limit)
}

func (s *SummaryMemory) GetMessages(ctx context.Context, key string, limit int) ([]Message, error) {
	return s.storage.GetMessages(ctx, key, limit)
}

func (s *SummaryMemory) SetTTL(ctx context.Context, key string, ttl time.Duration) error {
	return s.storage.SetTTL(ctx, key, ttl)
}

func (s *SummaryMemory) Close() error {
	return s.storage.Close()
}
