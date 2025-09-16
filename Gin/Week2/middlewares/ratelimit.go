package middlewares

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

var limiter = rate.NewLimiter(5, 3) // 1 request/giây, cho phép burst = 3

func RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "Quá nhiều request, vui lòng thử lại"})
			c.Abort()
			return
		}
		c.Next()
	}
}
