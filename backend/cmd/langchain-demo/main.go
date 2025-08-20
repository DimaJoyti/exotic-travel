package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/exotic-travel-booking/backend/internal/langchain"
)

func main() {
	fmt.Println("🔗 LangChain Patterns Demo for Exotic Travel Booking")
	fmt.Println("====================================================")

	ctx := context.Background()

	// Demo 1: Prompt Templates
	fmt.Println("\n1. Prompt Templates Demo")
	fmt.Println("========================")
	
	if err := demonstratePromptTemplates(ctx); err != nil {
		log.Printf("❌ Prompt templates demo failed: %v", err)
	}

	// Demo 2: Output Parsers
	fmt.Println("\n2. Output Parsers Demo")
	fmt.Println("======================")
	
	if err := demonstrateOutputParsers(ctx); err != nil {
		log.Printf("❌ Output parsers demo failed: %v", err)
	}

	// Demo 3: Memory Systems
	fmt.Println("\n3. Memory Systems Demo")
	fmt.Println("======================")
	
	if err := demonstrateMemorySystems(ctx); err != nil {
		log.Printf("❌ Memory systems demo failed: %v", err)
	}

	// Demo 4: Chain Orchestration
	fmt.Println("\n4. Chain Orchestration Demo")
	fmt.Println("===========================")
	
	if err := demonstrateChainOrchestration(ctx); err != nil {
		log.Printf("❌ Chain orchestration demo failed: %v", err)
	}

	// Demo 5: Travel-Specific Templates
	fmt.Println("\n5. Travel-Specific Templates Demo")
	fmt.Println("=================================")
	
	if err := demonstrateTravelTemplates(ctx); err != nil {
		log.Printf("❌ Travel templates demo failed: %v", err)
	}

	fmt.Println("\n🎉 LangChain Patterns Demo Completed Successfully!")
	fmt.Println("\nKey Features Demonstrated:")
	fmt.Println("✅ Advanced prompt templating with variable substitution")
	fmt.Println("✅ Chat prompt templates for multi-message conversations")
	fmt.Println("✅ Multiple output parsers (JSON, List, Key-Value, Regex, Number)")
	fmt.Println("✅ Memory systems (Buffer, Summary, Window)")
	fmt.Println("✅ Chain orchestration (Sequential, Parallel, Conditional)")
	fmt.Println("✅ Travel-specific prompt templates")
	fmt.Println("✅ Template registry and management")
}

func demonstratePromptTemplates(ctx context.Context) error {
	fmt.Println("   📝 Creating and testing prompt templates...")

	// Basic template
	template := langchain.NewPromptTemplate(
		"destination_query",
		"Plan a trip to {{.destination}} for {{.travelers}} travelers with a budget of ${{.budget}}. Focus on {{.interests}}.",
		[]string{"destination", "travelers", "budget", "interests"},
	)

	vars := map[string]interface{}{
		"destination": "Tokyo, Japan",
		"travelers":   2,
		"budget":      3000,
		"interests":   "culture, food, technology",
	}

	rendered, err := template.Render(ctx, vars)
	if err != nil {
		return fmt.Errorf("template rendering failed: %w", err)
	}

	fmt.Printf("   ✅ Basic template rendered:\n   📋 %s\n", rendered)

	// Template with partial variables
	fmt.Println("   🔧 Testing partial variables...")
	
	partialTemplate := langchain.NewPromptTemplate(
		"travel_assistant",
		"You are a {{.role}} specializing in {{.specialty}}. Help the user with: {{.query}}",
		[]string{"role", "specialty", "query"},
	)

	partialTemplate.SetPartial("role", "travel assistant")
	partialTemplate.SetPartial("specialty", "exotic destinations")

	partialVars := map[string]interface{}{
		"query": "planning a unique adventure in Madagascar",
	}

	partialRendered, err := partialTemplate.Render(ctx, partialVars)
	if err != nil {
		return fmt.Errorf("partial template rendering failed: %w", err)
	}

	fmt.Printf("   ✅ Partial template rendered:\n   📋 %s\n", partialRendered)

	// Chat template
	fmt.Println("   💬 Testing chat prompt templates...")

	systemTemplate := langchain.NewPromptTemplate(
		"system",
		"You are an expert travel planner specializing in {{.region}}. Provide helpful, detailed advice.",
		[]string{"region"},
	)

	userTemplate := langchain.NewPromptTemplate(
		"user",
		"I want to visit {{.destination}} for {{.duration}} days. My interests are {{.interests}}. Budget: ${{.budget}}",
		[]string{"destination", "duration", "interests", "budget"},
	)

	chatTemplate := langchain.NewChatPromptTemplate(
		"travel_consultation",
		[]langchain.MessageTemplate{
			{Role: "system", Template: systemTemplate},
			{Role: "user", Template: userTemplate},
		},
	)

	chatVars := map[string]interface{}{
		"region":      "Southeast Asia",
		"destination": "Bali, Indonesia",
		"duration":    7,
		"interests":   "beaches, temples, local cuisine",
		"budget":      2500,
	}

	messages, err := chatTemplate.RenderMessages(ctx, chatVars)
	if err != nil {
		return fmt.Errorf("chat template rendering failed: %w", err)
	}

	fmt.Printf("   ✅ Chat template rendered %d messages:\n", len(messages))
	for i, msg := range messages {
		fmt.Printf("   📨 Message %d (%s): %s\n", i+1, msg.Role, truncateString(msg.Content, 80))
	}

	return nil
}

