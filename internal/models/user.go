package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID uint `json:"id" gorm:"primaryKey"`
	Username string `json:"username" gorm:"uniqueIndex; not null"`
	Email string `json:"email" gorm:"uniqueIndex; not null"`
	Password string `json:"-" gorm:"not null"`
	IsActive bool `json:"is_active" gorm:"default:true"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}