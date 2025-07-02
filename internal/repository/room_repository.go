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
	IsUserMember(roomID, userID uint) (bool, error)
	GetMembers(roomID uint) ([]models.User, error)
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

func (r *roomRepository) GetByID(id uint) (*models.Room, error) {
	var room models.Room
	err := r.db.Preload("Creator").First(&room, id).Error
	if err != nil {
		return nil, err
	}
	return &room, nil
}

func (r *roomRepository) GetUserRooms(userID uint) ([]models.Room, error) {
	var rooms []models.Room
	err := r.db.Preload("Creator").
		Joins("JOIN room_members ON room_members.room_id = rooms.id").
		Where("room_members.user_id = ?", userID).
		Find(&rooms).Error

	if err != nil {
		return nil, err
	}
	return rooms, nil
}

func (r *roomRepository) AddMember(roomID, userID uint, role models.MemberRole) error {
	// Check if user is already a member
	var existingMember models.RoomMember
	err := r.db.Where("room_id = ? AND user_id = ?", roomID, userID).First(&existingMember).Error

	if err == nil {
		// User is already a member, update role if different
		if existingMember.Role != role {
			existingMember.Role = role
			return r.db.Save(&existingMember).Error
		}
		return nil // Already a member with same role
	}

	if err != gorm.ErrRecordNotFound {
		return err // Database error
	}

	// Add new member
	member := models.RoomMember{
		RoomID: roomID,
		UserID: userID,
		Role:   role,
	}

	return r.db.Create(&member).Error
}

func (r *roomRepository) RemoveMember(roomID, userID uint) error {
	return r.db.Where("room_id = ? AND user_id = ?", roomID, userID).
		Delete(&models.RoomMember{}).Error
}

func (r *roomRepository) IsUserMember(roomID, userID uint) (bool, error) {
	var count int64
	err := r.db.Model(&models.RoomMember{}).
		Where("room_id = ? AND user_id = ?", roomID, userID).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *roomRepository) GetMembers(roomID uint) ([]models.User, error) {
	var users []models.User
	err := r.db.Joins("JOIN room_members ON room_members.user_id = users.id").
		Where("room_members.room_id = ?", roomID).
		Find(&users).Error

	if err != nil {
		return nil, err
	}
	return users, nil
}

func (r *roomRepository) GetMemberRole(roomID, userID uint) (models.MemberRole, error) {
	var member models.RoomMember
	err := r.db.Where("room_id = ? AND user_id = ?", roomID, userID).
		First(&member).Error

	if err != nil {
		return "", err
	}

	return member.Role, nil
}

func (r *roomRepository) UpdateMemberRole(roomID, userID uint, role models.MemberRole) error {
	return r.db.Model(&models.RoomMember{}).
		Where("room_id = ? AND user_id = ?", roomID, userID).
		Update("role", role).Error
}
	
func (r *roomRepository) Update(room *models.Room) error {
	return r.db.Save(room).Error
}

func (r *roomRepository) Delete(id uint) error {
	// Start transaction
	tx := r.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Delete all room members
	if err := tx.Where("room_id = ?", id).Delete(&models.RoomMember{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Delete all messages in the room
	if err := tx.Where("room_id = ?", id).Delete(&models.Message{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Delete the room
	if err := tx.Delete(&models.Room{}, id).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}