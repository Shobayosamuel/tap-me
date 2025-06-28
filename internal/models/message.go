package models

import (
	"time"

	"gorm.io/gorm"
)

type Message struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Content   string         `json:"content" gorm:"not null"`
	UserID    uint           `json:"user_id" gorm:"not null"`
	User      User           `json:"user" gorm:"foreignKey:UserID"`
	RoomID    uint           `json:"room_id" gorm:"not null"`
	Room      Room           `json:"room" gorm:"foreignKey:RoomID"`
	Type      MessageType    `json:"type" gorm:"default:text"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

type MessageType string

const (
	MessageTypeText   MessageType = "text"
	MessageTypeImage  MessageType = "image"
	MessageTypeFile   MessageType = "file"
	MessageTypeSystem MessageType = "system"
)

type Room struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Name        string         `json:"name" gorm:"not null"`
	Description string         `json:"description"`
	IsPrivate   bool           `json:"is_private" gorm:"default:false"`
	CreatedBy   uint           `json:"created_by" gorm:"not null"`
	Creator     User           `json:"creator" gorm:"foreignKey:CreatedBy"`
	Messages    []Message      `json:"messages,omitempty" gorm:"foreignKey:RoomID"`
	Members     []User         `json:"members,omitempty" gorm:"many2many:room_members;"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

type RoomMember struct {
	RoomID		uint			`json:"room_id" gorm:"primaryKey"`
	UserID 		uint			`json:"user_id" gorm:"primaryKey"`
	JoinedAt time.Time `json:"joined_at"`
	Role     MemberRole `json:"role" gorm:"default:member"`
}

type MemberRole string

const (
	RoleAdmin     MemberRole = "admin"
	RoleModerator MemberRole = "moderator"
	RoleMember    MemberRole = "member"
)