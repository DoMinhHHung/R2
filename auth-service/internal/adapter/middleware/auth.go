package middleware

import (
	"strings"

	"github.com/DoMinhHHung/auth-service/internal/domain/port"
	"github.com/DoMinhHHung/auth-service/pkg/apperr"
	"github.com/DoMinhHHung/auth-service/pkg/response"
	"github.com/gin-gonic/gin"
)

func RequireAuth(tokenSvc port.TokenService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			response.Error(c, apperr.ErrTokenInvalid)
			c.Abort()
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := tokenSvc.ParseToken(tokenStr)
		if err != nil {
			response.Error(c, apperr.ErrTokenInvalid)
			c.Abort()
			return
		}

		if claims.TokenType != "access" {
			response.Error(c, apperr.ErrTokenInvalid)
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_role", claims.Role)
		c.Set("session_id", claims.SessionID)
		c.Next()
	}
}
