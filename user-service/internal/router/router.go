package router

import (
	"github.com/DoMinhHHung/user-service/internal/config"
	"github.com/DoMinhHHung/user-service/internal/handler"
	"github.com/DoMinhHHung/user-service/internal/logger"
	"github.com/DoMinhHHung/user-service/internal/middleware"
	"github.com/DoMinhHHung/user-service/pkg/response"
	"github.com/gin-gonic/gin"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Deps struct {
	UserHandler  *handler.UserHandler
	AdminHandler *handler.AdminHandler
	Log          *logger.Logger
	Cfg          *config.Config
}

func New(deps Deps) *gin.Engine {
	r := gin.New()

	r.Use(middleware.Recovery(deps.Log))
	r.Use(middleware.RequestID())
	r.Use(middleware.Logging(deps.Log))

	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "user-service"})
	})
	r.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	internal := r.Group("/internal")
	internal.Use(internalAuth(deps.Cfg.Internal.Token))
	{
		internal.POST("/users/profile", deps.UserHandler.CreateProfileInternal)
	}

	api := r.Group("/api/v1")
	api.Use(middleware.ExtractUser())
	{
		users := api.Group("/users")
		{
			users.GET("/me", deps.UserHandler.GetMyProfile)
			users.PUT("/me", deps.UserHandler.UpdateMyProfile)
			users.POST("/avatar", deps.UserHandler.UploadAvatar)
			users.DELETE("/avatar", deps.UserHandler.DeleteAvatar)
			users.GET("/:id", deps.UserHandler.GetUserByID)
		}

		admin := api.Group("/admin")
		admin.Use(middleware.RequireAdmin())
		{
			admin.GET("/users", deps.AdminHandler.ListUsers)
			admin.GET("/users/:id", deps.AdminHandler.GetUser)
			admin.PATCH("/users/:id/ban", deps.AdminHandler.BanUser)
			admin.PATCH("/users/:id/unban", deps.AdminHandler.UnbanUser)
		}
	}

	return r
}

func internalAuth(token string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetHeader("X-Internal-Token") != token {
			c.AbortWithStatusJSON(403, gin.H{
				"success": false,
				"code":    "FORBIDDEN",
				"message": "Invalid internal token",
			})
			return
		}
		c.Next()
	}
}
