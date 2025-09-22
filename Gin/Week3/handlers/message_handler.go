package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"WEEK3/services"
)

type MessageHandler struct {
	messageService *services.MessageService
}

func NewMessageHandler(messageService *services.MessageService) *MessageHandler {
	return &MessageHandler{messageService: messageService}
}

// GetMessages retrieves messages for a room
func (h *MessageHandler) GetMessages(c *gin.Context) {
	roomID := c.Param("roomId")
	if roomID == "" {
		roomID = "general"
	}

	// Parse query parameters
	limitStr := c.DefaultQuery("limit", "50")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 100 {
		limit = 50
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	messages, err := h.messageService.GetMessages(roomID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"messages": messages,
		"room_id":  roomID,
		"limit":    limit,
		"offset":   offset,
	})
}

// GetRecentMessages retrieves recent messages for a room
func (h *MessageHandler) GetRecentMessages(c *gin.Context) {
	roomID := c.Param("roomId")
	if roomID == "" {
		roomID = "general"
	}

	limitStr := c.DefaultQuery("limit", "20")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 100 {
		limit = 20
	}

	messages, err := h.messageService.GetRecentMessages(roomID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"messages": messages,
		"room_id":  roomID,
		"limit":    limit,
	})
}

// SearchMessages searches messages by content
func (h *MessageHandler) SearchMessages(c *gin.Context) {
	roomID := c.Param("roomId")
	if roomID == "" {
		roomID = "general"
	}

	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Từ khóa tìm kiếm không được để trống",
		})
		return
	}

	limitStr := c.DefaultQuery("limit", "20")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 100 {
		limit = 20
	}

	messages, err := h.messageService.SearchMessages(roomID, query, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"messages": messages,
		"room_id":  roomID,
		"query":    query,
		"limit":    limit,
	})
}

// DeleteMessage deletes a message
func (h *MessageHandler) DeleteMessage(c *gin.Context) {
	messageIDStr := c.Param("messageId")
	messageID, err := strconv.ParseUint(messageIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "ID tin nhắn không hợp lệ",
		})
		return
	}

	// Trong thực tế, lấy user ID từ JWT token
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Không có quyền truy cập",
		})
		return
	}

	userID := userIDStr.(uint)

	err = h.messageService.DeleteMessage(uint(messageID), userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Xóa tin nhắn thành công",
	})
}

// GetMessageStats returns message statistics
func (h *MessageHandler) GetMessageStats(c *gin.Context) {
	roomID := c.Param("roomId")
	if roomID == "" {
		roomID = "general"
	}

	// This would typically return statistics about messages in a room
	c.JSON(http.StatusOK, gin.H{
		"room_id": roomID,
		"stats": gin.H{
			"total_messages": 0, // Implement actual count
			"active_users":   0, // Implement actual count
		},
	})
}