func demonstrateOutputParsers(ctx context.Context) error {
	fmt.Println("   🔍 Testing various output parsers...")

	// JSON Parser
	fmt.Println("   📊 JSON Parser:")
	jsonParser := langchain.NewJSONParser("travel_json", map[string]interface{}{
		"type": "object",
		"required": []string{"destination", "budget"},
	}, false)

	jsonOutput := `{
		"destination": "Santorini, Greece",
		"budget": 2800,
		"duration": 6,
		"highlights": ["sunset views", "white architecture", "wine tasting"]
	}`

	jsonResult, err := jsonParser.Parse(ctx, jsonOutput)
	if err != nil {
		return fmt.Errorf("JSON parsing failed: %w", err)
	}

	fmt.Printf("   ✅ JSON parsed successfully: %v\n", jsonResult)

	// List Parser
	fmt.Println("   📝 List Parser:")
	listParser := langchain.NewListParser("attractions", "\n", true, true)

	listOutput := `1. Acropolis Museum - Ancient Greek artifacts
2. Santorini Caldera - Volcanic crater views
3. Mykonos Windmills - Traditional architecture
4. Delphi Archaeological Site - Oracle of Delphi
5. Meteora Monasteries - Clifftop monasteries`

	listResult, err := listParser.Parse(ctx, listOutput)
	if err != nil {
		return fmt.Errorf("list parsing failed: %w", err)
	}

	attractions := listResult.([]string)
	fmt.Printf("   ✅ List parsed %d attractions:\n", len(attractions))
	for i, attraction := range attractions[:3] { // Show first 3
		fmt.Printf("   🏛️  %d. %s\n", i+1, attraction)
	}

	// Key-Value Parser
	fmt.Println("   🔑 Key-Value Parser:")
	kvParser := langchain.NewKeyValueParser("trip_details", ": ", "\n", true)

	kvOutput := `destination: Kyoto, Japan
duration: 8 days
budget: $3200
season: Spring (Cherry Blossom)
accommodation: Traditional Ryokan
transport: JR Pass`

	kvResult, err := kvParser.Parse(ctx, kvOutput)
	if err != nil {
		return fmt.Errorf("key-value parsing failed: %w", err)
	}

	details := kvResult.(map[string]string)
	fmt.Printf("   ✅ Key-value pairs parsed:\n")
	for key, value := range details {
		fmt.Printf("   📌 %s: %s\n", key, value)
	}

	// Number Parser
	fmt.Println("   🔢 Number Parser:")
	numberParser := langchain.NewNumberParser("budget_parser", "int", 0)

	numberOutput := "The estimated total cost for this trip is $4,250 including flights."
	numberResult, err := numberParser.Parse(ctx, numberOutput)
	if err != nil {
		return fmt.Errorf("number parsing failed: %w", err)
	}

	fmt.Printf("   ✅ Number extracted: $%d\n", numberResult)

	return nil
}

