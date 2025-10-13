package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type RateLimiter struct {
	redisClient *redis.Client
	requests    int
	window      int
}

func NewRateLimiter(redisClient *redis.Client, requests, window int) *RateLimiter {
	return &RateLimiter{
		redisClient: redisClient,
		requests:    requests,
		window:      window,
	}
}

func (rl *RateLimiter) RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		key := "rate_limit:" + clientIP

		// Get current count
		count, err := rl.redisClient.Get(c.Request.Context(), key).Result()
		if err != nil && err != redis.Nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Rate limit check failed"})
			c.Abort()
			return
		}

		// Parse count
		currentCount := 0
		if count != "" {
			currentCount, _ = strconv.Atoi(count)
		}

		// Check if limit exceeded
		if currentCount >= rl.requests {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Rate limit exceeded",
				"limit":       rl.requests,
				"window":      rl.window,
				"retry_after": rl.window,
			})
			c.Abort()
			return
		}

		// Increment counter
		pipe := rl.redisClient.Pipeline()
		pipe.Incr(c.Request.Context(), key)
		pipe.Expire(c.Request.Context(), key, time.Duration(rl.window)*time.Second)
		_, err = pipe.Exec(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Rate limit update failed"})
			c.Abort()
			return
		}

		// Add rate limit headers
		c.Header("X-RateLimit-Limit", strconv.Itoa(rl.requests))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(rl.requests-currentCount-1))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(time.Duration(rl.window)*time.Second).Unix(), 10))

		c.Next()
	}
}
