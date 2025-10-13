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
	"news-aggregator/internal/kafka"
	"news-aggregator/internal/models"

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

	// Initialize Kafka producer
	producer := kafka.NewProducer(cfg.KafkaBroker, cfg.KafkaTopic)
	defer producer.Close()

	// Initialize Gin router
	router := gin.Default()

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	// Manual news collection endpoint
	router.POST("/collect", func(c *gin.Context) {
		collectNews(c, producer)
	})

	// Start background news collection
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go startBackgroundCollection(ctx, producer)

	// Start server
	server := &http.Server{
		Addr:    ":" + cfg.CollectorPort,
		Handler: router,
	}

	// Graceful shutdown
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.WithError(err).Fatal("Failed to start server")
		}
	}()

	logrus.WithField("port", cfg.CollectorPort).Info("News Collector service started")

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logrus.Info("Shutting down News Collector service...")
	cancel()

	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logrus.WithError(err).Fatal("Server forced to shutdown")
	}

	logrus.Info("News Collector service stopped")
}

func collectNews(c *gin.Context, producer *kafka.Producer) {
	// Simulate news collection from different sources
	articles := []models.NewsArticle{
		{
			ID:          fmt.Sprintf("news_%d", time.Now().Unix()),
			Title:       "Breaking: Technology News Update",
			Content:     "This is a sample technology news article content...",
			Source:      "TechNews",
			URL:         "https://example.com/tech-news",
			PublishedAt: time.Now(),
			CreatedAt:   time.Now(),
			Category:    "technology",
		},
		{
			ID:          fmt.Sprintf("news_%d", time.Now().Unix()+1),
			Title:       "Business Market Update",
			Content:     "This is a sample business news article content...",
			Source:      "BusinessDaily",
			URL:         "https://example.com/business-news",
			PublishedAt: time.Now(),
			CreatedAt:   time.Now(),
			Category:    "business",
		},
		{
			ID:          fmt.Sprintf("news_%d", time.Now().Unix()+2),
			Title:       "Sports Championship Results",
			Content:     "This is a sample sports news article content...",
			Source:      "SportsCentral",
			URL:         "https://example.com/sports-news",
			PublishedAt: time.Now(),
			CreatedAt:   time.Now(),
			Category:    "sports",
		},
	}

	// Publish each article to Kafka
	for _, article := range articles {
		if err := producer.PublishNews(c.Request.Context(), article); err != nil {
			logrus.WithError(err).WithField("article_id", article.ID).Error("Failed to publish article")
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "News collection completed",
		"articles": len(articles),
	})
}

func startBackgroundCollection(ctx context.Context, producer *kafka.Producer) {
	ticker := time.NewTicker(30 * time.Second) // Collect news every 30 seconds
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			logrus.Info("Starting background news collection...")
			
			// Simulate collecting news from external APIs
			articles := collectFromExternalSources()
			
			// Publish to Kafka
			for _, article := range articles {
				if err := producer.PublishNews(ctx, article); err != nil {
					logrus.WithError(err).WithField("article_id", article.ID).Error("Failed to publish article")
				}
			}
			
			logrus.WithField("count", len(articles)).Info("Background news collection completed")
		}
	}
}

func collectFromExternalSources() []models.NewsArticle {
	// This is a simplified version. In a real implementation, you would:
	// 1. Call external news APIs (NewsAPI, RSS feeds, etc.)
	// 2. Parse the responses
	// 3. Transform to your internal format
	
	now := time.Now()
	return []models.NewsArticle{
		{
			ID:          fmt.Sprintf("auto_%d", now.Unix()),
			Title:       "Latest Technology Trends",
			Content:     "Automated collection of technology news...",
			Source:      "AutoCollector",
			URL:         "https://example.com/auto-tech",
			PublishedAt: now,
			CreatedAt:   now,
			Category:    "technology",
		},
		{
			ID:          fmt.Sprintf("auto_%d", now.Unix()+1),
			Title:       "Market Analysis Report",
			Content:     "Automated collection of business news...",
			Source:      "AutoCollector",
			URL:         "https://example.com/auto-business",
			PublishedAt: now,
			CreatedAt:   now,
			Category:    "business",
		},
	}
}
