package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserProfile struct {
	gorm.Model
	// The Foreign Key connecting to your existing users table
	UserID uuid.UUID `gorm:"type:uuid;uniqueIndex;not null" json:"user_id"`

	ProfileImage string `gorm:"type:varchar(255)" json:"profile_image"`
	PhoneNumber  string `gorm:"type:varchar(20)" json:"phone_number"`
	DateOfBirth  string `gorm:"type:varchar(20)" json:"date_of_birth"` 
	Pincode      string `gorm:"type:varchar(10)" json:"pincode"`
	Address      string `gorm:"type:text" json:"address"`
	City         string `gorm:"type:varchar(100)" json:"city"`
	State        string `gorm:"type:varchar(100)" json:"state"`
	TrustScore   int    `gorm:"type:int;default:0" json:"trust_score"`
	PanCardNumber string `gorm:"type:varchar(20)" json:"pan_card_number"`
	CreditRating  string `gorm:"type:varchar(50)" json:"credit_rating"`
	// Relationship
	User *User `gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
}