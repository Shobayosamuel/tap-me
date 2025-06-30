package chat

import (
	"errors"

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

func NewService(roomRepo repository.RoomRepository, messageRepo repository.MessageRepository, userRepo repository.UserRepository) Service {
	return &service{
		roomRepo:    roomRepo,
		messageRepo: messageRepo,
		userRepo:    userRepo,
	}
}

func (s *service) CreateRoom(userID uint, req CreateRoomRequest) (*models.Room, error) {
	room := &models.Room{
		Name:        req.Name,
		Description: req.Description,
		IsPrivate:   req.IsPrivate,
		CreatedBy:   userID,
	}

	if err := s.roomRepo.Create(room); err != nil {
		return nil, err
	}

	// Add creator as room member with admin role
	if err := s.roomRepo.AddMember(room.ID, userID, models.RoleAdmin); err != nil {
		return nil, err
	}

	return room, nil
}

func (s *service) GetUserRooms(userID uint) ([]models.Room, error) {
	return s.roomRepo.GetUserRooms(userID)
}

func (s *service) GetRoomMessages(userID, roomID uint, limit, offset int) ([]models.Message, error) {
	// Check if user can access room
	canAccess, err := s.CanUserAccessRoom(userID, roomID)
	if err != nil {
		return nil, err
	}
	if !canAccess {
		return nil, errors.New("access denied")
	}

	return s.messageRepo.GetRoomMessages(roomID, limit, offset)
}

func (s *service) JoinRoom(userID, roomID uint) error {
	// Check if room exists
	room, err := s.roomRepo.GetByID(roomID)
	if err != nil {
		return err
	}

	// Check if room is private
	if room.IsPrivate {
		return errors.New("cannot join private room")
	}

	// Add user as member
	return s.roomRepo.AddMember(roomID, userID, models.RoleMember)
}

func (s *service) LeaveRoom(userID, roomID uint) error {
	return s.roomRepo.RemoveMember(roomID, userID)
}

func (s *service) CreateMessage(userID, roomID uint, content string) (*models.Message, error) {
	// Check if user can access room
	canAccess, err := s.CanUserAccessRoom(userID, roomID)
	if err != nil {
		return nil, err
	}
	if !canAccess {
		return nil, errors.New("access denied")
	}

	message := &models.Message{
		Content: content,
		UserID:  userID,
		RoomID:  roomID,
		Type:    models.MessageTypeText,
	}

	if err := s.messageRepo.Create(message); err != nil {
		return nil, err
	}

	// Load user and room data
	return s.messageRepo.GetByIDWithRelations(message.ID)
}

func (s *service) CanUserAccessRoom(userID, roomID uint) (bool, error) {
	return s.roomRepo.IsUserMember(roomID, userID)
}

func (s *service) GetRoomMembers(roomID uint) ([]models.User, error) {
	return s.roomRepo.GetMembers(roomID)
}