package repository

import (
	"github.com/Shobayosamuel/tap-me/internal/models"
	"gorm.io/gorm"
)

type RoomRepository interface {
	Create(room *models.Room) error
	AddMember(roomID, userID uint, role models.MemberRole) error
	GetUserRooms(userID uint) ([]models.Room, error)
	GetByID(roomID uint) (*models.Room, error)
	RemoveMember(roomID, userID uint) error
	IsUserMember(roomID, userID uint) error
}

type roomRepository struct {
	db *gorm.DB
}

func NewRoomRepository(db *gorm.DB) RoomRepository {
	return &roomRepository{db: db}
}

func (r *roomRepository) Create(room *models.Room) error {
	return r.db.Create(room).Error
}

func (r *roomRepository) AddMember(roomID, userID uint, role models.MemberRole) error {
	roomMember := &models.RoomMember{
		RoomID: roomID,
		UserID: userID,
		Role:   role,
	}
	return r.db.Create(roomMember).Error
}

func (r *roomRepository) GetUserRooms(userID uint) ([]models.Room, error) {
	var rooms []models.Room
	err := r.db.Joins("JOIN room_members ON room_members.room_id = rooms.id").
		Where("room_members.user_id = ?", userID).
		Find(&rooms).Error
	if err != nil {
		return nil, err
	}
	return rooms, nil
}

func (r *roomRepository) GetByID(roomID uint) (*models.Room, error) {
	var room models.Room
	err := r.db.Where("room_id = ?", roomID).First(&room).Error
	if err != nil {
		return nil, err
	}
	return &room, nil
}

func (r *roomRepository) RemoveMember(roomID, userID uint) error {
	return r.db.Where("room_id = ? AND user_id = ?", roomID, userID).Delete(&models.RoomMember{}).Error
}

// IsUserMember checks if a user is a member of the specified room.
func (r *roomRepository) IsUserMember(roomID, userID uint) error {
	var member models.RoomMember
	err := r.db.Where("room_id = ? AND user_id = ?", roomID, userID).First(&member).Error
	return err
}
