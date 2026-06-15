package models

import (
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt" // 👈 We use this directly to break the import cycle!
	"gorm.io/gorm"
)

type Staff struct {
	ID          uuid.UUID       `gorm:"type:uuid;primaryKey" json:"id"`
	FullName    string          `json:"full_name"`
	Email       string          `json:"email" gorm:"unique"`
	Avatar      string          `json:"avatar"`
	Password    string          `json:"-"`
	Role        string          `gorm:"type:varchar(50)" json:"role"`
	Status      string          `gorm:"type:varchar(20);default:'Active'" json:"status"`
	Permissions map[string]bool `gorm:"serializer:json" json:"permissions"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
	DeletedAt   gorm.DeletedAt  `gorm:"index" json:"-"`
}

func (s *Staff) BeforeCreate(tx *gorm.DB) (err error) {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	
	if s.Password != "" {
		// 👇 Your exact logic, just using bcrypt directly instead of the utils folder
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(s.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		s.Password = string(hashedPassword)
	}
	return nil
}