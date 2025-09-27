package models

import (
	"log"
	"time"

	"github.com/gorilla/websocket"
)

type WebSocketMessage struct {
	Type      string      `json:"type"`
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
}

const (
	WSMessageTypeMessage = "message"
)

type Client struct {
	ID   string
	User *User
	Conn *websocket.Conn
	Send chan []byte
	Hub  *Hub
}

func (c *Client) ReadPump() {
	defer func() {
		c.Hub.Unregister <- c
		c.Conn.Close()
	}()

	for {
		_, msg, err := c.Conn.ReadMessage()
		if err != nil {
			break
		}
		c.Hub.Broadcast <- msg
	}
}

func (c *Client) WritePump() {
	for msg := range c.Send {
		if err := c.Conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			log.Println("write error:", err)
			break
		}
	}
}

type Hub struct {
	Clients    map[string]*Client
	Register   chan *Client
	Unregister chan *Client
	Broadcast  chan []byte
}

func NewHub() *Hub {
	return &Hub{
		Clients:    make(map[string]*Client),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Broadcast:  make(chan []byte),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.Clients[client.ID] = client
			client.User.IsOnline = true

		case client := <-h.Unregister:
			if _, ok := h.Clients[client.ID]; ok {
				client.User.IsOnline = false
				delete(h.Clients, client.ID)
				close(client.Send)
			}

		case msg := <-h.Broadcast:
			for _, client := range h.Clients {
				select {
				case client.Send <- msg:
				default:
					close(client.Send)
					delete(h.Clients, client.ID)
				}
			}
		}
	}
}

func (h *Hub) GetOnlineUsers() []*User {
	var users []*User
	for _, client := range h.Clients {
		if client.User.IsOnline {
			users = append(users, client.User)
		}
	}
	return users
}
