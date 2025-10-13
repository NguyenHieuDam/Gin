package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"news-aggregator/internal/config"
	"news-aggregator/internal/kafka"
	"news-aggregator/internal/middleware"
	"news-aggregator/internal/models"
	"news-aggregator/internal/redis"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// Setup logger
	logrus.SetLevel(logrus.InfoLevel)

	// Initialize Redis client
	redisClient := redis.NewClient(cfg.RedisURL)
	defer redisClient.Close()

	// Initialize Kafka consumer
	consumer := kafka.NewConsumer(cfg.KafkaBroker, cfg.KafkaTopic, "news-api-group")
	defer consumer.Close()

	// Start Kafka consumer in background
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		err := consumer.ConsumeNews(ctx, func(article models.NewsArticle) error {
			return redisClient.StoreNews(ctx, article)
		})
		if err != nil {
			logrus.WithError(err).Error("Kafka consumer error")
		}
	}()

	// Initialize Gin router
	router := gin.Default()

	// Add CORS middleware
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		
		c.Next()
	})

	// Initialize rate limiter
	rateLimiter := middleware.NewRateLimiter(redisClient.GetClient(), cfg.RateLimitRequests, cfg.RateLimitWindow)

	// Health check endpoint (no rate limiting)
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy", "service": "news-api"})
	})

	// Serve static files
	router.Static("/web", "./web")
	router.StaticFile("/", "./web/index.html")

	// API routes with rate limiting
	api := router.Group("/api/v1")
	api.Use(rateLimiter.RateLimit())
	{
		api.GET("/news", getNews(redisClient))
		api.GET("/news/:id", getNewsByID(redisClient))
		api.GET("/news/category/:category", getNewsByCategory(redisClient))
	}

	// Start server
	server := &http.Server{
		Addr:    ":" + cfg.APIPort,
		Handler: router,
	}

	// Graceful shutdown
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.WithError(err).Fatal("Failed to start server")
		}
	}()

	logrus.WithField("port", cfg.APIPort).Info("News API service started")

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logrus.Info("Shutting down News API service...")
	cancel()

	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logrus.WithError(err).Fatal("Server forced to shutdown")
	}

	logrus.Info("News API service stopped")
}

func getNews(redisClient *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Parse query parameters
		limitStr := c.DefaultQuery("limit", "10")
		offsetStr := c.DefaultQuery("offset", "0")

		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit <= 0 || limit > 100 {
			limit = 10
		}

		offset, err := strconv.Atoi(offsetStr)
		if err != nil || offset < 0 {
			offset = 0
		}

		// Get news from Redis
		articles, err := redisClient.GetLatestNews(c.Request.Context(), limit, offset)
		if err != nil {
			logrus.WithError(err).Error("Failed to get latest news")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve news"})
			return
		}

		response := models.NewsResponse{
			Articles: articles,
			Total:    len(articles),
			Page:     (offset / limit) + 1,
			Limit:    limit,
		}

		c.JSON(http.StatusOK, response)
	}
}

func getNewsByID(redisClient *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Article ID is required"})
			return
		}

		article, err := redisClient.GetNews(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Article not found"})
			return
		}

		c.JSON(http.StatusOK, article)
	}
}

func getNewsByCategory(redisClient *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		category := c.Param("category")
		if category == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Category is required"})
			return
		}

		// Parse query parameters
		limitStr := c.DefaultQuery("limit", "10")
		offsetStr := c.DefaultQuery("offset", "0")

		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit <= 0 || limit > 100 {
			limit = 10
		}

		offset, err := strconv.Atoi(offsetStr)
		if err != nil || offset < 0 {
			offset = 0
		}

		// Get news by category from Redis
		articles, err := redisClient.GetNewsByCategory(c.Request.Context(), category, limit, offset)
		if err != nil {
			logrus.WithError(err).Error("Failed to get news by category")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve news"})
			return
		}

		response := models.NewsResponse{
			Articles: articles,
			Total:    len(articles),
			Page:     (offset / limit) + 1,
			Limit:    limit,
		}

		c.JSON(http.StatusOK, response)
	}
}
