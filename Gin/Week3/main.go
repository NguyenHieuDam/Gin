package main

import (
	"chat-app/controllers"
	"chat-app/middleware"
	"chat-app/models"
	"chat-app/services"
	"log"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	redisService := services.NewRedisService("localhost:6379", "", 0)

	hub := models.NewHub()
	go hub.Run()

	authCtrl := controllers.NewAuthController()
	msgCtrl := controllers.NewMessageController(hub, redisService.GetClient(), authCtrl.UserStore) // truyền userStore
	wsCtrl := controllers.NewWebSocketController(hub)

	rl := middleware.NewRateLimiter()
	r := gin.Default()

	r.Use(cors.Default())

	api := r.Group("/api")
	{
		api.POST("/auth/login", rl.Limit(5, time.Minute), authCtrl.Login)
		api.POST("/messages", rl.Limit(10, time.Minute), msgCtrl.SendMessage)
		api.GET("/messages", msgCtrl.GetMessages)
		api.GET("/online", msgCtrl.GetOnlineUsers)
	}

	r.GET("/ws", wsCtrl.HandleWebSocket)

	port := getEnv("PORT", "8080")
	log.Println("Server running at :" + port)
	r.Run(":" + port)
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
