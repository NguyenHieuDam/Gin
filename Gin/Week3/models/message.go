package models

import (
	"time"

	"gorm.io/gorm"
)

// Message represents a chat message
type Message struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    uint      `json:"user_id" gorm:"not null"`
	Username  string    `json:"username" gorm:"not null"`
	Content   string    `json:"content" gorm:"not null"`
	RoomID    string    `json:"room_id" gorm:"default:'general'"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	
	// Relations
	User User `json:"user" gorm:"foreignKey:UserID"`
}

// MessageRequest represents the request body for sending a message
type MessageRequest struct {
	Content string `json:"content" binding:"required,max=1000"`
	RoomID  string `json:"room_id" binding:"omitempty,max=50"`
}

// MessageResponse represents the response for message data
type MessageResponse struct {
	ID        uint      `json:"id"`
	UserID    uint      `json:"user_id"`
	Username  string    `json:"username"`
	Content   string    `json:"content"`
	RoomID    string    `json:"room_id"`
	CreatedAt time.Time `json:"created_at"`
}

// ToResponse converts Message to MessageResponse
func (m *Message) ToResponse() MessageResponse {
	return MessageResponse{
		ID:        m.ID,
		UserID:    m.UserID,
		Username:  m.Username,
		Content:   m.Content,
		RoomID:    m.RoomID,
		CreatedAt: m.CreatedAt,
	}
}

// WebSocketMessage represents a message sent through WebSocket
type WebSocketMessage struct {
	Type      string          `json:"type"` // message, user_joined, user_left, typing, etc.
	Data      MessageResponse `json:"data,omitempty"`
	User      UserResponse    `json:"user,omitempty"`
	OnlineUsers []UserResponse `json:"online_users,omitempty"`
	Timestamp time.Time       `json:"timestamp"`
}

// BeforeCreate hook
func (m *Message) BeforeCreate(tx *gorm.DB) error {
	if m.RoomID == "" {
		m.RoomID = "general"
	}
	return nil
}
