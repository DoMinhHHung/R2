package middleware

import (
	"github.com/DoMinhHHung/auth-service/internal/domain/entity"
	"github.com/DoMinhHHung/auth-service/pkg/apperr"
	"github.com/DoMinhHHung/auth-service/pkg/response"
	"github.com/gin-gonic/gin"
)

func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		roleRaw, exists := c.Get("user_role")
		if !exists {
			response.Error(c, apperr.ErrForbidden)
			c.Abort()
			return
		}

		role, ok := roleRaw.(string)
		if !ok || entity.Role(role) != entity.RoleAdmin {
			response.Error(c, apperr.ErrForbidden)
			c.Abort()
			return
		}

		c.Next()
	}
}
