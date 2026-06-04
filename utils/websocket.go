package utils

import (
    "log"
    "github.com/gofiber/contrib/websocket"
)

// SafeWriteJSON writes to a websocket connection and recovers from panics
// (e.g. writing to a closed connection)
func SafeWriteJSON(conn *websocket.Conn, payload any) (err error) {
    defer func() {
        if r := recover(); r != nil {
            log.Println("⚠️ Recovered from websocket panic:", r)
        }
    }()
    return conn.WriteJSON(payload)
}

// IsAdminUser returns true if the userID represents the admin
// Centralise this logic so you don't scatter userID == 0 checks everywhere
func IsAdminUser(userID string) bool {
    return userID == "0" || userID == "admin"
}