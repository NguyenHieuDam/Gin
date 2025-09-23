package websocket

import (
	"encoding/json"
	"log"
	"time"

	"WEEK3/models"
)

// BroadcastMessage represents a message to be broadcasted
type BroadcastMessage struct {
	Message *models.Message
	RoomID  string
	Type    string // message, typing, user_joined, user_left
}

// Hub quản lý tất cả client đang kết nối
type Hub struct {
	// Map of room ID to clients in that room
	rooms map[string]map[*Client]bool
	// All clients regardless of room
	clients map[*Client]bool
	// Channel for registering clients
	register chan *Client
	// Channel for unregistering clients
	unregister chan *Client
	// Channel for broadcasting messages
	broadcast chan *BroadcastMessage
}

func NewHub() *Hub {
	return &Hub{
		rooms:      make(map[string]map[*Client]bool),
		clients:    make(map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan *BroadcastMessage),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.registerClient(client)

		case client := <-h.unregister:
			h.unregisterClient(client)

		case message := <-h.broadcast:
			h.broadcastMessage(message)
		}
	}
}

// registerClient adds a client to the hub and appropriate room
func (h *Hub) registerClient(client *Client) {
	h.clients[client] = true

	// Add to room
	if h.rooms[client.roomID] == nil {
		h.rooms[client.roomID] = make(map[*Client]bool)
	}
	h.rooms[client.roomID][client] = true

	log.Printf("User %s joined room %s", client.user.Username, client.roomID)

	// Notify room about new user
	h.notifyUserJoined(client)
}

// unregisterClient removes a client from the hub and room
func (h *Hub) unregisterClient(client *Client) {
	if _, ok := h.clients[client]; ok {
		delete(h.clients, client)
		close(client.send)

		// Remove from room
		if room, exists := h.rooms[client.roomID]; exists {
			delete(room, client)
			if len(room) == 0 {
				delete(h.rooms, client.roomID)
			}
		}

		log.Printf("User %s left room %s", client.user.Username, client.roomID)

		// Notify room about user leaving
		h.notifyUserLeft(client)
	}
}

// broadcastMessage sends a message to all clients in a room
func (h *Hub) broadcastMessage(msg *BroadcastMessage) {
	room, exists := h.rooms[msg.RoomID]
	if !exists {
		return
	}

	var wsMessage models.WebSocketMessage

	switch msg.Type {
	case "message":
		wsMessage = models.WebSocketMessage{
			Type:      "message",
			Data:      msg.Message.ToResponse(),
			Timestamp: time.Now(),
		}
	case "typing":
		wsMessage = models.WebSocketMessage{
			Type: "typing",
			User: models.UserResponse{
				ID:       msg.Message.UserID,
				Username: msg.Message.Username,
			},
			Timestamp: time.Now(),
		}
	case "user_joined":
		wsMessage = models.WebSocketMessage{
			Type: "user_joined",
			User: models.UserResponse{
				ID:       msg.Message.UserID,
				Username: msg.Message.Username,
			},
			OnlineUsers: h.getOnlineUsersInRoom(msg.RoomID),
			Timestamp:   time.Now(),
		}
	case "user_left":
		wsMessage = models.WebSocketMessage{
			Type: "user_left",
			User: models.UserResponse{
				ID:       msg.Message.UserID,
				Username: msg.Message.Username,
			},
			OnlineUsers: h.getOnlineUsersInRoom(msg.RoomID),
			Timestamp:   time.Now(),
		}
	default:
		return
	}

	messageBytes, err := json.Marshal(wsMessage)
	if err != nil {
		log.Printf("Error marshaling message: %v", err)
		return
	}

	// Send to all clients in the room
	for client := range room {
		select {
		case client.send <- messageBytes:
		default:
			close(client.send)
			delete(h.clients, client)
			delete(room, client)
		}
	}
}

// notifyUserJoined sends a notification when a user joins
func (h *Hub) notifyUserJoined(client *Client) {
	h.broadcast <- &BroadcastMessage{
		Message: &models.Message{
			UserID:   client.user.ID,
			Username: client.user.Username,
		},
		RoomID: client.roomID,
		Type:   "user_joined",
	}
}

// notifyUserLeft sends a notification when a user leaves
func (h *Hub) notifyUserLeft(client *Client) {
	h.broadcast <- &BroadcastMessage{
		Message: &models.Message{
			UserID:   client.user.ID,
			Username: client.user.Username,
		},
		RoomID: client.roomID,
		Type:   "user_left",
	}
}

// getOnlineUsersInRoom returns list of online users in a room
func (h *Hub) getOnlineUsersInRoom(roomID string) []models.UserResponse {
	room, exists := h.rooms[roomID]
	if !exists {
		return []models.UserResponse{}
	}

	var users []models.UserResponse
	for client := range room {
		users = append(users, client.user.ToResponse())
	}
	return users
}

// GetOnlineUsersCount returns the number of online users in a room
func (h *Hub) GetOnlineUsersCount(roomID string) int {
	if room, exists := h.rooms[roomID]; exists {
		return len(room)
	}
	return 0
}

// GetAllRooms returns a list of all room IDs
func (h *Hub) GetAllRooms() []string {
	var rooms []string
	for roomID := range h.rooms {
		rooms = append(rooms, roomID)
	}
	return rooms
}
