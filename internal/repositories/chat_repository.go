package repositories

import (
	"lendogo-backend/structures/models"

	"gorm.io/gorm"
)

type ChatRepository interface {
	SaveMessage(msg *models.ChatMessage) error
	GetChatHistory(userID string) ([]models.ChatMessage, error)
	GetActiveChatSessions() ([]models.ChatMessage, error)
}

type chatRepositoryImpl struct {
	db *gorm.DB
}

func NewChatRepository(db *gorm.DB) ChatRepository {
	return &chatRepositoryImpl{db: db}
}

// SaveMessage saves a new chat to the database
func (r *chatRepositoryImpl) SaveMessage(msg *models.ChatMessage) error {
	return r.db.Create(msg).Error
}

// GetChatHistory fetches the full conversation for one specific user
func (r *chatRepositoryImpl) GetChatHistory(userID string) ([]models.ChatMessage, error) {
	var messages []models.ChatMessage
	
	// Preload("Sender") ensures we get the User's name and email!
	err := r.db.Preload("Sender").
		Where("sender_id = ? OR receiver_id = ?", userID, userID).
		Order("timestamp asc").
		Find(&messages).Error
		
	return messages, err
}

// GetActiveChatSessions fetches the latest message for every unique user 
// so the Admin can see a list of active conversations on the sidebar.
func (r *chatRepositoryImpl) GetActiveChatSessions() ([]models.ChatMessage, error) {
	var senderIDs []string

	// Step 1: Get a clean list of all unique user IDs who have sent a message
	err := r.db.Model(&models.ChatMessage{}).
		Where("is_from_admin = ?", false).
		Distinct("sender_id").
		Pluck("sender_id", &senderIDs).Error

	if err != nil {
		return nil, err
	}

	var latestMessages []models.ChatMessage

	// Step 2: Fetch the single latest message for each user, including their Profile
	for _, senderID := range senderIDs {
		var msg models.ChatMessage
		
		err := r.db.Preload("Sender").
			Where("sender_id = ?", senderID).
			Order("timestamp desc").
			First(&msg).Error

		// If we successfully found their latest message, add it to the list
		if err == nil {
			latestMessages = append(latestMessages, msg)
		}
	}

	return latestMessages, nil
}