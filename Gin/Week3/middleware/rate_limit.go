package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

type RateLimiter struct {
	redis *redis.Client
	ctx   context.Context
}

func NewRateLimiter(redis *redis.Client) *RateLimiter {
	return &RateLimiter{
		redis: redis,
		ctx:   context.Background(),
	}
}

// RateLimitConfig holds configuration for rate limiting
type RateLimitConfig struct {
	Window time.Duration // Time window for rate limiting
	Limit  int           // Maximum number of requests per window
	Prefix string        // Redis key prefix
}

// DefaultRateLimitConfig returns default rate limiting configuration
func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		Window: time.Minute,
		Limit:  60, // 60 requests per minute
		Prefix: "rate_limit",
	}
}

// MessageRateLimitConfig returns rate limiting configuration for messages
func MessageRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		Window: time.Minute,
		Limit:  10, // 10 messages per minute
		Prefix: "message_rate",
	}
}

// RateLimitMiddleware creates a rate limiting middleware
func (rl *RateLimiter) RateLimitMiddleware(config RateLimitConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get client identifier (IP address or user ID)
		clientID := rl.getClientID(c)
		
		// Create Redis key
		key := fmt.Sprintf("%s:%s", config.Prefix, clientID)
		
		// Check current count
		current, err := rl.redis.Get(rl.ctx, key).Result()
		if err != nil && err != redis.Nil {
			// If Redis error, allow request (fail open)
			c.Next()
			return
		}
		
		var count int
		if current != "" {
			count, _ = strconv.Atoi(current)
		}
		
		// Check if limit exceeded
		if count >= config.Limit {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Quá nhiều yêu cầu. Vui lòng thử lại sau.",
				"retry_after": config.Window.Seconds(),
			})
			c.Abort()
			return
		}
		
		// Increment counter
		pipe := rl.redis.Pipeline()
		incr := pipe.Incr(rl.ctx, key)
		pipe.Expire(rl.ctx, key, config.Window)
		_, err = pipe.Exec(rl.ctx)
		
		if err != nil {
			// If Redis error, allow request (fail open)
			c.Next()
			return
		}
		
		// Set rate limit headers
		c.Header("X-RateLimit-Limit", strconv.Itoa(config.Limit))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(config.Limit-count-1))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(config.Window).Unix(), 10))
		
		c.Next()
	}
}

// getClientID extracts client identifier from request
func (rl *RateLimiter) getClientID(c *gin.Context) string {
	// Try to get user ID from context first (if authenticated)
	if userID, exists := c.Get("user_id"); exists {
		return fmt.Sprintf("user:%v", userID)
	}
	
	// Fall back to IP address
	return fmt.Sprintf("ip:%s", c.ClientIP())
}

// MessageRateLimitMiddleware creates a rate limiting middleware specifically for messages
func (rl *RateLimiter) MessageRateLimitMiddleware() gin.HandlerFunc {
	config := MessageRateLimitConfig()
	return rl.RateLimitMiddleware(config)
}

// LoginRateLimitMiddleware creates a rate limiting middleware for login attempts
func (rl *RateLimiter) LoginRateLimitMiddleware() gin.HandlerFunc {
	config := RateLimitConfig{
		Window: time.Minute * 5, // 5 minutes
		Limit:  5,               // 5 login attempts per 5 minutes
		Prefix: "login_rate",
	}
	return rl.RateLimitMiddleware(config)
}

// RegisterRateLimitMiddleware creates a rate limiting middleware for registration
func (rl *RateLimiter) RegisterRateLimitMiddleware() gin.HandlerFunc {
	config := RateLimitConfig{
		Window: time.Hour, // 1 hour
		Limit:  3,         // 3 registrations per hour per IP
		Prefix: "register_rate",
	}
	return rl.RateLimitMiddleware(config)
}

// WebSocketRateLimitMiddleware creates a rate limiting middleware for WebSocket connections
func (rl *RateLimiter) WebSocketRateLimitMiddleware() gin.HandlerFunc {
	config := RateLimitConfig{
		Window: time.Minute,
		Limit:  5, // 5 WebSocket connections per minute per IP
		Prefix: "ws_rate",
	}
	return rl.RateLimitMiddleware(config)
}
