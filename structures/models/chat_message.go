package models

import (
	"time"
	"gorm.io/gorm"
)

type ChatMessage struct {
	gorm.Model
	
	// 1. Strings to accommodate both UUIDs and Admin "0"
	SenderID    string    `gorm:"type:varchar(255);not null;index" json:"sender_id"`
	ReceiverID  string    `gorm:"type:varchar(255);not null;index" json:"receiver_id"`

	IsFromAdmin bool      `gorm:"default:false" json:"is_from_admin"`
	MessageText string    `gorm:"type:text;not null" json:"message_text"`
	IsRead      bool      `gorm:"default:false" json:"is_read"`
	Timestamp   time.Time `gorm:"autoCreateTime" json:"timestamp"`

	// 👇 THE FIX: Make it a pointer (*User) and add `constraint:-`
	// This tells PostgreSQL to ignore strict rules so "0" doesn't crash the database!
	Sender      *User     `gorm:"foreignKey:SenderID;references:ID;constraint:-" json:"sender"`
}