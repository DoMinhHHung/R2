package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/DoMinhHHung/auth-service/internal/infrastructure/cache"
	"github.com/gin-gonic/gin"
)

// key format: "rl:{endpoint}:{ip}"
func RateLimit(rdb *cache.RedisClient, endpoint string, limit int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		key := fmt.Sprintf("rl:%s:%s", endpoint, ip)

		count, err := rdb.IncrWithTTL(c.Request.Context(), key, window)
		if err != nil {
			c.Next()
			return
		}

		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", limit))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", max(0, int64(limit)-count)))
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(window).Unix()))

		if count > int64(limit) {
			c.Header("Retry-After", fmt.Sprintf("%d", int(window.Seconds())))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"code":    "RATE_LIMIT_EXCEEDED",
				"message": "Too many requests",
			})
			return
		}

		c.Next()
	}
}

func max(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
