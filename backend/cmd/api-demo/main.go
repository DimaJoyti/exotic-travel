package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

const baseURL = "http://localhost:8080"

func main() {
	fmt.Println("🌐 AI API Endpoints Demo")
	fmt.Println("========================")
	fmt.Println("Make sure the AI server is running on port 8080")
	fmt.Println("Run: go run cmd/ai-server/main.go")
	fmt.Println()

	// Wait a moment for user to start server
	fmt.Println("⏳ Waiting 3 seconds for server to be ready...")
	time.Sleep(3 * time.Second)

	// Demo 1: Health Check
	fmt.Println("1. Health Check Demo")
	fmt.Println("===================")
	if err := demoHealthCheck(); err != nil {
		log.Printf("❌ Health check failed: %v", err)
	}

	// Demo 2: Knowledge Base Query
	fmt.Println("\n2. Knowledge Base Query Demo")
	fmt.Println("============================")
	if err := demoKnowledgeBaseQuery(); err != nil {
		log.Printf("❌ Knowledge base query failed: %v", err)
	}

	// Demo 3: AI Chat
	fmt.Println("\n3. AI Chat Demo")
	fmt.Println("===============")
	if err := demoChatInteraction(); err != nil {
		log.Printf("❌ Chat interaction failed: %v", err)
	}

	// Demo 4: Specialist Agents
	fmt.Println("\n4. Specialist Agents Demo")
	fmt.Println("=========================")
	if err := demoSpecialistAgents(); err != nil {
		log.Printf("❌ Specialist agents demo failed: %v", err)
	}

	// Demo 5: Conversation Management
	fmt.Println("\n5. Conversation Management Demo")
	fmt.Println("===============================")
	if err := demoConversationManagement(); err != nil {
		log.Printf("❌ Conversation management failed: %v", err)
	}

	fmt.Println("\n🎉 API Demo Completed!")
	fmt.Println("\n📋 Summary of Available Endpoints:")
	fmt.Println("   ✅ Health checks and monitoring")
	fmt.Println("   ✅ AI-powered chat with RAG")
	fmt.Println("   ✅ Knowledge base queries")
	fmt.Println("   ✅ Specialist travel agents")
	fmt.Println("   ✅ Conversation history management")
}

func demoHealthCheck() error {
	fmt.Println("   🔍 Checking API health...")

	// Check main health endpoint
	resp, err := http.Get(baseURL + "/api/v1/health")
	if err != nil {
		return fmt.Errorf("health check request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read health response: %w", err)
	}

	var healthResp map[string]interface{}
	if err := json.Unmarshal(body, &healthResp); err != nil {
		return fmt.Errorf("failed to parse health response: %w", err)
	}

	fmt.Printf("   ✅ Health Status: %s\n", healthResp["status"])
	fmt.Printf("   📊 Services: %v\n", len(healthResp["services"].(map[string]interface{})))

	// Check readiness
	resp, err = http.Get(baseURL + "/api/v1/health/ready")
	if err != nil {
		return fmt.Errorf("readiness check failed: %w", err)
	}
	defer resp.Body.Close()

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read readiness response: %w", err)
	}

	var readyResp map[string]interface{}
	if err := json.Unmarshal(body, &readyResp); err != nil {
		return fmt.Errorf("failed to parse readiness response: %w", err)
	}

	fmt.Printf("   ✅ Ready: %v\n", readyResp["ready"])

	return nil
}

func demoKnowledgeBaseQuery() error {
	fmt.Println("   🔍 Querying travel knowledge base...")

	queries := []struct {
		name        string
		query       string
		destination string
		category    string
	}{
		{
			name:  "General Paris Query",
			query: "What can you tell me about Paris?",
		},
		{
			name:        "Destination-Specific Query",
			query:       "What are the best attractions?",
			destination: "Tokyo",
		},
		{
			name:     "Category-Based Query",
			query:    "Tell me about romantic destinations",
			category: "romance",
		},
	}

	for _, q := range queries {
		fmt.Printf("   📝 %s: %s\n", q.name, q.query)

		reqBody := map[string]interface{}{
			"query": q.query,
		}

		if q.destination != "" {
			reqBody["destination"] = q.destination
		}
		if q.category != "" {
			reqBody["category"] = q.category
		}

		jsonBody, _ := json.Marshal(reqBody)
		resp, err := http.Post(baseURL+"/api/v1/ai/knowledge/query", "application/json", bytes.NewBuffer(jsonBody))
		if err != nil {
			return fmt.Errorf("knowledge query failed: %w", err)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read response: %w", err)
		}

		var result map[string]interface{}
		if err := json.Unmarshal(body, &result); err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}

		answer := result["answer"].(string)
		sources := result["sources"].([]interface{})
		duration := result["duration"].(float64)

		fmt.Printf("   💡 Answer: %s\n", truncateString(answer, 100))
		fmt.Printf("   📚 Sources: %d documents, Duration: %.2fms\n", len(sources), duration/1000000)
		fmt.Println()
	}

	return nil
}

