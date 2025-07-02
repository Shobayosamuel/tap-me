package repository

import (
	"github.com/Shobayosamuel/tap-me/internal/models"
	"gorm.io/gorm"
)

type MessageRepository interface {
	Create(message *models.Message) error
	GetByID(id uint) (*models.Message, error)
	GetByIDWithRelations(id uint) (*models.Message, error)
	GetRoomMessages(roomID uint, limit, offset int) ([]models.Message, error)
	Update(message *models.Message) error
	Delete(id uint) error
	GetUserMessages(userID uint, limit, offset int) ([]models.Message, error)
}

type messageRepository struct {
	db *gorm.DB
}

func NewMessageRepository(db *gorm.DB) MessageRepository {
	return &messageRepository{db: db}
}

func (r *messageRepository) Create(message *models.Message) error {
	return r.db.Create(message).Error
}

func (r *messageRepository) GetByID(id uint) (*models.Message, error) {
	var message models.Message
	err := r.db.First(&message, id).Error
	if err != nil {
		return nil, err
	}
	return &message, nil
}

func (r *messageRepository) GetByIDWithRelations(id uint) (*models.Message, error) {
	var message models.Message
	err := r.db.Preload("User").Preload("Room").First(&message, id).Error
	if err != nil {
		return nil, err
	}
	return &message, nil
}

func (r *messageRepository) GetRoomMessages(roomID uint, limit, offset int) ([]models.Message, error) {
	var messages []models.Message

	query := r.db.Preload("User").
		Where("room_id = ?", roomID).
		Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	if offset > 0 {
		query = query.Offset(offset)
	}

	err := query.Find(&messages).Error
	if err != nil {
		return nil, err
	}

	// Reverse the slice to get chronological order (oldest first)
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}

func (r *messageRepository) GetUserMessages(userID uint, limit, offset int) ([]models.Message, error) {
	var messages []models.Message

	query := r.db.Preload("User").Preload("Room").
		Where("user_id = ?", userID).
		Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	if offset > 0 {
		query = query.Offset(offset)
	}

	err := query.Find(&messages).Error
	if err != nil {
		return nil, err
	}

	return messages, nil
}

func (r *messageRepository) Update(message *models.Message) error {
	return r.db.Save(message).Error
}

func (r *messageRepository) Delete(id uint) error {
	return r.db.Delete(&models.Message{}, id).Error
}