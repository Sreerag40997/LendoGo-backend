package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
    ID        uuid.UUID      `gorm:"type:uuid;primaryKey"` // Use UUID instead of uint
    FullName  string         `json:"full_name"`
    Email     string         `json:"email" gorm:"unique"`
    Password  string         `json:"password"`
    Role      string         `gorm:"type:varchar(20);default:'user'"`
    CreatedAt time.Time
    UpdatedAt time.Time
    DeletedAt gorm.DeletedAt `gorm:"index"`
}
// GORM automatically runs this right before saving a new user to the database!
func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	// If the ID is completely blank (all zeros), generate a new real one!
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return
}