package llm

import (
	"context"
	"fmt"
	"sync"

	"github.com/exotic-travel-booking/backend/internal/llm/chains"
	"github.com/exotic-travel-booking/backend/internal/llm/memory"
	"github.com/exotic-travel-booking/backend/internal/llm/prompts"
	"github.com/exotic-travel-booking/backend/internal/llm/providers"
	"github.com/exotic-travel-booking/backend/internal/workflow"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// LLMManager manages LLM providers, chains, and memory
type LLMManager struct {
	providers       map[string]providers.LLMProvider
	defaultProvider string
	promptManager   *prompts.PromptManager
	memoryStorage   memory.Memory
	factory         *providers.ProviderFactory
	mutex           sync.RWMutex
	tracer          trace.Tracer
}

// NewLLMManager creates a new LLM manager
func NewLLMManager() *LLMManager {
	manager := &LLMManager{
		providers:     make(map[string]providers.LLMProvider),
		promptManager: prompts.NewPromptManager(),
		memoryStorage: memory.NewInMemoryStorage(),
		factory:       providers.NewProviderFactory(),
		tracer:        otel.Tracer("llm.manager"),
	}

	// Initialize travel prompts
	if err := prompts.InitializeTravelPrompts(manager.promptManager); err != nil {
		// Log error but don't fail initialization
		fmt.Printf("Warning: Failed to initialize travel prompts: %v\n", err)
	}

	return manager
}

// AddProvider adds an LLM provider
func (m *LLMManager) AddProvider(name string, config *providers.LLMConfig) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	provider, err := m.factory.CreateProvider(config)
	if err != nil {
		return fmt.Errorf("failed to create provider %s: %w", name, err)
	}

	m.providers[name] = provider

	// Set as default if it's the first provider
	if m.defaultProvider == "" {
		m.defaultProvider = name
	}

	return nil
}

// GetProvider retrieves a provider by name
func (m *LLMManager) GetProvider(name string) (providers.LLMProvider, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if name == "" {
		name = m.defaultProvider
	}

	provider, exists := m.providers[name]
	if !exists {
		return nil, fmt.Errorf("provider not found: %s", name)
	}

	return provider, nil
}

// SetDefaultProvider sets the default provider
func (m *LLMManager) SetDefaultProvider(name string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if _, exists := m.providers[name]; !exists {
		return fmt.Errorf("provider not found: %s", name)
	}

	m.defaultProvider = name
	return nil
}

// GetPromptManager returns the prompt manager
func (m *LLMManager) GetPromptManager() *prompts.PromptManager {
	return m.promptManager
}

// GetMemoryStorage returns the memory storage
func (m *LLMManager) GetMemoryStorage() memory.Memory {
	return m.memoryStorage
}

// SetMemoryStorage sets the memory storage
func (m *LLMManager) SetMemoryStorage(storage memory.Memory) {
	m.memoryStorage = storage
}

// GenerateResponse generates a response using the specified provider
func (m *LLMManager) GenerateResponse(ctx context.Context, providerName string, req *GenerateRequest) (*GenerateResponse, error) {
	ctx, span := m.tracer.Start(ctx, "llm_manager.generate_response")
	defer span.End()

	span.SetAttributes(
		attribute.String("llm.provider_name", providerName),
		attribute.String("llm.model", req.Model),
	)

	provider, err := m.GetProvider(providerName)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	// Convert request to provider format
	providerReq := m.convertToProviderRequest(req)

	// Generate response using provider
	providerResp, err := provider.GenerateResponse(ctx, providerReq)
	if err != nil {
		return nil, err
	}

	// Convert response back to main format
	return m.convertFromProviderResponse(providerResp), nil
}

// StreamResponse generates a streaming response using the specified provider
func (m *LLMManager) StreamResponse(ctx context.Context, providerName string, req *GenerateRequest) (<-chan *StreamChunk, error) {
	provider, err := m.GetProvider(providerName)
	if err != nil {
		return nil, err
	}

	// Convert request to provider format
	providerReq := m.convertToProviderRequest(req)

	// Get provider stream channel
	providerChan, err := provider.StreamResponse(ctx, providerReq)
	if err != nil {
		return nil, err
	}

	// Convert provider stream to llm stream
	llmChan := make(chan *StreamChunk)
	go func() {
		defer close(llmChan)
		for chunk := range providerChan {
			llmChan <- m.convertStreamChunk(chunk)
		}
	}()

	return llmChan, nil
}

