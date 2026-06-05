package models

import (
	"time"

	"gorm.io/gorm"
)

// RefreshToken represents a refresh token in the system
type RefreshToken struct {
	gorm.Model
	UserId    uint      `gorm:"not null" json:"user_id"`
	Token     string    `gorm:"uniqueIndex;not null" json:"token"`
	ExpiresAt time.Time `gorm:"not null" json:"expires_at"`
	Revoked   bool      `gorm:"not null" json:"revoked"`
}
