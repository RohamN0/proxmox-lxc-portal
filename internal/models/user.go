package models

import (
	"time"

	"gorm.io/gorm"
)

// User struct
type User struct {
	gorm.Model
	Username  string     `gorm:"uniqueIndex;not null" json:"username"`
	Email     string     `gorm:"uniqueIndex;not null" json:"email"`
	Password  string     `gorm:"not null" json:"-"`
	Names     string     `json:"names"`
	LastLogin *time.Time `json:"last_login"`
}

// CreateUserRequest contains request payload fields for user creation.
type CreateUserRequest struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}
