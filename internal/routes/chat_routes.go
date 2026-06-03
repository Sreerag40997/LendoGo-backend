package routes

import (
	"github.com/gofiber/fiber/v2"
	"lendogo-backend/internal/controllers/chat_controller"
)

func SetupChatRoutes(api fiber.Router, controller *chat_controller.ChatController) {
	// Notice we use api.Use for the Upgrade middleware
	chatGroup := api.Group("/ws")
	
	// First it hits the upgrade middleware, then it hits the connection handler
	chatGroup.Get("/chat", controller.UpgradeWebSocket, controller.HandleConnection())
}