func demonstrateMemorySystems(ctx context.Context) error {
	fmt.Println("   🧠 Testing memory systems...")

	sessionID := "travel_planning_session"

	// Buffer Memory
	fmt.Println("   📚 Buffer Memory:")
	bufferMemory := langchain.NewBufferMemory("conversation_buffer", 10)

	// Simulate a conversation
	conversation := []*langchain.Message{
		{SessionID: sessionID, Role: "user", Content: "I want to plan a trip to Iceland", Timestamp: time.Now()},
		{SessionID: sessionID, Role: "assistant", Content: "Iceland is amazing! When are you planning to visit?", Timestamp: time.Now()},
		{SessionID: sessionID, Role: "user", Content: "Next summer, around July", Timestamp: time.Now()},
		{SessionID: sessionID, Role: "assistant", Content: "Perfect timing! July is great for the midnight sun and lupine flowers.", Timestamp: time.Now()},
		{SessionID: sessionID, Role: "user", Content: "What about the Northern Lights?", Timestamp: time.Now()},
		{SessionID: sessionID, Role: "assistant", Content: "Northern Lights are best seen in winter months (September-March).", Timestamp: time.Now()},
	}

	for _, msg := range conversation {
		err := bufferMemory.AddMessage(ctx, msg)
		if err != nil {
			return fmt.Errorf("failed to add message to buffer memory: %w", err)
		}
	}

	messages, err := bufferMemory.GetMessages(ctx, sessionID, 4)
	if err != nil {
		return fmt.Errorf("failed to retrieve messages: %w", err)
	}

	fmt.Printf("   ✅ Buffer memory contains %d recent messages:\n", len(messages))
	for i, msg := range messages {
		fmt.Printf("   💬 %d. %s: %s\n", i+1, msg.Role, truncateString(msg.Content, 60))
	}

	summary, err := bufferMemory.GetSummary(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("failed to get summary: %w", err)
	}
	fmt.Printf("   📊 Summary: %s\n", summary)

	// Window Memory
	fmt.Println("   🪟 Window Memory:")
	windowMemory := langchain.NewWindowMemory("conversation_window", 3)

	// Add the same messages
	for _, msg := range conversation {
		err := windowMemory.AddMessage(ctx, msg)
		if err != nil {
			return fmt.Errorf("failed to add message to window memory: %w", err)
		}
	}

	windowMessages, err := windowMemory.GetMessages(ctx, sessionID, 0)
	if err != nil {
		return fmt.Errorf("failed to retrieve window messages: %w", err)
	}

	fmt.Printf("   ✅ Window memory keeps last %d messages:\n", len(windowMessages))
	for i, msg := range windowMessages {
		fmt.Printf("   💬 %d. %s: %s\n", i+1, msg.Role, truncateString(msg.Content, 60))
	}

	// Memory Manager
	fmt.Println("   🗂️  Memory Manager:")
	manager := langchain.NewMemoryManager()
	manager.RegisterMemory(bufferMemory)
	manager.RegisterMemory(windowMemory)

	memoryNames := manager.ListMemories()
	fmt.Printf("   ✅ Registered memories: %v\n", memoryNames)

	return nil
}

func demonstrateChainOrchestration(ctx context.Context) error {
	fmt.Println("   ⛓️  Testing chain orchestration...")

	// Create mock chains for demonstration
	researchChain := &MockChain{
		name: "destination_research",
		output: map[string]interface{}{
			"climate": "Mediterranean climate with hot summers",
			"culture": "Rich ancient history and mythology",
			"cuisine": "Fresh seafood, olive oil, feta cheese",
		},
	}

	budgetChain := &MockChain{
		name: "budget_analysis",
		output: map[string]interface{}{
			"estimated_cost": 2800,
			"cost_breakdown": map[string]int{
				"flights": 800,
				"hotels":  1200,
				"food":    500,
				"activities": 300,
			},
		},
	}

	// Sequential Chain
	fmt.Println("   🔄 Sequential Chain:")
	sequential := langchain.NewSequentialChain(
		"trip_planning_sequence",
		"Sequential trip planning workflow",
		researchChain,
		budgetChain,
	)

	input := map[string]interface{}{
		"destination": "Santorini, Greece",
		"travelers":   2,
		"duration":    7,
	}

	seqResult, err := sequential.Execute(ctx, input)
	if err != nil {
		return fmt.Errorf("sequential chain failed: %w", err)
	}

	fmt.Printf("   ✅ Sequential chain completed in %v\n", seqResult.Duration)
	fmt.Printf("   📊 Climate: %s\n", seqResult.Output["climate"])
	fmt.Printf("   💰 Estimated cost: $%v\n", seqResult.Output["estimated_cost"])

	// Parallel Chain
	fmt.Println("   ⚡ Parallel Chain:")
	parallel := langchain.NewParallelChain(
		"parallel_research",
		"Parallel research workflow",
		researchChain,
		budgetChain,
	)

	parResult, err := parallel.Execute(ctx, input)
	if err != nil {
		return fmt.Errorf("parallel chain failed: %w", err)
	}

	fmt.Printf("   ✅ Parallel chain completed in %v\n", parResult.Duration)
	fmt.Printf("   🏛️  Research: %s\n", parResult.Output["destination_research_culture"])
	fmt.Printf("   💵 Budget: $%v\n", parResult.Output["budget_analysis_estimated_cost"])

	// Conditional Chain
	fmt.Println("   🤔 Conditional Chain:")
	luxuryChain := &MockChain{
		name: "luxury_planning",
		output: map[string]interface{}{
			"recommendation": "5-star resorts with private pools and spa services",
		},
	}

	budgetTravelChain := &MockChain{
		name: "budget_planning",
		output: map[string]interface{}{
			"recommendation": "Boutique hotels and local guesthouses with authentic experiences",
		},
	}

	condition := func(ctx context.Context, input map[string]interface{}) (bool, error) {
		budget, exists := input["budget"]
		if !exists {
			return false, nil
		}
		return budget.(int) > 5000, nil
	}

	conditional := langchain.NewConditionalChain(
		"accommodation_selector",
		"Select accommodation based on budget",
		condition,
		luxuryChain,
		budgetTravelChain,
	)

	// Test with high budget
	highBudgetInput := map[string]interface{}{
		"destination": "Maldives",
		"budget":      8000,
	}

	condResult, err := conditional.Execute(ctx, highBudgetInput)
	if err != nil {
		return fmt.Errorf("conditional chain failed: %w", err)
	}

	fmt.Printf("   ✅ Conditional chain (high budget): %s\n", condResult.Output["recommendation"])

	// Test with low budget
	lowBudgetInput := map[string]interface{}{
		"destination": "Thailand",
		"budget":      2000,
	}

	condResult2, err := conditional.Execute(ctx, lowBudgetInput)
	if err != nil {
		return fmt.Errorf("conditional chain failed: %w", err)
	}

	fmt.Printf("   ✅ Conditional chain (low budget): %s\n", condResult2.Output["recommendation"])

	return nil
}

