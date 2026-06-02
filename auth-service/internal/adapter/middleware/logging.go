package middleware

import (
	"time"

	"github.com/DoMinhHHung/auth-service/internal/infrastructure/logger"
	"github.com/gin-gonic/gin"
)

func Logging(log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		rid, _ := c.Get("request_id")
		userID, _ := c.Get("user_id")

		log.Info("request",
			"request_id", rid,
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"latency_ms", time.Since(start).Milliseconds(),
			"ip", c.ClientIP(),
			"user_id", userID,
		)
	}
}