// CreateChain creates a new chain with the specified type
func (m *LLMManager) CreateChain(chainType, name, description string) (chains.Chain, error) {
	switch chainType {
	case "sequential":
		return chains.NewSequentialChain(name, description), nil
	case "parallel":
		return chains.NewParallelChain(name, description, 5), nil
	case "conditional":
		return chains.NewConditionalChain(name, description), nil
	case "mapreduce":
		return chains.NewMapReduceChain(name, description, 5), nil
	default:
		return nil, fmt.Errorf("unknown chain type: %s", chainType)
	}
}

// CreateConversationMemory creates conversation-specific memory
func (m *LLMManager) CreateConversationMemory(conversationID string, maxMessages int) memory.Memory {
	return memory.NewConversationMemory(m.memoryStorage, conversationID, maxMessages)
}

// CreateSummaryMemory creates memory with automatic summarization
func (m *LLMManager) CreateSummaryMemory(providerName string, maxMessages int) (memory.Memory, error) {
	provider, err := m.GetProvider(providerName)
	if err != nil {
		return nil, err
	}

	// Create adapter for provider interface compatibility
	adapter := &memoryProviderAdapter{provider: provider}
	return memory.NewSummaryMemory(m.memoryStorage, adapter, maxMessages), nil
}

// ExecutePromptTemplate executes a prompt template with variables
func (m *LLMManager) ExecutePromptTemplate(ctx context.Context, providerName, templateName string, variables map[string]interface{}) (*GenerateResponse, error) {
	ctx, span := m.tracer.Start(ctx, "llm_manager.execute_prompt_template")
	defer span.End()

	span.SetAttributes(
		attribute.String("llm.provider_name", providerName),
		attribute.String("template.name", templateName),
	)

	// Render the template
	messages, err := m.promptManager.RenderToMessages(ctx, templateName, variables)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to render template: %w", err)
	}

	// Convert prompts messages to llm messages
	llmMessages := convertPromptsMessages(messages)

	// Create request
	req := &GenerateRequest{
		Messages: llmMessages,
	}

	// Generate response
	return m.GenerateResponse(ctx, providerName, req)
}

// ListProviders returns all available provider names
func (m *LLMManager) ListProviders() []string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	names := make([]string, 0, len(m.providers))
	for name := range m.providers {
		names = append(names, name)
	}
	return names
}

// GetProviderModels returns available models for a provider
func (m *LLMManager) GetProviderModels(ctx context.Context, providerName string) ([]string, error) {
	provider, err := m.GetProvider(providerName)
	if err != nil {
		return nil, err
	}

	return provider.GetModels(ctx)
}

// Close cleans up all resources
func (m *LLMManager) Close() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Close all providers
	for name, provider := range m.providers {
		if err := provider.Close(); err != nil {
			fmt.Printf("Warning: Failed to close provider %s: %v\n", name, err)
		}
	}

	// Close memory storage
	if err := m.memoryStorage.Close(); err != nil {
		fmt.Printf("Warning: Failed to close memory storage: %v\n", err)
	}

	return nil
}

// ConversationManager manages conversation state and context
type ConversationManager struct {
	llmManager    *LLMManager
	conversations map[string]*ConversationContext
	mutex         sync.RWMutex
	tracer        trace.Tracer
}

