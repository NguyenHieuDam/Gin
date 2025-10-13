package models

import (
	"time"
)

// NewsArticle represents a news article
type NewsArticle struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Content     string    `json:"content"`
	Source      string    `json:"source"`
	URL         string    `json:"url"`
	PublishedAt time.Time `json:"published_at"`
	CreatedAt   time.Time `json:"created_at"`
	Category    string    `json:"category"`
}

// NewsRequest represents a request to get news
type NewsRequest struct {
	Category string `json:"category" form:"category"`
	Limit    int    `json:"limit" form:"limit"`
	Offset   int    `json:"offset" form:"offset"`
}

// NewsResponse represents the API response
type NewsResponse struct {
	Articles []NewsArticle `json:"articles"`
	Total    int           `json:"total"`
	Page     int           `json:"page"`
	Limit    int           `json:"limit"`
}
