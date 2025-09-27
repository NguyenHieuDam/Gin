package controllers

import (
	"chat-app/models"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type WebSocketController struct {
	hub      *models.Hub
	upgrader websocket.Upgrader
}

func NewWebSocketController(hub *models.Hub) *WebSocketController {
	return &WebSocketController{
		hub:      hub,
		upgrader: websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }},
	}
}

func (wc *WebSocketController) HandleWebSocket(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing token"})
		return
	}

	conn, err := wc.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}

	user := &models.User{ID: token, Username: "User_" + token[:6], Email: "demo@chat.com", IsOnline: true}
	client := &models.Client{ID: user.ID, User: user, Conn: conn, Send: make(chan []byte, 256), Hub: wc.hub}

	wc.hub.Register <- client
	go client.WritePump()
	go client.ReadPump()
}
