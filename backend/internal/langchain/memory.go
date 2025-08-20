package langchain

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Memory defines the interface for conversation memory
type Memory interface {
	// AddMessage adds a message to memory
	AddMessage(ctx context.Context, message *Message) error
	
	// GetMessages retrieves messages from memory
	GetMessages(ctx context.Context, sessionID string, limit int) ([]*Message, error)
	
	// Clear clears memory for a session
	Clear(ctx context.Context, sessionID string) error
	
	// GetSummary returns a summary of the conversation
	GetSummary(ctx context.Context, sessionID string) (string, error)
	
	// GetName returns the memory type name
	GetName() string
}

// Message represents a conversation message
type Message struct {
	ID        string                 `json:"id"`
	SessionID string                 `json:"session_id"`
	Role      string                 `json:"role"`      // "user", "assistant", "system"
	Content   string                 `json:"content"`
	Metadata  map[string]interface{} `json:"metadata"`
	Timestamp time.Time              `json:"timestamp"`
}

// BaseMemory provides common functionality for all memory types
type BaseMemory struct {
	Name   string       `json:"name"`
	tracer trace.Tracer `json:"-"`
}

// NewBaseMemory creates a new base memory
func NewBaseMemory(name string) *BaseMemory {
	return &BaseMemory{
		Name:   name,
		tracer: otel.Tracer("langchain.memory"),
	}
}

// GetName returns the memory name
func (m *BaseMemory) GetName() string {
	return m.Name
}

// BufferMemory stores conversation messages in memory with a size limit
type BufferMemory struct {
	*BaseMemory
	sessions map[string][]*Message
	maxSize  int
	mutex    sync.RWMutex
}

// NewBufferMemory creates a new buffer memory
func NewBufferMemory(name string, maxSize int) *BufferMemory {
	return &BufferMemory{
		BaseMemory: NewBaseMemory(name),
		sessions:   make(map[string][]*Message),
		maxSize:    maxSize,
	}
}

// AddMessage adds a message to buffer memory
func (m *BufferMemory) AddMessage(ctx context.Context, message *Message) error {
	ctx, span := m.tracer.Start(ctx, "buffer_memory.add_message")
	defer span.End()

	span.SetAttributes(
		attribute.String("memory.name", m.Name),
		attribute.String("message.session_id", message.SessionID),
		attribute.String("message.role", message.Role),
	)

	m.mutex.Lock()
	defer m.mutex.Unlock()

	if message.Timestamp.IsZero() {
		message.Timestamp = time.Now()
	}

	if message.ID == "" {
		message.ID = fmt.Sprintf("msg_%d", time.Now().UnixNano())
	}

	messages := m.sessions[message.SessionID]
	messages = append(messages, message)

	// Trim to max size if needed
	if len(messages) > m.maxSize {
		messages = messages[len(messages)-m.maxSize:]
	}

	m.sessions[message.SessionID] = messages

	span.SetAttributes(attribute.Int("session.message_count", len(messages)))
	return nil
}

// GetMessages retrieves messages from buffer memory
func (m *BufferMemory) GetMessages(ctx context.Context, sessionID string, limit int) ([]*Message, error) {
	ctx, span := m.tracer.Start(ctx, "buffer_memory.get_messages")
	defer span.End()

	span.SetAttributes(
		attribute.String("memory.name", m.Name),
		attribute.String("session_id", sessionID),
		attribute.Int("limit", limit),
	)

	m.mutex.RLock()
	defer m.mutex.RUnlock()

	messages, exists := m.sessions[sessionID]
	if !exists {
		return []*Message{}, nil
	}

	// Apply limit
	if limit > 0 && len(messages) > limit {
		messages = messages[len(messages)-limit:]
	}

	// Return copies to prevent external modification
	result := make([]*Message, len(messages))
	for i, msg := range messages {
		result[i] = &Message{
			ID:        msg.ID,
			SessionID: msg.SessionID,
			Role:      msg.Role,
			Content:   msg.Content,
			Metadata:  make(map[string]interface{}),
			Timestamp: msg.Timestamp,
		}
		
		// Copy metadata
		for k, v := range msg.Metadata {
			result[i].Metadata[k] = v
		}
	}

	span.SetAttributes(attribute.Int("returned.message_count", len(result)))
	return result, nil
}

