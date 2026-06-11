package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID              uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"` // Use UUID instead of uint
	FullName        string         `json:"full_name"`
	Email           string         `json:"email" gorm:"unique"`
	Password        string         `json:"password"`
	Role            string         `gorm:"type:varchar(20);default:'user'" json:"role"`
	Status          string         `gorm:"type:varchar(20);default:'Active'" json:"status"`
	IsEmailVerified bool           `json:"is_email_verified" gorm:"default:false"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
	
	// 👇 THE FIX: Added the '*' to make it a pointer!
	Profile         *UserProfile   `gorm:"foreignKey:UserID;references:ID" json:"profile"` 
}

// GORM automatically runs this right before saving a new user to the database!
func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	// If the ID is completely blank (all zeros), generate a new real one!
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return
}