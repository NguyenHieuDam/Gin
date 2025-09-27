package models

import (
	"time"

	"github.com/google/uuid"
)

type Message struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Username  string    `json:"username"`
	Content   string    `json:"content"`
	Type      string    `json:"type"`
	Timestamp time.Time `json:"timestamp"`
}

const (
	MessageTypeText = "text"
)

func NewMessage(userID, username, content string) *Message {
	return &Message{
		ID:        uuid.New().String(),
		UserID:    userID,
		Username:  username,
		Content:   content,
		Type:      MessageTypeText,
		Timestamp: time.Now(),
	}
}