func demoChatInteraction() error {
	fmt.Println("   💬 Testing AI chat interactions...")

	sessionID := fmt.Sprintf("demo_session_%d", time.Now().Unix())
	
	chatMessages := []string{
		"Hello! I'm planning a trip to Japan. Can you help me?",
		"What's the best time to visit Tokyo?",
		"What about food recommendations?",
		"How much should I budget for a week in Japan?",
	}

	for i, message := range chatMessages {
		fmt.Printf("   👤 User: %s\n", message)

		reqBody := map[string]interface{}{
			"message":    message,
			"session_id": sessionID,
			"user_id":    "demo_user",
			"stream":     false,
		}

		jsonBody, _ := json.Marshal(reqBody)
		resp, err := http.Post(baseURL+"/api/v1/ai/chat", "application/json", bytes.NewBuffer(jsonBody))
		if err != nil {
			return fmt.Errorf("chat request failed: %w", err)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read chat response: %w", err)
		}

		var chatResp map[string]interface{}
		if err := json.Unmarshal(body, &chatResp); err != nil {
			return fmt.Errorf("failed to parse chat response: %w", err)
		}

		response := chatResp["response"].(string)
		tokensUsed := chatResp["tokens_used"].(float64)
		duration := chatResp["duration"].(float64)

		fmt.Printf("   🤖 Assistant: %s\n", truncateString(response, 150))
		fmt.Printf("   📊 Tokens: %.0f, Duration: %.2fms\n", tokensUsed, duration/1000000)
		
		if i < len(chatMessages)-1 {
			fmt.Println()
			time.Sleep(1 * time.Second) // Brief pause between messages
		}
	}

	return nil
}

func demoSpecialistAgents() error {
	fmt.Println("   🎯 Testing specialist travel agents...")

	agentRequests := []struct {
		agentType  string
		query      string
		parameters map[string]interface{}
	}{
		{
			agentType: "flight",
			query:     "Find flights from New York to Tokyo",
			parameters: map[string]interface{}{
				"origin":      "New York",
				"destination": "Tokyo",
				"start_date":  "2024-08-15",
				"travelers":   2,
				"budget":      1500,
			},
		},
		{
			agentType: "hotel",
			query:     "Find hotels in Paris",
			parameters: map[string]interface{}{
				"destination": "Paris",
				"start_date":  "2024-09-01",
				"end_date":    "2024-09-05",
				"travelers":   2,
				"budget":      200,
			},
		},
		{
			agentType: "itinerary",
			query:     "Plan a 5-day itinerary for Rome",
			parameters: map[string]interface{}{
				"destination": "Rome",
				"duration":    5,
				"interests":   []string{"history", "food", "art"},
				"budget":      1000,
			},
		},
	}

	sessionID := fmt.Sprintf("agent_demo_%d", time.Now().Unix())

	for _, req := range agentRequests {
		fmt.Printf("   🎯 %s Agent: %s\n", req.agentType, req.query)

		reqBody := map[string]interface{}{
			"agent_type": req.agentType,
			"query":      req.query,
			"session_id": sessionID,
			"user_id":    "demo_user",
			"parameters": req.parameters,
		}

		jsonBody, _ := json.Marshal(reqBody)
		resp, err := http.Post(baseURL+"/api/v1/ai/agents/request", "application/json", bytes.NewBuffer(jsonBody))
		if err != nil {
			return fmt.Errorf("agent request failed: %w", err)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read agent response: %w", err)
		}

		var agentResp map[string]interface{}
		if err := json.Unmarshal(body, &agentResp); err != nil {
			return fmt.Errorf("failed to parse agent response: %w", err)
		}

		response := agentResp["response"].(string)
		confidence := agentResp["confidence"].(float64)
		duration := agentResp["duration"].(float64)

		fmt.Printf("   ✅ Response: %s\n", truncateString(response, 120))
		fmt.Printf("   📊 Confidence: %.2f, Duration: %.2fms\n", confidence, duration/1000000)
		fmt.Println()
	}

	return nil
}

func demoConversationManagement() error {
	fmt.Println("   📚 Testing conversation management...")

	sessionID := fmt.Sprintf("conv_demo_%d", time.Now().Unix())

	// First, have a short conversation
	fmt.Println("   💬 Creating conversation history...")
	messages := []string{
		"I want to visit Iceland",
		"What's the best time to see Northern Lights?",
		"How much should I budget for the trip?",
	}

	for _, message := range messages {
		reqBody := map[string]interface{}{
			"message":    message,
			"session_id": sessionID,
			"user_id":    "demo_user",
		}

		jsonBody, _ := json.Marshal(reqBody)
		http.Post(baseURL+"/api/v1/ai/chat", "application/json", bytes.NewBuffer(jsonBody))
		time.Sleep(500 * time.Millisecond)
	}

	// Get conversation history
	fmt.Println("   📖 Retrieving conversation history...")
	resp, err := http.Get(baseURL + "/api/v1/ai/chat/history/" + sessionID)
	if err != nil {
		return fmt.Errorf("failed to get conversation history: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read history response: %w", err)
	}

	var historyResp map[string]interface{}
	if err := json.Unmarshal(body, &historyResp); err != nil {
		return fmt.Errorf("failed to parse history response: %w", err)
	}

	messages_list := historyResp["messages"].([]interface{})
	summary := historyResp["summary"].(string)

	fmt.Printf("   ✅ Retrieved %d messages\n", len(messages_list))
	fmt.Printf("   📝 Summary: %s\n", summary)

	// Clear conversation
	fmt.Println("   🗑️  Clearing conversation...")
	clearReq := map[string]interface{}{
		"session_id": sessionID,
	}

	jsonBody, _ := json.Marshal(clearReq)
	req, _ := http.NewRequest("DELETE", baseURL+"/api/v1/ai/chat/history", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err = client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to clear conversation: %w", err)
	}
	defer resp.Body.Close()

	fmt.Println("   ✅ Conversation cleared successfully")

	return nil
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
