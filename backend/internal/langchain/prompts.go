package langchain

import (
	"bytes"
	"context"
	"fmt"
	"regexp"
	"strings"
	"text/template"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// PromptTemplate represents a template for generating prompts
type PromptTemplate struct {
	Name         string            `json:"name"`
	Template     string            `json:"template"`
	InputVars    []string          `json:"input_vars"`
	PartialVars  map[string]string `json:"partial_vars"`
	Metadata     map[string]interface{} `json:"metadata"`
	compiled     *template.Template `json:"-"`
	tracer       trace.Tracer       `json:"-"`
}

// NewPromptTemplate creates a new prompt template
func NewPromptTemplate(name, templateStr string, inputVars []string) *PromptTemplate {
	pt := &PromptTemplate{
		Name:        name,
		Template:    templateStr,
		InputVars:   inputVars,
		PartialVars: make(map[string]string),
		Metadata:    make(map[string]interface{}),
		tracer:      otel.Tracer("langchain.prompt"),
	}

	// Compile template
	tmpl, err := template.New(name).Parse(templateStr)
	if err == nil {
		pt.compiled = tmpl
	}

	return pt
}

// Render renders the template with the given variables
func (pt *PromptTemplate) Render(ctx context.Context, vars map[string]interface{}) (string, error) {
	ctx, span := pt.tracer.Start(ctx, "prompt_template.render")
	defer span.End()

	span.SetAttributes(
		attribute.String("template.name", pt.Name),
		attribute.Int("input_vars.count", len(pt.InputVars)),
	)

	// Merge partial variables
	allVars := make(map[string]interface{})
	for key, value := range pt.PartialVars {
		allVars[key] = value
	}
	for key, value := range vars {
		allVars[key] = value
	}

	// Check for missing variables
	missing := pt.getMissingVars(allVars)
	if len(missing) > 0 {
		err := fmt.Errorf("missing required variables: %v", missing)
		span.RecordError(err)
		return "", err
	}

	// Render template
	if pt.compiled != nil {
		var buf bytes.Buffer
		if err := pt.compiled.Execute(&buf, allVars); err != nil {
			span.RecordError(err)
			return "", fmt.Errorf("template execution failed: %w", err)
		}
		result := buf.String()
		span.SetAttributes(attribute.String("rendered.length", fmt.Sprintf("%d", len(result))))
		return result, nil
	}

	// Fallback to simple string replacement
	result := pt.Template
	for key, value := range allVars {
		placeholder := fmt.Sprintf("{{.%s}}", key)
		result = strings.ReplaceAll(result, placeholder, fmt.Sprintf("%v", value))
	}

	span.SetAttributes(attribute.String("rendered.length", fmt.Sprintf("%d", len(result))))
	return result, nil
}

// getMissingVars returns variables that are required but not provided
func (pt *PromptTemplate) getMissingVars(vars map[string]interface{}) []string {
	missing := []string{}
	for _, required := range pt.InputVars {
		if _, exists := vars[required]; !exists {
			missing = append(missing, required)
		}
	}
	return missing
}

// SetPartial sets partial variables that will be included in all renders
func (pt *PromptTemplate) SetPartial(key, value string) {
	pt.PartialVars[key] = value
}

// Validate validates the prompt template
func (pt *PromptTemplate) Validate() error {
	if pt.Name == "" {
		return fmt.Errorf("template name cannot be empty")
	}

	if pt.Template == "" {
		return fmt.Errorf("template string cannot be empty")
	}

	// Try to compile template to check for syntax errors
	_, err := template.New("validation").Parse(pt.Template)
	if err != nil {
		return fmt.Errorf("template syntax error: %w", err)
	}

	return nil
}

// Clone creates a copy of the prompt template
func (pt *PromptTemplate) Clone() *PromptTemplate {
	clone := &PromptTemplate{
		Name:        pt.Name,
		Template:    pt.Template,
		InputVars:   make([]string, len(pt.InputVars)),
		PartialVars: make(map[string]string),
		Metadata:    make(map[string]interface{}),
		tracer:      pt.tracer,
	}

	copy(clone.InputVars, pt.InputVars)

	for key, value := range pt.PartialVars {
		clone.PartialVars[key] = value
	}

	for key, value := range pt.Metadata {
		clone.Metadata[key] = value
	}

	// Recompile template
	tmpl, err := template.New(clone.Name).Parse(clone.Template)
	if err == nil {
		clone.compiled = tmpl
	}

	return clone
}

// ChatPromptTemplate represents a template for chat-based prompts
type ChatPromptTemplate struct {
	Name      string                 `json:"name"`
	Messages  []MessageTemplate      `json:"messages"`
	InputVars []string               `json:"input_vars"`
	Metadata  map[string]interface{} `json:"metadata"`
	tracer    trace.Tracer           `json:"-"`
}

// MessageTemplate represents a single message in a chat prompt
type MessageTemplate struct {
	Role     string          `json:"role"`     // "system", "user", "assistant"
	Template *PromptTemplate `json:"template"`
}

// NewChatPromptTemplate creates a new chat prompt template
func NewChatPromptTemplate(name string, messages []MessageTemplate) *ChatPromptTemplate {
	// Collect all input variables
	inputVars := make(map[string]bool)
	for _, msg := range messages {
		for _, variable := range msg.Template.InputVars {
			inputVars[variable] = true
		}
	}

	// Convert to slice
	vars := make([]string, 0, len(inputVars))
	for variable := range inputVars {
		vars = append(vars, variable)
	}

	return &ChatPromptTemplate{
		Name:      name,
		Messages:  messages,
		InputVars: vars,
		Metadata:  make(map[string]interface{}),
		tracer:    otel.Tracer("langchain.chat_prompt"),
	}
}

// RenderMessages renders all message templates
func (cpt *ChatPromptTemplate) RenderMessages(ctx context.Context, vars map[string]interface{}) ([]RenderedMessage, error) {
	ctx, span := cpt.tracer.Start(ctx, "chat_prompt_template.render_messages")
	defer span.End()

	span.SetAttributes(
		attribute.String("template.name", cpt.Name),
		attribute.Int("messages.count", len(cpt.Messages)),
	)

	messages := make([]RenderedMessage, len(cpt.Messages))
	for i, msgTemplate := range cpt.Messages {
		content, err := msgTemplate.Template.Render(ctx, vars)
		if err != nil {
			span.RecordError(err)
			return nil, fmt.Errorf("failed to render message %d: %w", i, err)
		}

		messages[i] = RenderedMessage{
			Role:    msgTemplate.Role,
			Content: content,
		}
	}

	return messages, nil
}

// RenderedMessage represents a rendered chat message
type RenderedMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Validate validates the chat prompt template
func (cpt *ChatPromptTemplate) Validate() error {
	if cpt.Name == "" {
		return fmt.Errorf("chat template name cannot be empty")
	}

	if len(cpt.Messages) == 0 {
		return fmt.Errorf("chat template must have at least one message")
	}

	for i, msg := range cpt.Messages {
		if msg.Role == "" {
			return fmt.Errorf("message %d role cannot be empty", i)
		}

		if msg.Template == nil {
			return fmt.Errorf("message %d template cannot be nil", i)
		}

		if err := msg.Template.Validate(); err != nil {
			return fmt.Errorf("message %d template validation failed: %w", i, err)
		}
	}

	return nil
}

// PromptTemplateRegistry manages a collection of prompt templates
type PromptTemplateRegistry struct {
	templates map[string]*PromptTemplate
	chatTemplates map[string]*ChatPromptTemplate
	tracer    trace.Tracer
}

// NewPromptTemplateRegistry creates a new prompt template registry
func NewPromptTemplateRegistry() *PromptTemplateRegistry {
	return &PromptTemplateRegistry{
		templates:     make(map[string]*PromptTemplate),
		chatTemplates: make(map[string]*ChatPromptTemplate),
		tracer:        otel.Tracer("langchain.template_registry"),
	}
}

// RegisterTemplate registers a prompt template
func (ptr *PromptTemplateRegistry) RegisterTemplate(template *PromptTemplate) error {
	if err := template.Validate(); err != nil {
		return fmt.Errorf("template validation failed: %w", err)
	}

	ptr.templates[template.Name] = template
	return nil
}

// RegisterChatTemplate registers a chat prompt template
func (ptr *PromptTemplateRegistry) RegisterChatTemplate(template *ChatPromptTemplate) error {
	if err := template.Validate(); err != nil {
		return fmt.Errorf("chat template validation failed: %w", err)
	}

	ptr.chatTemplates[template.Name] = template
	return nil
}

// GetTemplate retrieves a prompt template by name
func (ptr *PromptTemplateRegistry) GetTemplate(name string) (*PromptTemplate, error) {
	template, exists := ptr.templates[name]
	if !exists {
		return nil, fmt.Errorf("template '%s' not found", name)
	}
	return template.Clone(), nil
}

// GetChatTemplate retrieves a chat prompt template by name
func (ptr *PromptTemplateRegistry) GetChatTemplate(name string) (*ChatPromptTemplate, error) {
	template, exists := ptr.chatTemplates[name]
	if !exists {
		return nil, fmt.Errorf("chat template '%s' not found", name)
	}
	return template, nil
}

// ListTemplates returns all registered template names
func (ptr *PromptTemplateRegistry) ListTemplates() []string {
	names := make([]string, 0, len(ptr.templates))
	for name := range ptr.templates {
		names = append(names, name)
	}
	return names
}

// ListChatTemplates returns all registered chat template names
func (ptr *PromptTemplateRegistry) ListChatTemplates() []string {
	names := make([]string, 0, len(ptr.chatTemplates))
	for name := range ptr.chatTemplates {
		names = append(names, name)
	}
	return names
}

// TravelPromptTemplates provides pre-built templates for travel use cases
type TravelPromptTemplates struct {
	registry *PromptTemplateRegistry
}

// NewTravelPromptTemplates creates travel-specific prompt templates
func NewTravelPromptTemplates() *TravelPromptTemplates {
	registry := NewPromptTemplateRegistry()
	templates := &TravelPromptTemplates{registry: registry}
	templates.registerTravelTemplates()
	return templates
}

// registerTravelTemplates registers all travel-related templates
func (tpt *TravelPromptTemplates) registerTravelTemplates() {
	// Destination research template
	destResearch := NewPromptTemplate(
		"destination_research",
		`Research the travel destination: {{.destination}}

Please provide comprehensive information about:
1. Best time to visit and weather patterns
2. Top attractions and must-see places
3. Local culture, customs, and etiquette
4. Transportation options and getting around
5. Food and dining recommendations
6. Safety considerations and travel tips
7. Estimated costs and budget guidelines

Travel Details:
- Destination: {{.destination}}
- Travel dates: {{.start_date}} to {{.end_date}}
- Number of travelers: {{.travelers}}
- Budget: ${{.budget}}
- Interests: {{.interests}}

Provide detailed, practical information for trip planning.`,
		[]string{"destination", "start_date", "end_date", "travelers", "budget", "interests"},
	)

	// Flight analysis template
	flightAnalysis := NewPromptTemplate(
		"flight_analysis",
		`Analyze the following flight options for travel from {{.origin}} to {{.destination}}:

Flight Options:
{{.flight_options}}

Travel Details:
- Route: {{.origin}} to {{.destination}}
- Departure: {{.start_date}}
- Return: {{.end_date}}
- Travelers: {{.travelers}}
- Budget: ${{.budget}}

Please provide:
1. Price comparison and value analysis
2. Flight duration and routing assessment
3. Airline and service quality comparison
4. Best options for different priorities (price, time, comfort)
5. Booking recommendations and timing advice
6. Alternative dates or routes if beneficial

Format as a clear, actionable analysis.`,
		[]string{"origin", "destination", "flight_options", "start_date", "end_date", "travelers", "budget"},
	)

	// Hotel recommendations template
	hotelRecommendations := NewPromptTemplate(
		"hotel_recommendations",
		`Provide hotel recommendations for {{.destination}}:

Available Hotels:
{{.hotel_options}}

Travel Requirements:
- Destination: {{.destination}}
- Check-in: {{.start_date}}
- Check-out: {{.end_date}}
- Guests: {{.travelers}}
- Budget: ${{.budget}} per night
- Preferences: {{.preferences}}

Please recommend:
1. Top 3 hotels with detailed explanations
2. Best neighborhoods for different needs
3. Value for money analysis
4. Amenities and location considerations
5. Booking tips and strategies
6. Alternative accommodation types if appropriate

Focus on practical, actionable recommendations.`,
		[]string{"destination", "hotel_options", "start_date", "end_date", "travelers", "budget", "preferences"},
	)

	// Itinerary planning template
	itineraryPlanning := NewPromptTemplate(
		"itinerary_planning",
		`Create a detailed {{.duration}}-day itinerary for {{.destination}}:

Available Information:
- Destination info: {{.destination_info}}
- Weather forecast: {{.weather_info}}
- Attractions: {{.attractions}}
- Transportation: {{.transportation_info}}

Trip Details:
- Destination: {{.destination}}
- Dates: {{.start_date}} to {{.end_date}}
- Duration: {{.duration}} days
- Travelers: {{.travelers}}
- Budget: ${{.budget}}
- Interests: {{.interests}}

Create a comprehensive day-by-day itinerary including:
1. Daily schedule with specific times
2. Morning, afternoon, and evening activities
3. Restaurant recommendations for each meal
4. Transportation between locations
5. Estimated costs for each activity
6. Alternative options for different weather
7. Rest periods and flexibility
8. Local tips and cultural insights

Format as a practical, easy-to-follow itinerary.`,
		[]string{"destination", "duration", "start_date", "end_date", "travelers", "budget", "interests", "destination_info", "weather_info", "attractions", "transportation_info"},
	)

	// Register all templates
	tpt.registry.RegisterTemplate(destResearch)
	tpt.registry.RegisterTemplate(flightAnalysis)
	tpt.registry.RegisterTemplate(hotelRecommendations)
	tpt.registry.RegisterTemplate(itineraryPlanning)

	// Chat template for travel assistant
	travelAssistantChat := NewChatPromptTemplate(
		"travel_assistant_chat",
		[]MessageTemplate{
			{
				Role: "system",
				Template: NewPromptTemplate(
					"system_message",
					`You are an expert travel assistant specializing in {{.specialty}}. You provide helpful, accurate, and personalized travel advice based on the user's needs and preferences. Always be friendly, informative, and practical in your responses.

Current context:
- User location: {{.user_location}}
- Travel experience level: {{.experience_level}}
- Budget range: {{.budget_range}}
- Travel style: {{.travel_style}}`,
					[]string{"specialty", "user_location", "experience_level", "budget_range", "travel_style"},
				),
			},
			{
				Role: "user",
				Template: NewPromptTemplate(
					"user_message",
					`{{.user_query}}

Additional context:
{{.additional_context}}`,
					[]string{"user_query", "additional_context"},
				),
			},
		},
	)

	tpt.registry.RegisterChatTemplate(travelAssistantChat)
}

// GetRegistry returns the prompt template registry
func (tpt *TravelPromptTemplates) GetRegistry() *PromptTemplateRegistry {
	return tpt.registry
}

// ExtractVariables extracts variable names from a template string
func ExtractVariables(templateStr string) []string {
	// Regular expression to match {{.variable}} patterns
	re := regexp.MustCompile(`\{\{\.(\w+)\}\}`)
	matches := re.FindAllStringSubmatch(templateStr, -1)
	
	variables := make(map[string]bool)
	for _, match := range matches {
		if len(match) > 1 {
			variables[match[1]] = true
		}
	}
	
	// Convert to slice
	result := make([]string, 0, len(variables))
	for variable := range variables {
		result = append(result, variable)
	}
	
	return result
}
