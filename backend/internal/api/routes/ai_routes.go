package routes

import (
	"github.com/exotic-travel-booking/backend/internal/api/handlers"
	"github.com/gofiber/fiber/v2"
)

// SetupAIRoutes configures AI-related API routes
func SetupAIRoutes(app *fiber.App, aiHandler *handlers.AIHandler, healthHandler *handlers.HealthHandler) {
	// API version group
	api := app.Group("/api/v1")

	// Health and monitoring routes
	health := api.Group("/health")
	health.Get("/", healthHandler.Health)
	health.Get("/ready", healthHandler.Readiness)
	health.Get("/live", healthHandler.Liveness)
	health.Get("/metrics", healthHandler.Metrics)

	// AI routes group with middleware
	ai := api.Group("/ai")
	// Note: Using basic middleware for now, can be enhanced with Fiber-specific middleware

	// Chat endpoints
	chat := ai.Group("/chat")
	chat.Post("/", aiHandler.Chat)
	chat.Get("/history/:sessionId", aiHandler.GetConversationHistory)
	chat.Delete("/history", aiHandler.ClearConversation)

	// Knowledge base endpoints
	kb := ai.Group("/knowledge")
	kb.Post("/query", aiHandler.QueryKnowledgeBase)

	// Agent endpoints
	agents := ai.Group("/agents")
	agents.Post("/request", aiHandler.ProcessAgentRequest)

	// Streaming endpoints
	stream := ai.Group("/stream")
	stream.Get("/chat", aiHandler.Chat) // Will handle streaming based on request
}

// SetupWebSocketRoutes configures WebSocket routes for real-time AI interactions
func SetupWebSocketRoutes(app *fiber.App, aiHandler *handlers.AIHandler) {
	// WebSocket endpoint for real-time chat
	app.Get("/ws/chat", func(c *fiber.Ctx) error {
		// WebSocket upgrade logic would go here
		// For now, we'll use HTTP streaming
		return c.SendString("WebSocket endpoint - to be implemented")
	})
}
