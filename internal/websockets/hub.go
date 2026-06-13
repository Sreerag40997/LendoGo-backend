package websockets

import (
	"log"
	"sync"
	"github.com/gofiber/websocket/v2"
)

// Clients holds all active admin connections
var Clients = make(map[*websocket.Conn]bool)
var broadcast = make(chan interface{}, 100)
var Mutex = sync.Mutex{}

// StartHub listens for messages and sends them to all connected admins
func StartHub() {
	for {
		msg := <-broadcast
		Mutex.Lock()
		for client := range Clients {
			err := client.WriteJSON(msg)
			if err != nil {
				log.Printf("WebSocket Error: %v", err)
				client.Close()
				delete(Clients, client)
			}
		}
		Mutex.Unlock()
	}
}

// BroadcastMessage allows your REST controllers to trigger UI updates!
func BroadcastMessage(eventType string, data interface{}) {
	broadcast <- map[string]interface{}{
		"event": eventType,
		"data":  data,
	}
}