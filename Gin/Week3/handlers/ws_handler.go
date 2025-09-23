package handlers

import (
	"net/http"

	"WEEK3/services"
	"WEEK3/websocket"

	"github.com/gin-gonic/gin"
)

type WSHandler struct {
	hub             *websocket.Hub
	presenceService *services.PresenceService
	messageService  *services.MessageService
}

func NewWSHandler(hub *websocket.Hub, presenceService *services.PresenceService, messageService *services.MessageService) *WSHandler {
	return &WSHandler{
		hub:             hub,
		presenceService: presenceService,
		messageService:  messageService,
	}
}

// ServeWebSocket handles WebSocket connections
func (h *WSHandler) ServeWebSocket(c *gin.Context) {
	websocket.ServeWS(h.hub, c)
}

// GetOnlineUsers returns online users in a room
func (h *WSHandler) GetOnlineUsers(c *gin.Context) {
	roomID := c.Param("roomId")
	if roomID == "" {
		roomID = "general"
	}

	users, err := h.presenceService.GetOnlineUsersInRoom(roomID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	count, err := h.presenceService.GetOnlineUsersCount(roomID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"online_users": users,
		"count":        count,
		"room_id":      roomID,
	})
}

// GetRooms returns all available rooms
func (h *WSHandler) GetRooms(c *gin.Context) {
	rooms := h.hub.GetAllRooms()

	// Add default room if not exists
	hasGeneral := false
	for _, room := range rooms {
		if room == "general" {
			hasGeneral = true
			break
		}
	}
	if !hasGeneral {
		rooms = append([]string{"general"}, rooms...)
	}

	c.JSON(http.StatusOK, gin.H{
		"rooms": rooms,
	})
}

// GetRoomStats returns statistics for a specific room
func (h *WSHandler) GetRoomStats(c *gin.Context) {
	roomID := c.Param("roomId")
	if roomID == "" {
		roomID = "general"
	}

	count, err := h.presenceService.GetOnlineUsersCount(roomID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"room_id":      roomID,
		"online_users": count,
		"total_rooms":  len(h.hub.GetAllRooms()),
	})
}
