package chat_controller

import (
	"log"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"

	"lendogo-backend/internal/services"
	"lendogo-backend/structures/dto"
)

type ChatController struct {
	hub *services.ChatHub
}

func NewChatController(hub *services.ChatHub) *ChatController {
	return &ChatController{hub: hub}
}

// ==========================================
// 1. WEBSOCKET LOGIC (Real-time chatting)
// ==========================================

// Upgrade HTTP to WebSocket
func (cc *ChatController) UpgradeWebSocket(c *fiber.Ctx) error {
	if websocket.IsWebSocketUpgrade(c) {
		return c.Next()
	}
	return fiber.ErrUpgradeRequired
}

// Handle the active connection
func (cc *ChatController) HandleConnection() fiber.Handler {
	return websocket.New(func(c *websocket.Conn) {
		// Get the user ID from the connection URL as a string (UUID)
		userID := c.Query("user_id", "0")

		client := &services.Client{
			Conn:   c,
			UserID: userID,
		}

		cc.hub.Register <- client

		defer func() {
			cc.hub.Unregister <- client
		}()

		// Infinite loop to listen for incoming messages
		for {
			var incomingMsg dto.IncomingMessage

			err := c.ReadJSON(&incomingMsg)
			if err != nil {
				log.Println("WebSocket closed or error:", err)
				break
			}

			responseMsg := dto.OutgoingMessage{
				SenderID:    client.UserID,
				ReceiverID:  incomingMsg.ReceiverID,
				IsFromAdmin: incomingMsg.IsFromAdmin,
				Text:        incomingMsg.Text,
				Timestamp:   time.Now(),
			}

			// Push into the hub — it handles DB save + routing
			cc.hub.Broadcast <- services.BroadcastMessage{
				Sender:  client,
				Payload: responseMsg,
			}

			log.Printf("📩 Clean Message Received: %+v\n", responseMsg)
		}
	})
}

// ==========================================
// 2. REST API LOGIC (For Admin Initial Load)
// ==========================================

// GetAdminChatSessions returns the list of users currently chatting
func (cc *ChatController) GetAdminChatSessions(c *fiber.Ctx) error {
	// Fetch the preloaded active sessions from our repository
	sessions, err := cc.hub.Repo.GetActiveChatSessions()
	
	if err != nil {
		log.Println("Error fetching chat sessions:", err)
		return c.Status(500).JSON(fiber.Map{
			"success": false, 
			"error": "Failed to load chats",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    sessions,
	})
}