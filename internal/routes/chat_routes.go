package routes

import (
	"lendogo-backend/internal/controllers/chat_controller"
	"lendogo-backend/internal/middlewares"
	"lendogo-backend/internal/services"
	"github.com/gofiber/fiber/v2"
)

// SetupChatRoutes wires up both the WebSocket tunnels and the Admin REST APIs
func SetupChatRoutes(api fiber.Router, controller *chat_controller.ChatController, configService services.ConfigService) {

	// ==========================================
	// 1. WEBSOCKET ROUTE (The Live Tunnel)
	// ==========================================
	// React connects to this using: ws://localhost:8080/api/ws/chat?user_id=123
	wsGroup := api.Group("/ws")
	wsGroup.Get("/chat", middlewares.RequireFeature(configService, "chat_support"), controller.UpgradeWebSocket, controller.HandleConnection())

	// ==========================================
	// 2. REST API ROUTES (For Admin Dashboard)
	// ==========================================
	// React fetches this using: http://localhost:8080/api/admin/chats/sessions
	adminChatGroup := api.Group("/admin/chats")
	adminChatGroup.Get("/sessions", controller.GetAdminChatSessions)

}
