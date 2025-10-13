package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"news-aggregator/internal/models"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

type Client struct {
	rdb *redis.Client
}

func NewClient(redisURL string) *Client {
	rdb := redis.NewClient(&redis.Options{
		Addr: redisURL,
	})

	return &Client{
		rdb: rdb,
	}
}

func (c *Client) StoreNews(ctx context.Context, article models.NewsArticle) error {
	key := fmt.Sprintf("news:%s", article.ID)
	
	articleJSON, err := json.Marshal(article)
	if err != nil {
		return fmt.Errorf("failed to marshal article: %w", err)
	}

	// Store with expiration (24 hours)
	err = c.rdb.Set(ctx, key, articleJSON, 24*time.Hour).Err()
	if err != nil {
		return fmt.Errorf("failed to store article: %w", err)
	}

	// Add to category index
	if article.Category != "" {
		categoryKey := fmt.Sprintf("category:%s", article.Category)
		c.rdb.SAdd(ctx, categoryKey, article.ID)
		c.rdb.Expire(ctx, categoryKey, 24*time.Hour)
	}

	// Add to latest news list
	c.rdb.LPush(ctx, "latest_news", article.ID)
	c.rdb.LTrim(ctx, "latest_news", 0, 999) // Keep only latest 1000 articles
	c.rdb.Expire(ctx, "latest_news", 24*time.Hour)

	logrus.WithField("article_id", article.ID).Info("Successfully stored news article")
	return nil
}

func (c *Client) GetNews(ctx context.Context, id string) (*models.NewsArticle, error) {
	key := fmt.Sprintf("news:%s", id)
	
	val, err := c.rdb.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("article not found")
		}
		return nil, fmt.Errorf("failed to get article: %w", err)
	}

	var article models.NewsArticle
	if err := json.Unmarshal([]byte(val), &article); err != nil {
		return nil, fmt.Errorf("failed to unmarshal article: %w", err)
	}

	return &article, nil
}

func (c *Client) GetLatestNews(ctx context.Context, limit, offset int) ([]models.NewsArticle, error) {
	// Get article IDs from latest news list
	ids, err := c.rdb.LRange(ctx, "latest_news", int64(offset), int64(offset+limit-1)).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get latest news IDs: %w", err)
	}

	var articles []models.NewsArticle
	for _, id := range ids {
		article, err := c.GetNews(ctx, id)
		if err != nil {
			logrus.WithError(err).WithField("article_id", id).Warn("Failed to get article")
			continue
		}
		articles = append(articles, *article)
	}

	return articles, nil
}

func (c *Client) GetNewsByCategory(ctx context.Context, category string, limit, offset int) ([]models.NewsArticle, error) {
	categoryKey := fmt.Sprintf("category:%s", category)
	
	// Get article IDs from category
	ids, err := c.rdb.SMembers(ctx, categoryKey).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get category news IDs: %w", err)
	}

	// Apply pagination
	start := offset
	end := offset + limit
	if start >= len(ids) {
		return []models.NewsArticle{}, nil
	}
	if end > len(ids) {
		end = len(ids)
	}

	var articles []models.NewsArticle
	for _, id := range ids[start:end] {
		article, err := c.GetNews(ctx, id)
		if err != nil {
			logrus.WithError(err).WithField("article_id", id).Warn("Failed to get article")
			continue
		}
		articles = append(articles, *article)
	}

	return articles, nil
}

func (c *Client) GetClient() *redis.Client {
	return c.rdb
}

func (c *Client) Close() error {
	return c.rdb.Close()
}
