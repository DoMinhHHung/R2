package middleware

import (
	"net/http"

	"github.com/DoMinhHHung/user-service/internal/logger"
	"github.com/gin-gonic/gin"
)

func Recovery(log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if rec := recover(); rec != nil {
				rid, _ := c.Get("request_id")
				log.Error("panic recovered",
					"request_id", rid,
					"error", rec,
					"path", c.Request.URL.Path,
				)
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"code":    "INTERNAL_ERROR",
					"message": "An unexpected error occurred",
				})
			}
		}()
		c.Next()
	}
}
