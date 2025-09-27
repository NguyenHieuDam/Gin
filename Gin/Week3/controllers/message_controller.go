package controllers

import (
	"chat-app/models"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"golang.org/x/net/context"
)

// MessageController quản lý gửi/nhận tin nhắn
type MessageController struct {
	hub   *models.Hub
	redis *redis.Client
	ctx   context.Context
	users map[string]*models.User // tham chiếu userStore từ AuthController
}

func NewMessageController(hub *models.Hub, redisClient *redis.Client, userStore map[string]*models.User) *MessageController {
	return &MessageController{
		hub:   hub,
		redis: redisClient,
		ctx:   context.Background(),
		users: userStore,
	}
}

// ✅ thêm định nghĩa SendMessageRequest
type SendMessageRequest struct {
	Content string `json:"content" binding:"required"`
}

func (mc *MessageController) SendMessage(c *gin.Context) {
	var req SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token := c.GetHeader("Authorization")
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing token"})
		return
	}

	// 🔎 Tìm trong Hub trước
	var user *models.User
	for _, client := range mc.hub.Clients {
		if client.User.ID == token {
			user = client.User
			break
		}
	}

	// 🔎 Nếu chưa có trong Hub thì fallback về userStore
	if user == nil {
		if u, ok := mc.users[token]; ok {
			user = u
		}
	}

	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	// ✅ Đến đây chắc chắn có user
	message := models.NewMessage(user.ID, user.Username, req.Content)

	msgJSON, _ := json.Marshal(message)
	mc.redis.LPush(mc.ctx, "messages", msgJSON)
	mc.redis.LTrim(mc.ctx, "messages", 0, 99)

	wsMsg := models.WebSocketMessage{
		Type:      models.WSMessageTypeMessage,
		Data:      message,
		Timestamp: time.Now(),
	}
	wsJSON, _ := json.Marshal(wsMsg)
	mc.hub.Broadcast <- wsJSON

	c.JSON(http.StatusOK, gin.H{"message": message})
}

// GetMessages trả về lịch sử tin nhắn
func (mc *MessageController) GetMessages(c *gin.Context) {
	results, err := mc.redis.LRange(mc.ctx, "messages", 0, -1).Result()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch messages"})
		return
	}

	var messages []*models.Message
	for _, r := range results {
		var msg models.Message
		if err := json.Unmarshal([]byte(r), &msg); err == nil {
			messages = append(messages, &msg)
		}
	}

	c.JSON(http.StatusOK, gin.H{"messages": messages})
}

// GetOnlineUsers trả về danh sách user đang online
func (mc *MessageController) GetOnlineUsers(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"users": mc.hub.GetOnlineUsers()})
}
