package middleware

import (
	"fmt"
	"strconv"

	"github.com/DoMinhHHung/Rental/internal/domain/entity"
	"github.com/DoMinhHHung/Rental/internal/usecase/ratelimit"
	"github.com/gin-gonic/gin"
)

func RateLimitLua(rl *ratelimit.LuaUseCase) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		keys := buildKeys(c)

		for _, k := range keys {
			allowed, retryAfterMS, err := rl.Allow(ctx, k)
			if err != nil {
				c.Next()
				return
			}

			if !allowed {
				retryAfterSec := strconv.FormatInt(retryAfterMS/1000+1, 10)
				c.Header("Retry-After", retryAfterSec)
				c.Header("X-RateLimit-Type", k.Type)
				c.AbortWithStatusJSON(429, gin.H{
					"error":          "rate_limit_exceeded",
					"limit_type":     k.Type,
					"retry_after_ms": fmt.Sprintf("%d", retryAfterMS),
				})
				return
			}

			c.Header(fmt.Sprintf("X-RateLimit-%s-Key", k.Type), k.Value)
		}

		c.Next()
	}
}

func buildRLKey(k entity.RateLimitKey) string {
	return fmt.Sprintf("rl:%s:%s", k.Type, k.Value)
}