// ConversationContext represents the context of a conversation
type ConversationContext struct {
	ID        string                 `json:"id"`
	UserID    string                 `json:"user_id"`
	Messages  []Message              `json:"messages"`
	State     map[string]interface{} `json:"state"`
	Memory    memory.Memory          `json:"-"`
	CreatedAt int64                  `json:"created_at"`
	UpdatedAt int64                  `json:"updated_at"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// NewConversationManager creates a new conversation manager
func NewConversationManager(llmManager *LLMManager) *ConversationManager {
	return &ConversationManager{
		llmManager:    llmManager,
		conversations: make(map[string]*ConversationContext),
		tracer:        otel.Tracer("llm.conversation_manager"),
	}
}

// StartConversation starts a new conversation
func (cm *ConversationManager) StartConversation(ctx context.Context, conversationID, userID string) (*ConversationContext, error) {
	ctx, span := cm.tracer.Start(ctx, "conversation_manager.start_conversation")
	defer span.End()

	span.SetAttributes(
		attribute.String("conversation.id", conversationID),
		attribute.String("user.id", userID),
	)

	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	// Check if conversation already exists
	if _, exists := cm.conversations[conversationID]; exists {
		return nil, fmt.Errorf("conversation already exists: %s", conversationID)
	}

	// Create conversation memory
	conversationMemory := cm.llmManager.CreateConversationMemory(conversationID, 100)

	// Create conversation context
	conversation := &ConversationContext{
		ID:        conversationID,
		UserID:    userID,
		Messages:  make([]Message, 0),
		State:     make(map[string]interface{}),
		Memory:    conversationMemory,
		CreatedAt: ctx.Value("timestamp").(int64),
		UpdatedAt: ctx.Value("timestamp").(int64),
		Metadata:  make(map[string]interface{}),
	}

	cm.conversations[conversationID] = conversation

	return conversation, nil
}

// GetConversation retrieves a conversation by ID
func (cm *ConversationManager) GetConversation(conversationID string) (*ConversationContext, error) {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	conversation, exists := cm.conversations[conversationID]
	if !exists {
		return nil, fmt.Errorf("conversation not found: %s", conversationID)
	}

	return conversation, nil
}

// AddMessage adds a message to a conversation
func (cm *ConversationManager) AddMessage(ctx context.Context, conversationID string, message Message) error {
	ctx, span := cm.tracer.Start(ctx, "conversation_manager.add_message")
	defer span.End()

	span.SetAttributes(
		attribute.String("conversation.id", conversationID),
		attribute.String("message.role", message.Role),
	)

	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	conversation, exists := cm.conversations[conversationID]
	if !exists {
		span.RecordError(fmt.Errorf("conversation not found"))
		return fmt.Errorf("conversation not found: %s", conversationID)
	}

	// Add to conversation messages
	conversation.Messages = append(conversation.Messages, message)
	conversation.UpdatedAt = ctx.Value("timestamp").(int64)

	// Add to memory (convert to memory.Message)
	memoryMessage := convertToMemoryMessage(message)
	if err := conversation.Memory.AddMessage(ctx, "messages", memoryMessage); err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to add message to memory: %w", err)
	}

	return nil
}

// GenerateResponse generates a response for a conversation
func (cm *ConversationManager) GenerateResponse(ctx context.Context, conversationID, providerName string, userMessage string) (*GenerateResponse, error) {
	ctx, span := cm.tracer.Start(ctx, "conversation_manager.generate_response")
	defer span.End()

	span.SetAttributes(
		attribute.String("conversation.id", conversationID),
		attribute.String("llm.provider_name", providerName),
	)

	// Get conversation
	conversation, err := cm.GetConversation(conversationID)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	// Add user message
	userMsg := Message{
		Role:    "user",
		Content: userMessage,
	}

	if err := cm.AddMessage(ctx, conversationID, userMsg); err != nil {
		span.RecordError(err)
		return nil, err
	}

	// Get conversation history
	messages, err := conversation.Memory.GetMessages(ctx, "messages", 20) // Last 20 messages
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	// Convert memory messages to llm messages
	llmMessages := convertMemoryMessages(messages)

	// Generate response
	req := &GenerateRequest{
		Messages: llmMessages,
	}

	response, err := cm.llmManager.GenerateResponse(ctx, providerName, req)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	// Add assistant response to conversation
	if len(response.Choices) > 0 {
		assistantMsg := response.Choices[0].Message
		if err := cm.AddMessage(ctx, conversationID, assistantMsg); err != nil {
			span.RecordError(err)
			return nil, err
		}
	}

	return response, nil
}

// EndConversation ends a conversation and cleans up resources
func (cm *ConversationManager) EndConversation(ctx context.Context, conversationID string) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	conversation, exists := cm.conversations[conversationID]
	if !exists {
		return fmt.Errorf("conversation not found: %s", conversationID)
	}

	// Close conversation memory
	if err := conversation.Memory.Close(); err != nil {
		fmt.Printf("Warning: Failed to close conversation memory: %v\n", err)
	}

	// Remove from active conversations
	delete(cm.conversations, conversationID)

	return nil
}

// ListConversations returns all active conversation IDs
func (cm *ConversationManager) ListConversations() []string {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	ids := make([]string, 0, len(cm.conversations))
	for id := range cm.conversations {
		ids = append(ids, id)
	}
	return ids
}

// convertToProviderRequest converts from llm.GenerateRequest to providers.GenerateRequest
func (m *LLMManager) convertToProviderRequest(req *GenerateRequest) *providers.GenerateRequest {
	if req == nil {
		return nil
	}

	// Convert messages
	providerMessages := make([]providers.Message, len(req.Messages))
	for i, msg := range req.Messages {
		providerMessages[i] = providers.Message{
			Role:       msg.Role,
			Content:    msg.Content,
			ToolCalls:  convertToolCalls(msg.ToolCalls),
			ToolCallID: msg.ToolCallID,
			Metadata:   msg.Metadata,
		}
	}

	// Convert tools
	providerTools := make([]providers.Tool, len(req.Tools))
	for i, tool := range req.Tools {
		providerTools[i] = providers.Tool{
			Type: tool.Type,
			Function: providers.ToolFunction{
				Name:        tool.Function.Name,
				Description: tool.Function.Description,
				Parameters:  tool.Function.Parameters,
			},
		}
	}

	return &providers.GenerateRequest{
		Messages:     providerMessages,
		Model:        req.Model,
		MaxTokens:    req.MaxTokens,
		Temperature:  req.Temperature,
		TopP:         req.TopP,
		SystemPrompt: req.SystemPrompt,
		Tools:        providerTools,
		ToolChoice:   req.ToolChoice,
		Stream:       req.Stream,
		Metadata:     req.Metadata,
	}
}

// convertFromProviderResponse converts from providers.GenerateResponse to llm.GenerateResponse
func (m *LLMManager) convertFromProviderResponse(resp *providers.GenerateResponse) *GenerateResponse {
	if resp == nil {
		return nil
	}

	// Convert choices
	choices := make([]Choice, len(resp.Choices))
	for i, choice := range resp.Choices {
		choices[i] = Choice{
			Index: choice.Index,
			Message: Message{
				Role:       choice.Message.Role,
				Content:    choice.Message.Content,
				ToolCalls:  convertProviderToolCalls(choice.Message.ToolCalls),
				ToolCallID: choice.Message.ToolCallID,
				Metadata:   choice.Message.Metadata,
			},
			FinishReason: choice.FinishReason,
			Logprobs:     convertLogprobs(choice.Logprobs),
		}
	}

	return &GenerateResponse{
		ID:      resp.ID,
		Object:  resp.Object,
		Created: resp.Created,
		Model:   resp.Model,
		Choices: choices,
		Usage: Usage{
			PromptTokens:     resp.Usage.PromptTokens,
			CompletionTokens: resp.Usage.CompletionTokens,
			TotalTokens:      resp.Usage.TotalTokens,
		},
		SystemFingerprint: resp.SystemFingerprint,
		Metadata:          resp.Metadata,
	}
}

// convertToolCalls converts from llm.ToolCall to providers.ToolCall
func convertToolCalls(toolCalls []ToolCall) []providers.ToolCall {
	if toolCalls == nil {
		return nil
	}

	providerToolCalls := make([]providers.ToolCall, len(toolCalls))
	for i, tc := range toolCalls {
		providerToolCalls[i] = providers.ToolCall{
			ID:   tc.ID,
			Type: tc.Type,
			Function: providers.Function{
				Name:      tc.Function.Name,
				Arguments: tc.Function.Arguments,
			},
		}
	}
	return providerToolCalls
}

// convertProviderToolCalls converts from providers.ToolCall to llm.ToolCall
func convertProviderToolCalls(toolCalls []providers.ToolCall) []ToolCall {
	if toolCalls == nil {
		return nil
	}

	llmToolCalls := make([]ToolCall, len(toolCalls))
	for i, tc := range toolCalls {
		llmToolCalls[i] = ToolCall{
			ID:   tc.ID,
			Type: tc.Type,
			Function: struct {
				Name      string `json:"name"`
				Arguments string `json:"arguments"`
			}{
				Name:      tc.Function.Name,
				Arguments: tc.Function.Arguments,
			},
		}
	}
	return llmToolCalls
}

// convertLogprobs converts logprobs from provider format to llm format
func convertLogprobs(logprobs *struct {
	Content []providers.TokenLogprob `json:"content"`
}) *struct {
	Content []TokenLogprob `json:"content"`
} {
	if logprobs == nil {
		return nil
	}

	content := make([]TokenLogprob, len(logprobs.Content))
	for i, lp := range logprobs.Content {
		content[i] = TokenLogprob{
			Token:   lp.Token,
			Logprob: lp.Logprob,
			Bytes:   lp.Bytes,
		}
	}

	return &struct {
		Content []TokenLogprob `json:"content"`
	}{
		Content: content,
	}
}

// convertStreamChunk converts from providers.StreamChunk to llm.StreamChunk
func (m *LLMManager) convertStreamChunk(chunk *providers.StreamChunk) *StreamChunk {
	if chunk == nil {
		return nil
	}

	// Convert choices
	choices := make([]StreamChoice, len(chunk.Choices))
	for i, choice := range chunk.Choices {
		choices[i] = StreamChoice{
			Index: choice.Index,
			Delta: MessageDelta{
				Role:      choice.Delta.Role,
				Content:   choice.Delta.Content,
				ToolCalls: convertProviderToolCalls(choice.Delta.ToolCalls),
			},
			FinishReason: choice.FinishReason,
			Logprobs:     convertLogprobs(choice.Logprobs),
		}
	}

	return &StreamChunk{
		ID:      chunk.ID,
		Object:  chunk.Object,
		Created: chunk.Created,
		Model:   chunk.Model,
		Choices: choices,
		Usage: &Usage{
			PromptTokens:     chunk.Usage.PromptTokens,
			CompletionTokens: chunk.Usage.CompletionTokens,
			TotalTokens:      chunk.Usage.TotalTokens,
		},
		SystemFingerprint: chunk.SystemFingerprint,
	}
}

// convertPromptsMessages converts from prompts.Message to llm.Message
func convertPromptsMessages(messages []prompts.Message) []Message {
	llmMessages := make([]Message, len(messages))
	for i, msg := range messages {
		llmMessages[i] = Message{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}
	return llmMessages
}

// convertMemoryMessages converts from memory.Message to llm.Message
func convertMemoryMessages(messages []memory.Message) []Message {
	llmMessages := make([]Message, len(messages))
	for i, msg := range messages {
		llmMessages[i] = Message{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}
	return llmMessages
}

// convertToMemoryMessage converts from llm.Message to memory.Message
func convertToMemoryMessage(msg Message) memory.Message {
	// Convert tool calls
	var memoryToolCalls []memory.ToolCall
	for _, tc := range msg.ToolCalls {
		memoryToolCalls = append(memoryToolCalls, memory.ToolCall{
			ID:   tc.ID,
			Type: tc.Type,
			Function: memory.Function{
				Name:      tc.Function.Name,
				Arguments: tc.Function.Arguments,
			},
		})
	}

	return memory.Message{
		Role:      msg.Role,
		Content:   msg.Content,
		ToolCalls: memoryToolCalls,
	}
}

// memoryProviderAdapter adapts providers.LLMProvider to memory.LLMProvider
type memoryProviderAdapter struct {
	provider providers.LLMProvider
}

func (a *memoryProviderAdapter) GenerateResponse(ctx context.Context, req *memory.GenerateRequest) (*memory.GenerateResponse, error) {
	// Convert memory request to providers request
	providerReq := &providers.GenerateRequest{
		Model:       req.Model,
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
	}

	// Convert messages
	providerMessages := make([]providers.Message, len(req.Messages))
	for i, msg := range req.Messages {
		// Convert tool calls
		var providerToolCalls []providers.ToolCall
		for _, tc := range msg.ToolCalls {
			providerToolCalls = append(providerToolCalls, providers.ToolCall{
				ID:   tc.ID,
				Type: tc.Type,
				Function: providers.Function{
					Name:      tc.Function.Name,
					Arguments: tc.Function.Arguments,
				},
			})
		}

		providerMessages[i] = providers.Message{
			Role:      msg.Role,
			Content:   msg.Content,
			ToolCalls: providerToolCalls,
		}
	}
	providerReq.Messages = providerMessages

	// Call provider
	providerResp, err := a.provider.GenerateResponse(ctx, providerReq)
	if err != nil {
		return nil, err
	}

	// Convert response back
	choices := make([]memory.Choice, len(providerResp.Choices))
	for i, choice := range providerResp.Choices {
		// Convert tool calls in response
		var memoryToolCalls []memory.ToolCall
		for _, tc := range choice.Message.ToolCalls {
			memoryToolCalls = append(memoryToolCalls, memory.ToolCall{
				ID:   tc.ID,
				Type: tc.Type,
				Function: memory.Function{
					Name:      tc.Function.Name,
					Arguments: tc.Function.Arguments,
				},
			})
		}

		choices[i] = memory.Choice{
			Index: choice.Index,
			Message: memory.Message{
				Role:      choice.Message.Role,
				Content:   choice.Message.Content,
				ToolCalls: memoryToolCalls,
			},
			FinishReason: choice.FinishReason,
		}
	}

	return &memory.GenerateResponse{
		ID:      providerResp.ID,
		Object:  providerResp.Object,
		Created: providerResp.Created,
		Model:   providerResp.Model,
		Choices: choices,
		Usage: memory.Usage{
			PromptTokens:     providerResp.Usage.PromptTokens,
			CompletionTokens: providerResp.Usage.CompletionTokens,
			TotalTokens:      providerResp.Usage.TotalTokens,
		},
	}, nil
}

// agentLLMManagerAdapter adapts LLMManager to agents.LLMManagerInterface
type agentLLMManagerAdapter struct {
	llmManager *LLMManager
}

func NewAgentLLMManagerAdapter(llmManager *LLMManager) *agentLLMManagerAdapter {
	return &agentLLMManagerAdapter{llmManager: llmManager}
}

func (a *agentLLMManagerAdapter) GenerateResponse(ctx context.Context, providerName string, req *workflow.GenerateRequest) (*workflow.GenerateResponse, error) {
	// Convert workflow request to llm request
	llmReq := &GenerateRequest{
		Model:       req.Model,
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
	}

	// Convert messages
	llmMessages := make([]Message, len(req.Messages))
	for i, msg := range req.Messages {
		// Convert tool calls
		var llmToolCalls []ToolCall
		for _, tc := range msg.ToolCalls {
			llmToolCalls = append(llmToolCalls, ToolCall{
				ID:   tc.ID,
				Type: tc.Type,
				Function: struct {
					Name      string `json:"name"`
					Arguments string `json:"arguments"`
				}{
					Name:      tc.Function.Name,
					Arguments: tc.Function.Arguments,
				},
			})
		}

		llmMessages[i] = Message{
			Role:      msg.Role,
			Content:   msg.Content,
			ToolCalls: llmToolCalls,
		}
	}
	llmReq.Messages = llmMessages

	// Convert tools if present
	if len(req.Tools) > 0 {
		llmTools := make([]Tool, len(req.Tools))
		for i, tool := range req.Tools {
			llmTools[i] = Tool{
				Type: tool.Type,
				Function: ToolFunction{
					Name:        tool.Function.Name,
					Description: tool.Function.Description,
					Parameters:  tool.Function.Parameters,
				},
			}
		}
		llmReq.Tools = llmTools
	}

	// Call LLM manager
	llmResp, err := a.llmManager.GenerateResponse(ctx, providerName, llmReq)
	if err != nil {
		return nil, err
	}

	// Convert response back to workflow format
	workflowChoices := make([]workflow.Choice, len(llmResp.Choices))
	for i, choice := range llmResp.Choices {
		// Convert tool calls
		var workflowToolCalls []workflow.ToolCall
		for _, tc := range choice.Message.ToolCalls {
			workflowToolCalls = append(workflowToolCalls, workflow.ToolCall{
				ID:   tc.ID,
				Type: tc.Type,
				Function: workflow.Function{
					Name:      tc.Function.Name,
					Arguments: tc.Function.Arguments,
				},
			})
		}

		workflowChoices[i] = workflow.Choice{
			Index: choice.Index,
			Message: workflow.Message{
				Role:      choice.Message.Role,
				Content:   choice.Message.Content,
				ToolCalls: workflowToolCalls,
			},
			FinishReason: choice.FinishReason,
		}
	}

	return &workflow.GenerateResponse{
		ID:      llmResp.ID,
		Object:  llmResp.Object,
		Created: llmResp.Created,
		Model:   llmResp.Model,
		Choices: workflowChoices,
		Usage: workflow.Usage{
			PromptTokens:     llmResp.Usage.PromptTokens,
			CompletionTokens: llmResp.Usage.CompletionTokens,
			TotalTokens:      llmResp.Usage.TotalTokens,
		},
	}, nil
}

func (a *agentLLMManagerAdapter) GetProvider(name string) (workflow.LLMProvider, error) {
	provider, err := a.llmManager.GetProvider(name)
	if err != nil {
		return nil, err
	}

	// Create adapter for the provider
	return &workflowProviderAdapter{provider: provider}, nil
}

func (a *agentLLMManagerAdapter) ListProviders() []string {
	return a.llmManager.ListProviders()
}

func (a *agentLLMManagerAdapter) AddProvider(name string, provider workflow.LLMProvider) error {
	// This method is not directly compatible since workflow.LLMProvider is different from providers.LLMProvider
	// For now, return an error indicating this is not supported via this interface
	return fmt.Errorf("AddProvider not supported via adapter - use LLMManager.AddProvider with LLMConfig instead")
}

func (a *agentLLMManagerAdapter) RemoveProvider(name string) error {
	// Since LLMManager doesn't have RemoveProvider, we'll implement basic functionality
	// This would need to be added to LLMManager for full support
	return fmt.Errorf("RemoveProvider not implemented in LLMManager")
}

func (a *agentLLMManagerAdapter) SetDefaultProvider(name string) error {
	return a.llmManager.SetDefaultProvider(name)
}

func (a *agentLLMManagerAdapter) Close() error {
	return a.llmManager.Close()
}

// workflowProviderAdapter adapts providers.LLMProvider to workflow.LLMProvider
type workflowProviderAdapter struct {
	provider providers.LLMProvider
}

func (a *workflowProviderAdapter) GenerateResponse(ctx context.Context, req *workflow.GenerateRequest) (*workflow.GenerateResponse, error) {
	// Convert workflow request to providers request
	providerReq := &providers.GenerateRequest{
		Model:       req.Model,
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
	}

	// Convert messages
	providerMessages := make([]providers.Message, len(req.Messages))
	for i, msg := range req.Messages {
		// Convert tool calls
		var providerToolCalls []providers.ToolCall
		for _, tc := range msg.ToolCalls {
			providerToolCalls = append(providerToolCalls, providers.ToolCall{
				ID:   tc.ID,
				Type: tc.Type,
				Function: providers.Function{
					Name:      tc.Function.Name,
					Arguments: tc.Function.Arguments,
				},
			})
		}

		providerMessages[i] = providers.Message{
			Role:      msg.Role,
			Content:   msg.Content,
			ToolCalls: providerToolCalls,
		}
	}
	providerReq.Messages = providerMessages

	// Call provider
	providerResp, err := a.provider.GenerateResponse(ctx, providerReq)
	if err != nil {
		return nil, err
	}

	// Convert response back to workflow format
	workflowChoices := make([]workflow.Choice, len(providerResp.Choices))
	for i, choice := range providerResp.Choices {
		// Convert tool calls
		var workflowToolCalls []workflow.ToolCall
		for _, tc := range choice.Message.ToolCalls {
			workflowToolCalls = append(workflowToolCalls, workflow.ToolCall{
				ID:   tc.ID,
				Type: tc.Type,
				Function: workflow.Function{
					Name:      tc.Function.Name,
					Arguments: tc.Function.Arguments,
				},
			})
		}

		workflowChoices[i] = workflow.Choice{
			Index: choice.Index,
			Message: workflow.Message{
				Role:      choice.Message.Role,
				Content:   choice.Message.Content,
				ToolCalls: workflowToolCalls,
			},
			FinishReason: choice.FinishReason,
		}
	}

	return &workflow.GenerateResponse{
		ID:      providerResp.ID,
		Object:  providerResp.Object,
		Created: providerResp.Created,
		Model:   providerResp.Model,
		Choices: workflowChoices,
		Usage: workflow.Usage{
			PromptTokens:     providerResp.Usage.PromptTokens,
			CompletionTokens: providerResp.Usage.CompletionTokens,
			TotalTokens:      providerResp.Usage.TotalTokens,
		},
	}, nil
}

func (a *workflowProviderAdapter) Close() error {
	return a.provider.Close()
}