// Clear clears memory for a session
func (m *BufferMemory) Clear(ctx context.Context, sessionID string) error {
	ctx, span := m.tracer.Start(ctx, "buffer_memory.clear")
	defer span.End()

	span.SetAttributes(
		attribute.String("memory.name", m.Name),
		attribute.String("session_id", sessionID),
	)

	m.mutex.Lock()
	defer m.mutex.Unlock()

	delete(m.sessions, sessionID)
	return nil
}

// GetSummary returns a summary of the conversation
func (m *BufferMemory) GetSummary(ctx context.Context, sessionID string) (string, error) {
	messages, err := m.GetMessages(ctx, sessionID, 0)
	if err != nil {
		return "", err
	}

	if len(messages) == 0 {
		return "No conversation history", nil
	}

	// Simple summary - count messages by role
	userCount := 0
	assistantCount := 0
	systemCount := 0

	for _, msg := range messages {
		switch msg.Role {
		case "user":
			userCount++
		case "assistant":
			assistantCount++
		case "system":
			systemCount++
		}
	}

	return fmt.Sprintf("Conversation with %d user messages, %d assistant responses, %d system messages",
		userCount, assistantCount, systemCount), nil
}

// SummaryMemory stores a summary of the conversation instead of all messages
type SummaryMemory struct {
	*BaseMemory
	sessions    map[string]*ConversationSummary
	maxMessages int
	mutex       sync.RWMutex
}

// ConversationSummary represents a summarized conversation
type ConversationSummary struct {
	SessionID     string                 `json:"session_id"`
	Summary       string                 `json:"summary"`
	RecentMessages []*Message            `json:"recent_messages"`
	MessageCount  int                    `json:"message_count"`
	LastUpdated   time.Time              `json:"last_updated"`
	Metadata      map[string]interface{} `json:"metadata"`
}

// NewSummaryMemory creates a new summary memory
func NewSummaryMemory(name string, maxMessages int) *SummaryMemory {
	return &SummaryMemory{
		BaseMemory:  NewBaseMemory(name),
		sessions:    make(map[string]*ConversationSummary),
		maxMessages: maxMessages,
	}
}

// AddMessage adds a message to summary memory
func (m *SummaryMemory) AddMessage(ctx context.Context, message *Message) error {
	ctx, span := m.tracer.Start(ctx, "summary_memory.add_message")
	defer span.End()

	span.SetAttributes(
		attribute.String("memory.name", m.Name),
		attribute.String("message.session_id", message.SessionID),
		attribute.String("message.role", message.Role),
	)

	m.mutex.Lock()
	defer m.mutex.Unlock()

	if message.Timestamp.IsZero() {
		message.Timestamp = time.Now()
	}

	if message.ID == "" {
		message.ID = fmt.Sprintf("msg_%d", time.Now().UnixNano())
	}

	summary, exists := m.sessions[message.SessionID]
	if !exists {
		summary = &ConversationSummary{
			SessionID:      message.SessionID,
			Summary:        "",
			RecentMessages: make([]*Message, 0),
			MessageCount:   0,
			LastUpdated:    time.Now(),
			Metadata:       make(map[string]interface{}),
		}
		m.sessions[message.SessionID] = summary
	}

	// Add to recent messages
	summary.RecentMessages = append(summary.RecentMessages, message)
	summary.MessageCount++
	summary.LastUpdated = time.Now()

	// Trim recent messages if needed
	if len(summary.RecentMessages) > m.maxMessages {
		// TODO: In a real implementation, you'd generate a summary of the older messages
		// and update the summary field before removing them
		summary.RecentMessages = summary.RecentMessages[len(summary.RecentMessages)-m.maxMessages:]
	}

	span.SetAttributes(
		attribute.Int("session.message_count", summary.MessageCount),
		attribute.Int("session.recent_count", len(summary.RecentMessages)),
	)

	return nil
}

