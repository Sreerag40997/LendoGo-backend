package models

import (
	"time"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"lendogo-backend/utils"
)

type Staff struct {
	ID          uuid.UUID       `gorm:"type:uuid;primaryKey" json:"id"`
	FullName    string          `json:"full_name"`
	Email       string          `json:"email" gorm:"unique"`
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
		hashedPassword, err := utils.HashPassword(s.Password)
		if err != nil {
			return err
		}
		s.Password = hashedPassword
	}
	return nil
}