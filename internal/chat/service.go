package chat

import (
	"github.com/Shobayosamuel/tap-me/internal/models"
	"github.com/Shobayosamuel/tap-me/internal/repository"
)

type Service interface {
	CreateRoom(userID uint, req CreateRoomRequest) (*models.Room, error)
	GetUserRooms(userID uint) ([]models.Room, error)
	GetRoomMessages(userID, roomID uint, limit, offset int) ([]models.Message, error)
	JoinRoom(userID, roomID uint) error
	LeaveRoom(userID, roomID uint) error
	CreateMessage(userID, roomID uint, content string) (*models.Message, error)
	CanUserAccessRoom(userID, roomID uint) (bool, error)
	GetRoomMembers(roomID uint) ([]models.User, error)
}

type service struct {
	roomRepo    repository.RoomRepository
	messageRepo repository.MessageRepository
	userRepo    repository.UserRepository
}

