package prompts

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"text/template"
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

// PromptTemplate represents a template for generating prompts
type PromptTemplate struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Template    string                 `json:"template"`
	Variables   []string               `json:"variables"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	template    *template.Template
	tracer      trace.Tracer
}

// NewPromptTemplate creates a new prompt template
func NewPromptTemplate(name, description, templateStr string) (*PromptTemplate, error) {
	tmpl, err := template.New(name).Parse(templateStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	// Extract variables from template
	variables := extractVariables(templateStr)

	return &PromptTemplate{
		Name:        name,
		Description: description,
		Template:    templateStr,
		Variables:   variables,
		template:    tmpl,
		tracer:      otel.Tracer("llm.prompts.template"),
	}, nil
}

// Render renders the template with the given variables
func (pt *PromptTemplate) Render(ctx context.Context, variables map[string]interface{}) (string, error) {
	ctx, span := pt.tracer.Start(ctx, "prompt_template.render")
	defer span.End()

	span.SetAttributes(
		attribute.String("template.name", pt.Name),
		attribute.Int("template.variables", len(variables)),
	)

	// Add default variables
	enrichedVars := pt.addDefaultVariables(variables)

	// Validate required variables
	if err := pt.validateVariables(enrichedVars); err != nil {
		span.RecordError(err)
		return "", err
	}

	var buf bytes.Buffer
	if err := pt.template.Execute(&buf, enrichedVars); err != nil {
		span.RecordError(err)
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	result := buf.String()
	span.SetAttributes(attribute.Int("template.output_length", len(result)))

	return result, nil
}

// RenderToMessages renders the template and converts to messages
func (pt *PromptTemplate) RenderToMessages(ctx context.Context, variables map[string]interface{}) ([]Message, error) {
	rendered, err := pt.Render(ctx, variables)
	if err != nil {
		return nil, err
	}

	// Parse the rendered template into messages
	// This supports a simple format like:
	// SYSTEM: system message
	// USER: user message
	// ASSISTANT: assistant message

	messages := make([]Message, 0)
	lines := strings.Split(rendered, "\n")

	var currentRole string
	var currentContent strings.Builder

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Check if line starts with a role indicator
		if strings.HasPrefix(line, "SYSTEM:") {
			if currentRole != "" {
				messages = append(messages, Message{
					Role:    strings.ToLower(currentRole),
					Content: strings.TrimSpace(currentContent.String()),
				})
			}
			currentRole = "SYSTEM"
			currentContent.Reset()
			currentContent.WriteString(strings.TrimSpace(line[7:]))
		} else if strings.HasPrefix(line, "USER:") {
			if currentRole != "" {
				messages = append(messages, Message{
					Role:    strings.ToLower(currentRole),
					Content: strings.TrimSpace(currentContent.String()),
				})
			}
			currentRole = "USER"
			currentContent.Reset()
			currentContent.WriteString(strings.TrimSpace(line[5:]))
		} else if strings.HasPrefix(line, "ASSISTANT:") {
			if currentRole != "" {
				messages = append(messages, Message{
					Role:    strings.ToLower(currentRole),
					Content: strings.TrimSpace(currentContent.String()),
				})
			}
			currentRole = "ASSISTANT"
			currentContent.Reset()
			currentContent.WriteString(strings.TrimSpace(line[10:]))
		} else {
			// Continue current message content
			if currentContent.Len() > 0 {
				currentContent.WriteString("\n")
			}
			currentContent.WriteString(line)
		}
	}

	// Add the last message
	if currentRole != "" {
		messages = append(messages, Message{
			Role:    strings.ToLower(currentRole),
			Content: strings.TrimSpace(currentContent.String()),
		})
	}

	// If no role indicators found, treat as single user message
	if len(messages) == 0 && rendered != "" {
		messages = append(messages, Message{
			Role:    "user",
			Content: rendered,
		})
	}

	return messages, nil
}

// addDefaultVariables adds default variables like timestamp, date, etc.
func (pt *PromptTemplate) addDefaultVariables(variables map[string]interface{}) map[string]interface{} {
	enriched := make(map[string]interface{})

	// Copy provided variables
	for k, v := range variables {
		enriched[k] = v
	}

	// Add default variables if not already provided
	now := time.Now()

	if _, exists := enriched["timestamp"]; !exists {
		enriched["timestamp"] = now.Format(time.RFC3339)
	}

	if _, exists := enriched["date"]; !exists {
		enriched["date"] = now.Format("2006-01-02")
	}

	if _, exists := enriched["time"]; !exists {
		enriched["time"] = now.Format("15:04:05")
	}

	if _, exists := enriched["datetime"]; !exists {
		enriched["datetime"] = now.Format("2006-01-02 15:04:05")
	}

	return enriched
}

// validateVariables checks if all required variables are provided
func (pt *PromptTemplate) validateVariables(variables map[string]interface{}) error {
	missing := make([]string, 0)

	for _, required := range pt.Variables {
		if _, exists := variables[required]; !exists {
			missing = append(missing, required)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required variables: %s", strings.Join(missing, ", "))
	}

	return nil
}

// extractVariables extracts variable names from template string
func extractVariables(templateStr string) []string {
	variables := make([]string, 0)
	variableSet := make(map[string]bool)

	// Simple regex-like extraction for {{.Variable}} patterns
	start := 0
	for {
		openIndex := strings.Index(templateStr[start:], "{{")
		if openIndex == -1 {
			break
		}
		openIndex += start

		closeIndex := strings.Index(templateStr[openIndex:], "}}")
		if closeIndex == -1 {
			break
		}
		closeIndex += openIndex

		// Extract variable name
		varExpr := templateStr[openIndex+2 : closeIndex]
		varExpr = strings.TrimSpace(varExpr)

		// Handle .Variable format
		if strings.HasPrefix(varExpr, ".") {
			varName := varExpr[1:]
			if !variableSet[varName] {
				variables = append(variables, varName)
				variableSet[varName] = true
			}
		}

		start = closeIndex + 2
	}

	return variables
}

// PromptManager manages a collection of prompt templates
type PromptManager struct {
	templates map[string]*PromptTemplate
	tracer    trace.Tracer
}

// NewPromptManager creates a new prompt manager
func NewPromptManager() *PromptManager {
	return &PromptManager{
		templates: make(map[string]*PromptTemplate),
		tracer:    otel.Tracer("llm.prompts.manager"),
	}
}

// AddTemplate adds a template to the manager
func (pm *PromptManager) AddTemplate(template *PromptTemplate) error {
	if template.Name == "" {
		return fmt.Errorf("template name cannot be empty")
	}

	pm.templates[template.Name] = template
	return nil
}

// GetTemplate retrieves a template by name
func (pm *PromptManager) GetTemplate(name string) (*PromptTemplate, error) {
	template, exists := pm.templates[name]
	if !exists {
		return nil, fmt.Errorf("template not found: %s", name)
	}

	return template, nil
}

// RenderTemplate renders a template by name
func (pm *PromptManager) RenderTemplate(ctx context.Context, name string, variables map[string]interface{}) (string, error) {
	ctx, span := pm.tracer.Start(ctx, "prompt_manager.render_template")
	defer span.End()

	span.SetAttributes(attribute.String("template.name", name))

	template, err := pm.GetTemplate(name)
	if err != nil {
		span.RecordError(err)
		return "", err
	}

	return template.Render(ctx, variables)
}

// RenderToMessages renders a template to messages by name
func (pm *PromptManager) RenderToMessages(ctx context.Context, name string, variables map[string]interface{}) ([]Message, error) {
	template, err := pm.GetTemplate(name)
	if err != nil {
		return nil, err
	}

	return template.RenderToMessages(ctx, variables)
}

// ListTemplates returns all template names
func (pm *PromptManager) ListTemplates() []string {
	names := make([]string, 0, len(pm.templates))
	for name := range pm.templates {
		names = append(names, name)
	}
	return names
}

// RemoveTemplate removes a template
func (pm *PromptManager) RemoveTemplate(name string) error {
	if _, exists := pm.templates[name]; !exists {
		return fmt.Errorf("template not found: %s", name)
	}

	delete(pm.templates, name)
	return nil
}

// LoadFromMap loads templates from a map
func (pm *PromptManager) LoadFromMap(templates map[string]string) error {
	for name, templateStr := range templates {
		template, err := NewPromptTemplate(name, "", templateStr)
		if err != nil {
			return fmt.Errorf("failed to create template %s: %w", name, err)
		}

		if err := pm.AddTemplate(template); err != nil {
			return fmt.Errorf("failed to add template %s: %w", name, err)
		}
	}

	return nil
}

// ChainableTemplate represents a template that can be chained with others
type ChainableTemplate struct {
	*PromptTemplate
	nextTemplate *ChainableTemplate
}

// NewChainableTemplate creates a new chainable template
func NewChainableTemplate(name, description, templateStr string) (*ChainableTemplate, error) {
	base, err := NewPromptTemplate(name, description, templateStr)
	if err != nil {
		return nil, err
	}

	return &ChainableTemplate{
		PromptTemplate: base,
	}, nil
}

// Chain chains this template with another
func (ct *ChainableTemplate) Chain(next *ChainableTemplate) *ChainableTemplate {
	ct.nextTemplate = next
	return ct
}

// RenderChain renders this template and all chained templates
func (ct *ChainableTemplate) RenderChain(ctx context.Context, variables map[string]interface{}) ([]string, error) {
	results := make([]string, 0)

	current := ct
	for current != nil {
		rendered, err := current.Render(ctx, variables)
		if err != nil {
			return nil, fmt.Errorf("failed to render template %s: %w", current.Name, err)
		}

		results = append(results, rendered)

		// Add the result as a variable for the next template
		variables["previous_result"] = rendered

		current = current.nextTemplate
	}

	return results, nil
}
