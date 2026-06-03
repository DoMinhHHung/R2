package middleware

import (
	"time"

	"github.com/DoMinhHHung/user-service/internal/logger"
	"github.com/gin-gonic/gin"
)

func Logging(log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		rid, _ := c.Get("request_id")
		uid, _ := c.Get(CtxUserID)

		log.Info("request",
			"request_id", rid,
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"latency_ms", time.Since(start).Milliseconds(),
			"user_id", uid,
			"ip", c.ClientIP(),
		)
	}
}
