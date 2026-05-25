package models

import (
    "time"
)

type Consultation struct {
    ID          uint      `gorm:"primaryKey"`
    FullName    string    `gorm:"type:varchar(100);not null"`
    Email       string    `gorm:"type:varchar(100);not null"`
    PhoneNumber string    `gorm:"type:varchar(20);not null"`
    CreatedAt   time.Time
    UpdatedAt   time.Time
}