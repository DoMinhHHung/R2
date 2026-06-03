package router

import (
	"time"

	"github.com/DoMinhHHung/auth-service/internal/adapter/handler"
	"github.com/DoMinhHHung/auth-service/internal/adapter/middleware"
	"github.com/DoMinhHHung/auth-service/internal/domain/port"
	"github.com/DoMinhHHung/auth-service/internal/infrastructure/cache"
	"github.com/DoMinhHHung/auth-service/internal/infrastructure/logger"
	"github.com/gin-gonic/gin"

	// _ "github.com/DoMinhHHung/auth-service/docs"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Deps struct {
	AuthHandler    *handler.AuthHandler
	SessionHandler *handler.SessionHandler
	AdminHandler   *handler.AdminHandler
	TokenSvc       port.TokenService
	Redis          *cache.RedisClient
	Logger         *logger.Logger
}

func New(deps Deps) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.RequestID())
	r.Use(middleware.Logging(deps.Logger))

	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "auth-service"})
	})

	r.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	rl5pm := middleware.RateLimit(deps.Redis, "general", 5, time.Minute)
	rl10pm := middleware.RateLimit(deps.Redis, "login", 10, time.Minute)
	rl3pm := middleware.RateLimit(deps.Redis, "otp", 3, time.Minute)
	rl20pm := middleware.RateLimit(deps.Redis, "refresh", 20, time.Minute)

	auth := r.Group("/api/v1/auth")
	{
		// Signup
		auth.POST("/signup", rl5pm, deps.AuthHandler.Signup)
		auth.POST("/signup/verify", rl5pm, deps.AuthHandler.VerifySignupOTP)
		auth.POST("/signup/resend", rl3pm, deps.AuthHandler.ResendSignupOTP)

		// Login
		auth.POST("/login", rl10pm, deps.AuthHandler.Login)

		// Token
		auth.POST("/refresh", rl20pm, deps.AuthHandler.RefreshToken)

		// Password recovery
		auth.POST("/forgot-password", rl3pm, deps.AuthHandler.ForgotPassword)
		auth.POST("/forgot-password/verify", rl3pm, deps.AuthHandler.VerifyRecoveryOTP)
		auth.POST("/forgot-password/resend", rl3pm, deps.AuthHandler.ResendForgotOTP)
		auth.POST("/reset-password", rl5pm, deps.AuthHandler.ResetPassword)

		// Protected routes — require valid access token
		protected := auth.Group("")
		protected.Use(middleware.RequireAuth(deps.TokenSvc))
		{
			protected.POST("/logout", deps.SessionHandler.Logout)
			protected.POST("/logout/all", deps.SessionHandler.LogoutAll)
			protected.GET("/sessions", deps.SessionHandler.GetActiveSessions)
			protected.DELETE("/sessions/:id", deps.SessionHandler.RevokeSession)
		}

		admin := auth.Group("/admin")
		{
			rl3p5m := middleware.RateLimit(deps.Redis, "admin-login", 3, 5*time.Minute)
			admin.POST("/login", rl3p5m, deps.AdminHandler.Login)

			adminProtected := admin.Group("")
			adminProtected.Use(
				middleware.RequireAuth(deps.TokenSvc),
				middleware.RequireAdmin(),
			)
			{
				adminProtected.POST("/users", deps.AdminHandler.CreateAdmin)
			}
		}
	}

	return r
}
