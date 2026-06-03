package models

import (
	"time"
	"gorm.io/gorm"
)

// ChatMessage represents a single text message sent between a user and the admin
type ChatMessage struct {
	gorm.Model
	SenderID    uint      `gorm:"not null" json:"sender_id"`
	ReceiverID  uint      `gorm:"not null" json:"receiver_id"` // 0 if it's broad system support
	IsFromAdmin bool      `gorm:"default:false" json:"is_from_admin"`
	MessageText string    `gorm:"type:text;not null" json:"message_text"`
	IsRead      bool      `gorm:"default:false" json:"is_read"`
	Timestamp   time.Time `gorm:"autoCreateTime" json:"timestamp"`
}