// GetMessages retrieves recent messages from summary memory
func (m *SummaryMemory) GetMessages(ctx context.Context, sessionID string, limit int) ([]*Message, error) {
	ctx, span := m.tracer.Start(ctx, "summary_memory.get_messages")
	defer span.End()

	span.SetAttributes(
		attribute.String("memory.name", m.Name),
		attribute.String("session_id", sessionID),
		attribute.Int("limit", limit),
	)

	m.mutex.RLock()
	defer m.mutex.RUnlock()

	summary, exists := m.sessions[sessionID]
	if !exists {
		return []*Message{}, nil
	}

	messages := summary.RecentMessages

	// Apply limit
	if limit > 0 && len(messages) > limit {
		messages = messages[len(messages)-limit:]
	}

	// Return copies
	result := make([]*Message, len(messages))
	for i, msg := range messages {
		result[i] = &Message{
			ID:        msg.ID,
			SessionID: msg.SessionID,
			Role:      msg.Role,
			Content:   msg.Content,
			Metadata:  make(map[string]interface{}),
			Timestamp: msg.Timestamp,
		}
		
		for k, v := range msg.Metadata {
			result[i].Metadata[k] = v
		}
	}

	span.SetAttributes(attribute.Int("returned.message_count", len(result)))
	return result, nil
}

// Clear clears memory for a session
func (m *SummaryMemory) Clear(ctx context.Context, sessionID string) error {
	ctx, span := m.tracer.Start(ctx, "summary_memory.clear")
	defer span.End()

	span.SetAttributes(
		attribute.String("memory.name", m.Name),
		attribute.String("session_id", sessionID),
	)

	m.mutex.Lock()
	defer m.mutex.Unlock()

	delete(m.sessions, sessionID)
	return nil
}

// GetSummary returns the conversation summary
func (m *SummaryMemory) GetSummary(ctx context.Context, sessionID string) (string, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	summary, exists := m.sessions[sessionID]
	if !exists {
		return "No conversation history", nil
	}

	if summary.Summary != "" {
		return fmt.Sprintf("%s (Total messages: %d, Recent: %d)",
			summary.Summary, summary.MessageCount, len(summary.RecentMessages)), nil
	}

	return fmt.Sprintf("Conversation with %d messages (%d recent)",
		summary.MessageCount, len(summary.RecentMessages)), nil
}

// WindowMemory keeps only the most recent N messages
type WindowMemory struct {
	*BaseMemory
	sessions   map[string][]*Message
	windowSize int
	mutex      sync.RWMutex
}

// NewWindowMemory creates a new window memory
func NewWindowMemory(name string, windowSize int) *WindowMemory {
	return &WindowMemory{
		BaseMemory: NewBaseMemory(name),
		sessions:   make(map[string][]*Message),
		windowSize: windowSize,
	}
}

// AddMessage adds a message to window memory
func (m *WindowMemory) AddMessage(ctx context.Context, message *Message) error {
	ctx, span := m.tracer.Start(ctx, "window_memory.add_message")
	defer span.End()

	span.SetAttributes(
		attribute.String("memory.name", m.Name),
		attribute.String("message.session_id", message.SessionID),
		attribute.Int("window_size", m.windowSize),
	)

	m.mutex.Lock()
	defer m.mutex.Unlock()

	if message.Timestamp.IsZero() {
		message.Timestamp = time.Now()
	}

	if message.ID == "" {
		message.ID = fmt.Sprintf("msg_%d", time.Now().UnixNano())
	}

	messages := m.sessions[message.SessionID]
	messages = append(messages, message)

	// Keep only the most recent windowSize messages
	if len(messages) > m.windowSize {
		messages = messages[len(messages)-m.windowSize:]
	}

	m.sessions[message.SessionID] = messages

	span.SetAttributes(attribute.Int("session.message_count", len(messages)))
	return nil
}

