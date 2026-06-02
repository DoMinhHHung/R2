package middleware

import (
	"github.com/DoMinhHHung/Rental/internal/domain/entity"
	"github.com/DoMinhHHung/Rental/internal/domain/port"
	"github.com/gin-gonic/gin"
)

func RateLimit(rl port.RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		keys := buildKeys(c)
		for _, k := range keys {
			allowed, err := rl.Allow(ctx, k)
			if err != nil {
				c.Next()
				return
			}
			if !allowed {
				c.AbortWithStatusJSON(429, gin.H{
					"error":      "rate_limit_exceeded",
					"limit_type": k.Type,
				})
				return
			}
		}

		c.Next()
	}
}

func buildKeys(c *gin.Context) []entity.RateLimitKey {
	keys := []entity.RateLimitKey{
		{Type: "ip", Value: c.ClientIP()},
	}

	if uid := c.GetHeader("X-User-ID"); uid != "" {
		keys = append(keys, entity.RateLimitKey{Type: "user", Value: uid})
	}

	if apiKey := c.GetHeader("X-API-Key"); apiKey != "" {
		keys = append(keys, entity.RateLimitKey{Type: "apikey", Value: apiKey})
	}

	return keys
}
