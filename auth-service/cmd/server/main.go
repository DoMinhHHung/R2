package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/DoMinhHHung/auth-service/internal/infrastructure/cache"
	"github.com/DoMinhHHung/auth-service/internal/infrastructure/config"
	"github.com/DoMinhHHung/auth-service/internal/infrastructure/database"
	"github.com/DoMinhHHung/auth-service/internal/infrastructure/logger"
	"github.com/gin-gonic/gin"
)

// @title           Auth Service API
// @version         1.0
// @description     Authentication & Authorization Service
// @host            localhost:8082
// @BasePath        /api/v1/auth
// @securityDefinitions.apikey BearerAuth
// @in              header
// @name            Authorization
func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	appLog := logger.New(cfg.App.Env)
	appLog.Info("starting auth service", "env", cfg.App.Env, "port", cfg.App.Port)

	// Database
	ctx := context.Background()
	db, err := database.NewPool(ctx, cfg.Database)
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}
	defer db.Close()
	appLog.Info("database connected")

	// Redis
	rdb, err := cache.NewRedisClient(cfg.Redis)
	if err != nil {
		log.Fatalf("connect redis: %v", err)
	}
	defer rdb.Close()
	appLog.Info("redis connected")

	if cfg.App.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "auth-service"})
	})

	srv := &http.Server{
		Addr:         ":" + cfg.App.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		appLog.Info("http server started", "port", cfg.App.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			appLog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	appLog.Info("shutting down gracefully...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		appLog.Error("shutdown error", "error", err)
	}
	appLog.Info("auth service stopped")
}
