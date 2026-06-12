package models

import (
	"time"

	"github.com/google/uuid"
)

// EMISchedule represents a single monthly payment row in the database
type EMISchedule struct {
	ID            uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	LoanID        uuid.UUID `gorm:"type:uuid;not null;index"`
	UserID        uuid.UUID `gorm:"type:uuid;not null;index"`
	InstallmentNo int       `gorm:"not null"`
	DueDate       time.Time `gorm:"not null"`
	EMI           float64   `gorm:"not null"`
	PrincipalPart float64   `gorm:"not null"`
	InterestPart  float64   `gorm:"not null"`
	Status        string    `gorm:"type:varchar(20);default:'PENDING'"` // PENDING, PAID, OVERDUE
	CreatedAt     time.Time
	UpdatedAt     time.Time
}