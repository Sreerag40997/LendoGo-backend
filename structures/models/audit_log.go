package models

import (
	"time"
	"github.com/google/uuid"
)
type AuditLog struct {
	ID          uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"` 
	ActorID     uuid.UUID `gorm:"type:uuid;not null" json:"actor_id"`
	ActorName   string    `gorm:"type:varchar(100);not null" json:"actor_name"`
	ActionType  string    `gorm:"type:varchar(50);index" json:"action_type"`
	EntityType  string    `gorm:"type:varchar(50)" json:"entity_type"`
	EntityID    string    `gorm:"type:varchar(100)" json:"entity_id"`
	Description string    `gorm:"type:text;not null" json:"description"`
	IPAddress   string    `gorm:"type:varchar(45)" json:"ip_address"`
	CreatedAt   time.Time `gorm:"index" json:"created_at"`
}