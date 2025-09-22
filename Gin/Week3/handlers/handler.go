package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
	"WEEK3/services"
	"WEEK3/websocket"
)

type Handler struct {
	userService     *services.UserService
	messageService  *services.MessageService
	presenceService *services.PresenceService
	authHandler     *AuthHandler
	messageHandler  *MessageHandler
	wsHandler       *WSHandler
	hub             *websocket.Hub
	db              *gorm.DB
	redis           *redis.Client
}

func NewHandler(db *gorm.DB, redis *redis.Client) *Handler {
	// Initialize services
	userService := services.NewUserService(db)
	messageService := services.NewMessageService(db, redis)
	presenceService := services.NewPresenceService(db, redis)

	// Create WebSocket hub
	hub := websocket.NewHub()

	// Initialize handlers
	authHandler := NewAuthHandler(userService)
	messageHandler := NewMessageHandler(messageService)
	wsHandler := NewWSHandler(hub, presenceService, messageService)

	return &Handler{
		userService:     userService,
		messageService:  messageService,
		presenceService: presenceService,
		authHandler:     authHandler,
		messageHandler:  messageHandler,
		wsHandler:       wsHandler,
		hub:             hub,
		db:              db,
		redis:           redis,
	}
}

// SetupRoutes sets up all the routes
func (h *Handler) SetupRoutes(r *gin.Engine) {
	// CORS middleware
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		
		c.Next()
	})

	// Health check
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Chat App API is running",
			"time":    time.Now(),
		})
	})

	// API routes
	api := r.Group("/api/v1")
	{
		// Auth routes
		auth := api.Group("/auth")
		{
			auth.POST("/register", h.authHandler.Register)
			auth.POST("/login", h.authHandler.Login)
			auth.GET("/profile", h.authHandler.GetProfile)
			auth.PUT("/profile", h.authHandler.UpdateProfile)
			auth.POST("/logout", h.authHandler.Logout)
		}

		// Message routes
		messages := api.Group("/messages")
		{
			messages.GET("/:roomId", h.messageHandler.GetMessages)
			messages.GET("/:roomId/recent", h.messageHandler.GetRecentMessages)
			messages.GET("/:roomId/search", h.messageHandler.SearchMessages)
			messages.DELETE("/:messageId", h.messageHandler.DeleteMessage)
			messages.GET("/:roomId/stats", h.messageHandler.GetMessageStats)
		}

		// WebSocket routes
		ws := api.Group("/ws")
		{
			ws.GET("/", h.wsHandler.ServeWebSocket)
			ws.GET("/rooms", h.wsHandler.GetRooms)
			ws.GET("/:roomId/users", h.wsHandler.GetOnlineUsers)
			ws.GET("/:roomId/stats", h.wsHandler.GetRoomStats)
		}

		// User routes
		users := api.Group("/users")
		{
			users.GET("/", func(c *gin.Context) {
				users, err := h.userService.GetAllUsers()
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{
						"error": err.Error(),
					})
					return
				}
				c.JSON(http.StatusOK, gin.H{
					"users": users,
				})
			})
			users.GET("/online", func(c *gin.Context) {
				users, err := h.userService.GetOnlineUsers()
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{
						"error": err.Error(),
					})
					return
				}
				c.JSON(http.StatusOK, gin.H{
					"online_users": users,
				})
			})
		}
	}

	// Start WebSocket hub
	go h.hub.Run()
}

// GetHub returns the WebSocket hub
func (h *Handler) GetHub() *websocket.Hub {
	return h.hub
}

// GetDB returns the database connection
func (h *Handler) GetDB() *gorm.DB {
	return h.db
}

// GetRedis returns the Redis client
func (h *Handler) GetRedis() *redis.Client {
	return h.redis
}

// GetAuthHandler returns the auth handler
func (h *Handler) GetAuthHandler() *AuthHandler {
	return h.authHandler
}

// GetMessageHandler returns the message handler
func (h *Handler) GetMessageHandler() *MessageHandler {
	return h.messageHandler
}

// GetWSHandler returns the WebSocket handler
func (h *Handler) GetWSHandler() *WSHandler {
	return h.wsHandler
}