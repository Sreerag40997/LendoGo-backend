package services

import (
	"log"
	"sync"

	"lendogo-backend/internal/repositories"
	"lendogo-backend/structures/dto"
	"lendogo-backend/structures/models"
	"lendogo-backend/utils"

	"github.com/gofiber/contrib/websocket"
)

type Client struct {
	Conn   *websocket.Conn
	UserID string
}

type BroadcastMessage struct {
	Sender  *Client
	Payload dto.OutgoingMessage
}

type ChatHub struct {
	Clients    map[*Client]bool
	Register   chan *Client
	Unregister chan *Client
	Broadcast  chan BroadcastMessage
	mu         sync.Mutex
	Repo       repositories.ChatRepository // Exported field for cross-package access
}

// NewChatHub initializes the WebSocket hub and binds the repository dependency
func NewChatHub(repo repositories.ChatRepository) *ChatHub {
	return &ChatHub{
		Clients:    make(map[*Client]bool),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Broadcast:  make(chan BroadcastMessage, 256),
		Repo:       repo, // Bind dependency
	}
}

// Run executes the core event loop for connection state and message broadcasting
func (h *ChatHub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.mu.Lock()
			h.Clients[client] = true
			h.mu.Unlock()
			log.Printf("🔌 User %s connected. Total: %d\n", client.UserID, len(h.Clients))

		case client := <-h.Unregister:
			h.mu.Lock()
			if _, ok := h.Clients[client]; ok {
				delete(h.Clients, client)
				client.Conn.Close()
				log.Printf("❌ User %s disconnected. Total: %d\n", client.UserID, len(h.Clients))
			}
			h.mu.Unlock()

		case bm := <-h.Broadcast:
			// 1. Persist to database prior to broadcast
			msg := &models.ChatMessage{
				SenderID:    bm.Payload.SenderID,
				ReceiverID:  bm.Payload.ReceiverID,
				IsFromAdmin: bm.Payload.IsFromAdmin,
				MessageText: bm.Payload.Text,
			}
			if err := h.Repo.SaveMessage(msg); err != nil {
				log.Println("DB save error:", err)
			}

			h.mu.Lock()
			for client := range h.Clients {
				// 2. Unicast to intended receiver
				if client.UserID == bm.Payload.ReceiverID {
					if err := utils.SafeWriteJSON(client.Conn, bm.Payload); err != nil {
						log.Println("Write error:", err)
						delete(h.Clients, client)
						client.Conn.Close()
					}
				}

				// 3. Mirror payload to active admin sessions if inbound from standard user
				if utils.IsAdminUser(client.UserID) && !bm.Payload.IsFromAdmin {
					if err := utils.SafeWriteJSON(client.Conn, bm.Payload); err != nil {
						log.Println("Admin write error:", err)
						delete(h.Clients, client)
						client.Conn.Close()
					}
				}
			}
			h.mu.Unlock()
		}
	}
}