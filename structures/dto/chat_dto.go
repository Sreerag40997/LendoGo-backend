package dto

import "time"

// IncomingMessage is what React sends to the Go server
type IncomingMessage struct {
	Text        string `json:"text"`
	IsFromAdmin bool   `json:"is_from_admin"`
	ReceiverID  string `json:"receiver_id"` // Who is this going to?
}

// OutgoingMessage is what Go broadcasts back to React
type OutgoingMessage struct {
	SenderID    string    `json:"sender_id"`
	ReceiverID  string    `json:"receiver_id"`
	IsFromAdmin bool      `json:"is_from_admin"`
	Text        string    `json:"text"`
	Timestamp   time.Time `json:"timestamp"`
}