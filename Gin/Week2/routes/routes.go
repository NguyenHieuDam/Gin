package routes

import (
	"week2/controllers"
	"week2/middlewares"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	// Gắn middleware
	r.Use(middlewares.LoggerMiddleware())
	r.Use(middlewares.RateLimitMiddleware())

	// Auth
	r.POST("/register", controllers.Register)
	r.POST("/login", controllers.Login)

	// Protected routes
	auth := r.Group("/api")
	{
		api := auth.Group("/tasks")
		{
			api.GET("", controllers.GetTasks)
			api.POST("", controllers.CreateTask)
			api.PUT("/:id", controllers.UpdateTask)
			api.DELETE("/:id", controllers.DeleteTask)
		}
	}

	return r
}