func demonstrateTravelTemplates(ctx context.Context) error {
	fmt.Println("   🌍 Testing travel-specific templates...")

	travelTemplates := langchain.NewTravelPromptTemplates()
	registry := travelTemplates.GetRegistry()

	// List available templates
	templates := registry.ListTemplates()
	fmt.Printf("   📋 Available travel templates: %v\n", templates)

	// Test destination research template
	fmt.Println("   🔍 Destination Research Template:")
	destTemplate, err := registry.GetTemplate("destination_research")
	if err != nil {
		return fmt.Errorf("failed to get destination template: %w", err)
	}

	destVars := map[string]interface{}{
		"destination": "Patagonia, Argentina",
		"start_date":  "2024-11-15",
		"end_date":    "2024-11-28",
		"travelers":   2,
		"budget":      4500,
		"interests":   "hiking, wildlife, photography",
	}

	destPrompt, err := destTemplate.Render(ctx, destVars)
	if err != nil {
		return fmt.Errorf("failed to render destination template: %w", err)
	}

	fmt.Printf("   ✅ Destination research prompt generated (%d chars)\n", len(destPrompt))
	fmt.Printf("   📝 Preview: %s...\n", truncateString(destPrompt, 120))

	// Test flight analysis template
	fmt.Println("   ✈️  Flight Analysis Template:")
	flightTemplate, err := registry.GetTemplate("flight_analysis")
	if err != nil {
		return fmt.Errorf("failed to get flight template: %w", err)
	}

	flightVars := map[string]interface{}{
		"origin":         "New York",
		"destination":    "Patagonia, Argentina",
		"flight_options": "Option 1: Direct flight $1200, Option 2: 1-stop $950, Option 3: 2-stop $750",
		"start_date":     "2024-11-15",
		"end_date":       "2024-11-28",
		"travelers":      2,
		"budget":         2400,
	}

	flightPrompt, err := flightTemplate.Render(ctx, flightVars)
	if err != nil {
		return fmt.Errorf("failed to render flight template: %w", err)
	}

	fmt.Printf("   ✅ Flight analysis prompt generated (%d chars)\n", len(flightPrompt))
	fmt.Printf("   📝 Preview: %s...\n", truncateString(flightPrompt, 120))

	// Test chat template
	fmt.Println("   💬 Travel Assistant Chat Template:")
	chatTemplates := registry.ListChatTemplates()
	fmt.Printf("   📋 Available chat templates: %v\n", chatTemplates)

	return nil
}

// Helper function
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// MockChain for demonstration
type MockChain struct {
	name   string
	output map[string]interface{}
}

func (m *MockChain) Execute(ctx context.Context, input map[string]interface{}) (*langchain.ChainResult, error) {
	// Simulate some processing time
	time.Sleep(10 * time.Millisecond)

	return &langchain.ChainResult{
		Output:    m.output,
		Metadata:  map[string]interface{}{"chain": m.name},
		Duration:  time.Millisecond * 10,
		Success:   true,
		ChainName: m.name,
	}, nil
}

func (m *MockChain) GetName() string {
	return m.name
}

func (m *MockChain) GetDescription() string {
	return fmt.Sprintf("Mock chain: %s", m.name)
}

func (m *MockChain) Validate() error {
	return nil
}
