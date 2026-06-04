package models

import (
	"time"
	"gorm.io/gorm"
)

type ChatMessage struct {
	gorm.Model
	// 1. Change these to string to match your UUIDs!
	SenderID    string    `gorm:"type:varchar(255);not null;index" json:"sender_id"`
	ReceiverID  string    `gorm:"type:varchar(255);not null;index" json:"receiver_id"` 
	
	IsFromAdmin bool      `gorm:"default:false" json:"is_from_admin"`
	MessageText string    `gorm:"type:text;not null" json:"message_text"`
	IsRead      bool      `gorm:"default:false" json:"is_read"`
	Timestamp   time.Time `gorm:"autoCreateTime" json:"timestamp"`

	// 👇 THE MAGIC: This tells GORM to link this message to the actual User table
	Sender      User      `gorm:"foreignKey:SenderID;references:ID" json:"sender"`
}