// GetMessages retrieves messages from window memory
func (m *WindowMemory) GetMessages(ctx context.Context, sessionID string, limit int) ([]*Message, error) {
	ctx, span := m.tracer.Start(ctx, "window_memory.get_messages")
	defer span.End()

	span.SetAttributes(
		attribute.String("memory.name", m.Name),
		attribute.String("session_id", sessionID),
		attribute.Int("limit", limit),
	)

	m.mutex.RLock()
	defer m.mutex.RUnlock()

	messages, exists := m.sessions[sessionID]
	if !exists {
		return []*Message{}, nil
	}

	// Apply limit
	if limit > 0 && len(messages) > limit {
		messages = messages[len(messages)-limit:]
	}

	// Return copies
	result := make([]*Message, len(messages))
	for i, msg := range messages {
		result[i] = &Message{
			ID:        msg.ID,
			SessionID: msg.SessionID,
			Role:      msg.Role,
			Content:   msg.Content,
			Metadata:  make(map[string]interface{}),
			Timestamp: msg.Timestamp,
		}
		
		for k, v := range msg.Metadata {
			result[i].Metadata[k] = v
		}
	}

	span.SetAttributes(attribute.Int("returned.message_count", len(result)))
	return result, nil
}

// Clear clears memory for a session
func (m *WindowMemory) Clear(ctx context.Context, sessionID string) error {
	ctx, span := m.tracer.Start(ctx, "window_memory.clear")
	defer span.End()

	span.SetAttributes(
		attribute.String("memory.name", m.Name),
		attribute.String("session_id", sessionID),
	)

	m.mutex.Lock()
	defer m.mutex.Unlock()

	delete(m.sessions, sessionID)
	return nil
}

// GetSummary returns a summary of the conversation
func (m *WindowMemory) GetSummary(ctx context.Context, sessionID string) (string, error) {
	messages, err := m.GetMessages(ctx, sessionID, 0)
	if err != nil {
		return "", err
	}

	if len(messages) == 0 {
		return "No conversation history", nil
	}

	return fmt.Sprintf("Recent conversation window with %d messages (max %d)",
		len(messages), m.windowSize), nil
}

// MemoryManager manages multiple memory instances
type MemoryManager struct {
	memories map[string]Memory
	mutex    sync.RWMutex
	tracer   trace.Tracer
}

// NewMemoryManager creates a new memory manager
func NewMemoryManager() *MemoryManager {
	return &MemoryManager{
		memories: make(map[string]Memory),
		tracer:   otel.Tracer("langchain.memory_manager"),
	}
}

// RegisterMemory registers a memory instance
func (mm *MemoryManager) RegisterMemory(memory Memory) {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()
	
	mm.memories[memory.GetName()] = memory
}

// GetMemory retrieves a memory instance by name
func (mm *MemoryManager) GetMemory(name string) (Memory, error) {
	mm.mutex.RLock()
	defer mm.mutex.RUnlock()
	
	memory, exists := mm.memories[name]
	if !exists {
		return nil, fmt.Errorf("memory '%s' not found", name)
	}
	
	return memory, nil
}

// ListMemories returns all registered memory names
func (mm *MemoryManager) ListMemories() []string {
	mm.mutex.RLock()
	defer mm.mutex.RUnlock()
	
	names := make([]string, 0, len(mm.memories))
	for name := range mm.memories {
		names = append(names, name)
	}
	
	sort.Strings(names)
	return names
}

// AddMessageToMemory adds a message to a specific memory
func (mm *MemoryManager) AddMessageToMemory(ctx context.Context, memoryName string, message *Message) error {
	memory, err := mm.GetMemory(memoryName)
	if err != nil {
		return err
	}
	
	return memory.AddMessage(ctx, message)
}

// GetMessagesFromMemory retrieves messages from a specific memory
func (mm *MemoryManager) GetMessagesFromMemory(ctx context.Context, memoryName, sessionID string, limit int) ([]*Message, error) {
	memory, err := mm.GetMemory(memoryName)
	if err != nil {
		return nil, err
	}
	
	return memory.GetMessages(ctx, sessionID, limit)
}
