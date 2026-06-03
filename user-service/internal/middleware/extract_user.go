package middleware

import (
	"strings"

	"github.com/DoMinhHHung/user-service/pkg/apperr"
	"github.com/DoMinhHHung/user-service/pkg/response"
	"github.com/gin-gonic/gin"
)

const (
	CtxUserID    = "user_id"
	CtxUserEmail = "user_email"
	CtxUserRoles = "user_roles"

	HeaderUserID    = "X-User-ID"
	HeaderUserEmail = "X-User-Email"
	HeaderUserRoles = "X-User-Roles"
)

func ExtractUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetHeader(HeaderUserID)
		if userID == "" {
			response.Error(c, apperr.ErrUnauthorized)
			c.Abort()
			return
		}

		rolesHeader := c.GetHeader(HeaderUserRoles)
		var roles []string
		if rolesHeader != "" {
			roles = strings.Split(rolesHeader, ",")
		}

		c.Set(CtxUserID, userID)
		c.Set(CtxUserEmail, c.GetHeader(HeaderUserEmail))
		c.Set(CtxUserRoles, roles)
		c.Next()
	}
}

func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		roles := GetRoles(c)
		for _, r := range roles {
			if strings.EqualFold(r, "ADMIN") {
				c.Next()
				return
			}
		}
		response.Error(c, apperr.ErrForbidden)
		c.Abort()
	}
}

func GetUserID(c *gin.Context) string {
	v, _ := c.Get(CtxUserID)
	s, _ := v.(string)
	return s
}

func GetRoles(c *gin.Context) []string {
	v, _ := c.Get(CtxUserRoles)
	roles, _ := v.([]string)
	return roles
}
