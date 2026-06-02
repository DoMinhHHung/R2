package middleware

import (
	"github.com/DoMinhHHung/Rental/pkg/uuidv7"
	"github.com/gin-gonic/gin"
)

const HeaderRequestID = "X-Request-ID"

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		rid := c.GetHeader(HeaderRequestID)
		if rid == "" {
			rid = uuidv7.New()
		}
		c.Set("request_id", rid)
		c.Header(HeaderRequestID, rid)
		c.Next()
	}
}
