package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/exotic-travel-booking/backend/internal/llm/providers"
	"github.com/exotic-travel-booking/backend/internal/services"
)

func main() {
	fmt.Println("🦙 Ollama Integration Demo for Exotic Travel Booking")
	fmt.Println("===================================================")

	ctx := context.Background()

	// Create Ollama service
	ollamaService := services.NewOllamaService("http://localhost:11434")

	// Check if Ollama is running
	fmt.Println("\n1. Checking Ollama health...")
	if err := ollamaService.CheckHealth(ctx); err != nil {
		log.Fatalf("❌ Ollama is not running or accessible: %v\n", err)
	}
	fmt.Println("✅ Ollama is healthy and running!")

	// List available models
	fmt.Println("\n2. Listing available models...")
	models, err := ollamaService.ListModels(ctx)
	if err != nil {
		log.Fatalf("❌ Failed to list models: %v\n", err)
	}

	if len(models) == 0 {
		fmt.Println("⚠️  No models found. Let's pull a recommended model...")
		
		// Pull a small model for demo
		modelName := "llama3.2:1b" // Small 1B parameter model
		fmt.Printf("📥 Pulling model: %s (this may take a while...)\n", modelName)
		
		if err := ollamaService.PullModel(ctx, modelName); err != nil {
			log.Fatalf("❌ Failed to pull model: %v\n", err)
		}
		
		fmt.Printf("✅ Successfully pulled model: %s\n", modelName)
		
		// List models again
		models, err = ollamaService.ListModels(ctx)
		if err != nil {
			log.Fatalf("❌ Failed to list models after pull: %v\n", err)
		}
	}

	fmt.Printf("📋 Found %d models:\n", len(models))
	for i, model := range models {
		fmt.Printf("   %d. %s (%.2f GB)\n", i+1, model.Name, float64(model.Size)/(1024*1024*1024))
	}

	// Use the first available model for demo
	if len(models) == 0 {
		log.Fatal("❌ No models available for demo")
	}

	selectedModel := models[0].Name
	fmt.Printf("\n3. Using model: %s\n", selectedModel)

	// Demo 1: Simple travel query
	fmt.Println("\n4. Demo 1: Simple Travel Query")
	fmt.Println("   Query: 'What are the best travel destinations in Japan?'")
	
	response, err := ollamaService.GenerateResponse(ctx, selectedModel, 
		"What are the best travel destinations in Japan? Please provide a brief list with 3-4 destinations.")
	if err != nil {
		log.Printf("❌ Failed to generate response: %v\n", err)
	} else {
		fmt.Printf("   Response: %s\n", response)
	}

	// Demo 2: Travel planning query
	fmt.Println("\n5. Demo 2: Travel Planning Query")
	fmt.Println("   Query: 'Plan a 3-day itinerary for Tokyo'")
	
	response, err = ollamaService.GenerateResponse(ctx, selectedModel,
		"Plan a 3-day itinerary for Tokyo, Japan. Include must-see attractions, recommended restaurants, and transportation tips. Keep it concise.")
	if err != nil {
		log.Printf("❌ Failed to generate response: %v\n", err)
	} else {
		fmt.Printf("   Response: %s\n", response)
	}

	// Demo 3: Streaming response
	fmt.Println("\n6. Demo 3: Streaming Response")
	fmt.Println("   Query: 'Describe the culture and cuisine of Thailand'")
	fmt.Print("   Streaming Response: ")
	
	stream, err := ollamaService.StreamResponse(ctx, selectedModel,
		"Describe the culture and cuisine of Thailand in 2-3 paragraphs.")
	if err != nil {
		log.Printf("❌ Failed to start streaming: %v\n", err)
	} else {
		for chunk := range stream {
			fmt.Print(chunk)
		}
		fmt.Println()
	}

	// Demo 4: Provider integration
	fmt.Println("\n7. Demo 4: LLM Provider Integration")
	
	config := &providers.LLMConfig{
		Provider: "ollama",
		Model:    selectedModel,
		BaseURL:  "http://localhost:11434",
		Timeout:  60 * time.Second,
	}

	provider, err := providers.NewOllamaProvider(config)
	if err != nil {
		log.Printf("❌ Failed to create provider: %v\n", err)
	} else {
		req := &providers.GenerateRequest{
			Model: selectedModel,
			Messages: []providers.Message{
				{
					Role:    "system",
					Content: "You are a helpful travel assistant specializing in exotic destinations.",
				},
				{
					Role:    "user",
					Content: "What makes Bhutan a unique travel destination?",
				},
			},
			MaxTokens:   200,
			Temperature: 0.7,
		}

		resp, err := provider.GenerateResponse(ctx, req)
		if err != nil {
			log.Printf("❌ Failed to generate response via provider: %v\n", err)
		} else {
			fmt.Printf("   Provider Response: %s\n", resp.Choices[0].Message.Content)
		}
	}

	// Demo 5: Model recommendations
	fmt.Println("\n8. Demo 5: Model Recommendations")
	recommended := ollamaService.GetRecommendedModels()
	fmt.Printf("📋 Recommended models for travel use cases:\n")
	for i, model := range recommended {
		fmt.Printf("   %d. %s\n", i+1, model)
	}

	// Check status of recommended models
	fmt.Println("\n9. Model Status Check")
	status, err := ollamaService.GetModelStatus(ctx)
	if err != nil {
		log.Printf("❌ Failed to get model status: %v\n", err)
	} else {
		fmt.Printf("📊 Model availability status:\n")
		for _, s := range status {
			if s.Available {
				fmt.Printf("   ✅ %s (%.2f GB)\n", s.Name, float64(s.Size)/(1024*1024*1024))
			} else {
				fmt.Printf("   ❌ %s (not installed)\n", s.Name)
			}
		}
	}

	fmt.Println("\n🎉 Ollama integration demo completed successfully!")
	fmt.Println("\nNext steps:")
	fmt.Println("1. Start the travel booking server: go run cmd/server/main.go")
	fmt.Println("2. Test Ollama endpoints:")
	fmt.Println("   - GET  /api/v1/ollama/health")
	fmt.Println("   - GET  /api/v1/ollama/models")
	fmt.Println("   - POST /api/v1/ollama/generate")
	fmt.Println("3. Integrate with travel agents for AI-powered trip planning")
}
