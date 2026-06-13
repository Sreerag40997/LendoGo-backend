package utils

import (
	"log"

	"lendogo-backend/database"
	"lendogo-backend/internal/websockets"
	"lendogo-backend/structures/models"

	"github.com/google/uuid"
)

// RecordAudit writes a permanent log to the DB and alerts the real-time Hub
func RecordAudit(actorID uuid.UUID, actorName, actionType, entityType, entityID, description, ipAddress string) {
	auditEntry := models.AuditLog{
		ActorID:     actorID,
		ActorName:   actorName,
		ActionType:  actionType,
		EntityType:  entityType,
		EntityID:    entityID,
		Description: description,
		IPAddress:   ipAddress,
	}

	// 1. Save to PostgreSQL Vault
	if err := database.DB.Create(&auditEntry).Error; err != nil {
		log.Printf("⚠️ CRITICAL COMPLIANCE ERROR: Failed to write audit log: %v\n", err)
		return // We don't crash the server, but we must log the failure
	}

	// 2. Broadcast to React UI (The little notification bell will ring!)
	websockets.BroadcastMessage("NEW_AUDIT_LOG", auditEntry)
}