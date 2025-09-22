package models

import (
	"time"

	"gorm.io/gorm"
)

// User represents a user in the chat system
type User struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Username  string    `json:"username" gorm:"uniqueIndex;not null"`
	Email     string    `json:"email" gorm:"uniqueIndex;not null"`
	Password  string    `json:"-" gorm:"not null"` // Ẩn password khỏi JSON
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	IsOnline  bool      `json:"is_online" gorm:"default:false"`
	LastSeen  time.Time `json:"last_seen"`
}

// UserRequest represents the request body for user registration/login
type UserRequest struct {
	Username string `json:"username" binding:"required,min=3,max=20"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// LoginRequest represents the request body for user login
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// UserResponse represents the response for user data (without password)
type UserResponse struct {
	ID        uint      `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	IsOnline  bool      `json:"is_online"`
	LastSeen  time.Time `json:"last_seen"`
}

// ToResponse converts User to UserResponse
func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:        u.ID,
		Username:  u.Username,
		Email:     u.Email,
		CreatedAt: u.CreatedAt,
		IsOnline:  u.IsOnline,
		LastSeen:  u.LastSeen,
	}
}

// BeforeCreate hook để hash password trước khi tạo user
func (u *User) BeforeCreate(tx *gorm.DB) error {
	// Hash password ở đây nếu cần
	return nil
}
