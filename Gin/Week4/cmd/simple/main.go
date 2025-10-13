package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"news-aggregator/internal/config"
	"news-aggregator/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// Simple in-memory storage
var newsStorage []models.NewsArticle

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// Setup logger
	logrus.SetLevel(logrus.InfoLevel)

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

	// Serve static files
	router.Static("/web", "./web")
	router.StaticFile("/", "./web/index.html")

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy", "service": "simple-news-api"})
	})

	// API routes
	api := router.Group("/api/v1")
	{
		api.GET("/news", getNews)
		api.GET("/news/:id", getNewsByID)
		api.GET("/news/category/:category", getNewsByCategory)
		api.POST("/collect", collectNews)
	}

	// Start background news generation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go generateNews(ctx)

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

	logrus.WithField("port", cfg.APIPort).Info("Simple News API service started")
	logrus.Info("🌐 Web interface: http://localhost:8080")
	logrus.Info("📡 API: http://localhost:8080/api/v1/news")

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logrus.Info("Shutting down Simple News API service...")
	cancel()

	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logrus.WithError(err).Fatal("Server forced to shutdown")
	}

	logrus.Info("Simple News API service stopped")
}

func getNews(c *gin.Context) {
	// Parse query parameters
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	limit := 10
	offset := 0

	if l, err := fmt.Sscanf(limitStr, "%d", &limit); err == nil && l == 1 {
		if limit <= 0 || limit > 100 {
			limit = 10
		}
	}

	if o, err := fmt.Sscanf(offsetStr, "%d", &offset); err == nil && o == 1 {
		if offset < 0 {
			offset = 0
		}
	}

	// Get news from storage
	start := offset
	end := offset + limit
	if start >= len(newsStorage) {
		start = len(newsStorage)
	}
	if end > len(newsStorage) {
		end = len(newsStorage)
	}

	articles := newsStorage[start:end]

	response := models.NewsResponse{
		Articles: articles,
		Total:    len(articles),
		Page:     (offset / limit) + 1,
		Limit:    limit,
	}

	c.JSON(http.StatusOK, response)
}

func getNewsByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Article ID is required"})
		return
	}

	for _, article := range newsStorage {
		if article.ID == id {
			c.JSON(http.StatusOK, article)
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "Article not found"})
}

func getNewsByCategory(c *gin.Context) {
	category := c.Param("category")
	if category == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Category is required"})
		return
	}

	// Parse query parameters
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	limit := 10
	offset := 0

	if l, err := fmt.Sscanf(limitStr, "%d", &limit); err == nil && l == 1 {
		if limit <= 0 || limit > 100 {
			limit = 10
		}
	}

	if o, err := fmt.Sscanf(offsetStr, "%d", &offset); err == nil && o == 1 {
		if offset < 0 {
			offset = 0
		}
	}

	// Filter by category
	var filteredArticles []models.NewsArticle
	for _, article := range newsStorage {
		if article.Category == category {
			filteredArticles = append(filteredArticles, article)
		}
	}

	// Apply pagination
	start := offset
	end := offset + limit
	if start >= len(filteredArticles) {
		start = len(filteredArticles)
	}
	if end > len(filteredArticles) {
		end = len(filteredArticles)
	}

	articles := filteredArticles[start:end]

	response := models.NewsResponse{
		Articles: articles,
		Total:    len(articles),
		Page:     (offset / limit) + 1,
		Limit:    limit,
	}

	c.JSON(http.StatusOK, response)
}

func generateSampleNews() {
	// Generate some sample news
	articles := []models.NewsArticle{
		{
			ID:          fmt.Sprintf("news_%d", time.Now().Unix()),
			Title:       "Breaking: Technology News Update",
			Content:     "This is a sample technology news article content. It contains information about the latest developments in the tech industry.",
			Source:      "TechNews",
			URL:         "https://example.com/tech-news",
			PublishedAt: time.Now(),
			CreatedAt:   time.Now(),
			Category:    "technology",
		},
		{
			ID:          fmt.Sprintf("news_%d", time.Now().Unix()+1),
			Title:       "Business Market Update",
			Content:     "This is a sample business news article content. It covers the latest market trends and business developments.",
			Source:      "BusinessDaily",
			URL:         "https://example.com/business-news",
			PublishedAt: time.Now(),
			CreatedAt:   time.Now(),
			Category:    "business",
		},
		{
			ID:          fmt.Sprintf("news_%d", time.Now().Unix()+2),
			Title:       "Sports Championship Results",
			Content:     "This is a sample sports news article content. It reports on the latest sports events and championship results.",
			Source:      "SportsCentral",
			URL:         "https://example.com/sports-news",
			PublishedAt: time.Now(),
			CreatedAt:   time.Now(),
			Category:    "sports",
		},
	}

	// Add to storage
	newsStorage = append(articles, newsStorage...)

	// Keep only latest 100 articles
	if len(newsStorage) > 100 {
		newsStorage = newsStorage[:100]
	}
}

func collectNews(c *gin.Context) {
	generateSampleNews()
	
	c.JSON(http.StatusOK, gin.H{
		"message":  "News collection completed",
		"articles": 3,
		"total":    len(newsStorage),
	})
}

func generateNews(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second) // Generate news every 30 seconds
	defer ticker.Stop()

	// Generate initial news
	generateSampleNews()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			logrus.Info("Generating background news...")
			
			// Generate some news
			articles := []models.NewsArticle{
				{
					ID:          fmt.Sprintf("auto_%d", time.Now().Unix()),
					Title:       "Latest Technology Trends",
					Content:     "Automated generation of technology news. This covers the latest trends and developments in the tech industry.",
					Source:      "AutoCollector",
					URL:         "https://example.com/auto-tech",
					PublishedAt: time.Now(),
					CreatedAt:   time.Now(),
					Category:    "technology",
				},
				{
					ID:          fmt.Sprintf("auto_%d", time.Now().Unix()+1),
					Title:       "Market Analysis Report",
					Content:     "Automated generation of business news. This provides insights into current market conditions and business trends.",
					Source:      "AutoCollector",
					URL:         "https://example.com/auto-business",
					PublishedAt: time.Now(),
					CreatedAt:   time.Now(),
					Category:    "business",
				},
			}

			// Add to storage
			newsStorage = append(articles, newsStorage...)

			// Keep only latest 100 articles
			if len(newsStorage) > 100 {
				newsStorage = newsStorage[:100]
			}

			logrus.WithField("count", len(articles)).Info("Background news generation completed")
		}
	}
}
