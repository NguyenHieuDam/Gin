package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	IsOnline  bool      `json:"is_online"`
	LastSeen  time.Time `json:"last_seen"`
	CreatedAt time.Time `json:"created_at"`
}

func NewUser(username, email string) *User {
	return &User{
		ID:        uuid.New().String(),
		Username:  username,
		Email:     email,
		IsOnline:  false,
		LastSeen:  time.Now(),
		CreatedAt: time.Now(),
	}
}
