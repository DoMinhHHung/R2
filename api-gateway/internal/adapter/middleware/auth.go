package middleware

import (
	"context"
	"strings"

	"github.com/DoMinhHHung/R2/internal/domain/entity"
	"github.com/DoMinhHHung/R2/internal/domain/port"
	infraJWT "github.com/DoMinhHHung/R2/internal/infrastructure/jwt"
	proxyUC "github.com/DoMinhHHung/R2/internal/usecase/proxy"
	"github.com/gin-gonic/gin"
)

func Auth(parser *infraJWT.Parser) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(401, gin.H{"error": "missing_token"})
			return
		}

		claims, err := parser.Parse(authHeader)
		if err != nil {
			c.AbortWithStatusJSON(401, gin.H{"error": "invalid_token", "detail": err.Error()})
			return
		}

		setClaims(c, claims)

		c.Next()
	}
}

func OptionalAuth(parser *infraJWT.Parser) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
			if claims, err := parser.Parse(authHeader); err == nil {
				setClaims(c, claims)
			}
		}
		c.Next()
	}
}

func setClaims(c *gin.Context, claims *entity.Claims) {
	c.Set("claims", claims)
	c.Set("user_id", claims.UserID)
	c.Set("user_roles", claims.Roles)

	ctx := context.WithValue(c.Request.Context(), proxyUC.ValidatedClaimsKey(), claims)
	c.Request = c.Request.WithContext(ctx)
}

func RBAC(checker port.RBACChecker) gin.HandlerFunc {
	return func(c *gin.Context) {
		claimsRaw, exists := c.Get("claims")
		if !exists {
			c.AbortWithStatusJSON(403, gin.H{"error": "forbidden"})
			return
		}

		claims, ok := claimsRaw.(*entity.Claims)
		if !ok {
			c.AbortWithStatusJSON(403, gin.H{"error": "forbidden"})
			return
		}

		svc := resolveService(c.Request.URL.Path)
		perm := entity.Permission{
			Service: svc,
			Method:  c.Request.Method,
			Path:    c.Request.URL.Path,
		}

		allowed := checker.Check(c.Request.Context(), perm, claims.Roles)
		if !allowed {
			c.AbortWithStatusJSON(403, gin.H{"error": "forbidden", "required_service": svc})
			return
		}

		c.Next()
	}
}

func resolveService(path string) string {
	parts := strings.Split(strings.TrimPrefix(path, "/"), "/")
	if len(parts) >= 3 {
		return parts[2]
	}
	return ""
}
