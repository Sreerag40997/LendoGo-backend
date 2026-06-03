package services

import (
	"lendogo-backend/internal/repositories"
	"log"
	"sync"

	"github.com/gofiber/contrib/websocket"
)

// Client represents a single live user's phone line to the server
type Client struct {
	Conn   *websocket.Conn
	UserID uint // Which user is this? (We can use 0 for the Admin)
}

// ChatHub keeps track of everyone currently online
type ChatHub struct {
	Clients    map[*Client]bool
	Register   chan *Client
	Unregister chan *Client
	mu         sync.Mutex 
	repo       repositories.ChatRepository
}

// NewChatHub creates a fresh hub when the server boots up
func NewChatHub(repo repositories.ChatRepository) *ChatHub {
	return &ChatHub{
		Clients:    make(map[*Client]bool),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		repo:       repo,
	}
}

// Run is an infinite loop that runs in the background listening for people logging on or off
func (h *ChatHub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.mu.Lock()
			h.Clients[client] = true
			h.mu.Unlock()
			log.Printf("🔌 User %d connected to Live Chat. Total online: %d\n", client.UserID, len(h.Clients))

		case client := <-h.Unregister:
			h.mu.Lock()
			if _, ok := h.Clients[client]; ok {
				delete(h.Clients, client)
				client.Conn.Close()
				log.Printf("❌ User %d disconnected. Total online: %d\n", client.UserID, len(h.Clients))
			}
			h.mu.Unlock()
		}
	}
}