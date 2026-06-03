// internal/repositories/chat_repository.go
package repositories

import (
	"lendogo-backend/structures/models"
	"gorm.io/gorm"
)

type ChatRepository interface {
	SaveMessage(msg *models.ChatMessage) error
	GetChatHistory(userID uint) ([]models.ChatMessage, error)
}

type chatRepositoryImpl struct {
	db *gorm.DB
}

func NewChatRepository(db *gorm.DB) ChatRepository {
	return &chatRepositoryImpl{db: db}
}

func (r *chatRepositoryImpl) SaveMessage(msg *models.ChatMessage) error {
	return r.db.Create(msg).Error
}

func (r *chatRepositoryImpl) GetChatHistory(userID uint) ([]models.ChatMessage, error) {
	var messages []models.ChatMessage
	err := r.db.Where("sender_id = ? OR receiver_id = ?", userID, userID).
		Order("timestamp asc").Find(&messages).Error
	return messages, err
}