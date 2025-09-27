package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/patrickmn/go-cache"
)

type RateLimiter struct {
	cache *cache.Cache
}

func NewRateLimiter() *RateLimiter {
	return &RateLimiter{cache: cache.New(1*time.Minute, 2*time.Minute)}
}

func (rl *RateLimiter) Limit(max int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.ClientIP()
		now := time.Now()

		reqs, found := rl.cache.Get(key)
		var times []time.Time
		if found {
			times = reqs.([]time.Time)
		}

		cutoff := now.Add(-window)
		var valid []time.Time
		for _, t := range times {
			if t.After(cutoff) {
				valid = append(valid, t)
			}
		}

		if len(valid) >= max {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "Rate limit exceeded"})
			c.Abort()
			return
		}

		valid = append(valid, now)
		rl.cache.Set(key, valid, window)
		c.Next()
	}
}
