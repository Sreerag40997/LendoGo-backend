package chat_controller

import (
	"log"
	"strconv"
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

// Upgrade HTTP to WebSocket
func (cc *ChatController) UpgradeWebSocket(c *fiber.Ctx) error {
	// Is this a websocket request?
	if websocket.IsWebSocketUpgrade(c) {
		return c.Next() // Allow it to pass to the handler
	}
	return fiber.ErrUpgradeRequired
}

// Handle the active connection
func (cc *ChatController) HandleConnection() fiber.Handler {
	return websocket.New(func(c *websocket.Conn) {
		// 1. Get the user ID from the connection URL (e.g., ws://.../chat?user_id=5)
		// For the admin dashboard, we can pass user_id=0
	userIDString := c.Query("user_id", "0")
userID, _ := strconv.Atoi(userIDString)

		// 2. Create the client object
		client := &services.Client{
			Conn:   c,
			UserID: uint(userID),
		}

		// 3. Register the user into the Hub
		cc.hub.Register <- client

		// 4. When they close the browser, unregister them
		defer func() {
			cc.hub.Unregister <- client
		}()

		// 5. Infinite loop to listen for incoming messages from this user
// 5. Infinite loop to listen for incoming messages
		for {
			var incomingMsg dto.IncomingMessage 
			
			err := c.ReadJSON(&incomingMsg)
			if err != nil {
				log.Println("WebSocket closed or error:", err)
				break 
			}

			// Format the response DTO
			responseMsg := dto.OutgoingMessage{
				SenderID:    client.UserID,
				IsFromAdmin: incomingMsg.IsFromAdmin,
				Text:        incomingMsg.Text,
				Timestamp:   time.Now(),
			}

			// Right now, just log it. (Next step: broadcast it and save to DB!)
			log.Printf("📩 Clean Message Received: %+v\n", responseMsg)
		}
	})
}