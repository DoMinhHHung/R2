package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func MaxBodySize(maxBytes int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Body != nil {
			c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes)
		}
		c.Next()

		if c.Request.Body != nil {
			if err := c.Request.Body.Close(); err != nil {
				if err.Error() == "http: request body too large" {
					c.AbortWithStatusJSON(http.StatusRequestEntityTooLarge, gin.H{
						"error":  "request_too_large",
						"detail": "request body exceeds 4MB limit",
					})
				}
			}
		}
	}
}
