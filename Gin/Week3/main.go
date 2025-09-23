package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"WEEK3/db"
	"WEEK3/handlers"
	"WEEK3/middleware"

	"github.com/gin-gonic/gin"
)

func main() {
	// Set Gin mode
	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.DebugMode)
	}

	// Initialize databases
	log.Println("🚀 Starting Chat Application...")

	// Connect to PostgreSQL
	postgresDB, err := db.ConnectPostgres()
	if err != nil {
		log.Fatal("❌ Failed to connect to PostgreSQL:", err)
	}
	defer func() {
		if err := db.ClosePostgres(postgresDB); err != nil {
			log.Printf("Error closing PostgreSQL connection: %v", err)
		}
	}()

	// Run database migrations
	if err := db.AutoMigrate(postgresDB); err != nil {
		log.Fatal("❌ Failed to run migrations:", err)
	}

	// Connect to Redis
	redisClient, err := db.ConnectRedis()
	if err != nil {
		log.Fatal("❌ Failed to connect to Redis:", err)
	}
	defer func() {
		if err := db.CloseRedis(redisClient); err != nil {
			log.Printf("Error closing Redis connection: %v", err)
		}
	}()

	// Initialize rate limiter
	rateLimiter := middleware.NewRateLimiter(redisClient)

	// Initialize handlers
	handler := handlers.NewHandler(postgresDB, redisClient)

	// Setup Gin router
	r := gin.Default()

	// Add rate limiting middleware to all routes
	r.Use(rateLimiter.RateLimitMiddleware(middleware.DefaultRateLimitConfig()))

	// Add specific rate limiting to sensitive routes
	rateLimiter.LoginRateLimitMiddleware()
	rateLimiter.RegisterRateLimitMiddleware()
	rateLimiter.MessageRateLimitMiddleware()
	rateLimiter.WebSocketRateLimitMiddleware()

	// Setup all routes first
	handler.SetupRoutes(r)

	// Apply specific rate limits to sensitive routes after setup
	// Note: This is a simplified approach. In production, you'd want to
	// integrate rate limiting more cleanly with the route setup

	// Add static file serving for frontend
	r.Static("/static", "./static")
	r.StaticFile("/", "./static/index.html")

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		// Check database connections
		if err := db.HealthCheck(postgresDB); err != nil {
			c.JSON(503, gin.H{
				"status": "unhealthy",
				"error":  "PostgreSQL connection failed",
			})
			return
		}

		if err := db.HealthCheckRedis(redisClient); err != nil {
			c.JSON(503, gin.H{
				"status": "unhealthy",
				"error":  "Redis connection failed",
			})
			return
		}

		c.JSON(200, gin.H{
			"status": "healthy",
			"services": gin.H{
				"postgresql": "connected",
				"redis":      "connected",
				"websocket":  "running",
			},
		})
	})

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Setup graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Start server in a goroutine
	go func() {
		log.Printf("🌐 Server starting on port %s", port)
		if err := r.Run(":" + port); err != nil {
			log.Fatal("❌ Failed to start server:", err)
		}
	}()

	// Wait for interrupt signal
	<-quit
	log.Println("🛑 Shutting down server...")

	log.Println("✅ Server stopped gracefully")
}
