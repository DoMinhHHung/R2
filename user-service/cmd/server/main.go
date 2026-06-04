package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/DoMinhHHung/user-service/internal/config"
	"github.com/DoMinhHHung/user-service/internal/handler"
	"github.com/DoMinhHHung/user-service/internal/infrastructure/postgres"
	"github.com/DoMinhHHung/user-service/internal/infrastructure/rabbitmq"
	redisinfra "github.com/DoMinhHHung/user-service/internal/infrastructure/redis"
	"github.com/DoMinhHHung/user-service/internal/infrastructure/storage"
	"github.com/DoMinhHHung/user-service/internal/logger"
	"github.com/DoMinhHHung/user-service/internal/repository"
	"github.com/DoMinhHHung/user-service/internal/router"
	"github.com/DoMinhHHung/user-service/internal/service"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	appLog := logger.New(cfg.App.Env)
	ctx := context.Background()

	db, err := postgres.New(ctx, cfg.Database)
	if err != nil {
		log.Fatalf("connect db: %v", err)
	}
	defer db.Close()

	rdb, err := redisinfra.New(cfg.Redis)
	if err != nil {
		appLog.Warn("redis unavailable, caching disabled", "error", err)
	}
	if rdb != nil {
		defer rdb.Close()
	}

	store, err := storage.New(cfg.Cloudinary)
	if err != nil {
		log.Fatalf("init storage: %v", err)
	}

	userRepo := repository.New(db)
	userSvc := service.New(userRepo, store, appLog, rdb)

	userH := handler.NewUser(userSvc)
	adminH := handler.NewAdmin(userSvc)

	eventHandler := rabbitmq.NewUserEventHandler(userSvc)
	consumer, err := rabbitmq.NewConsumer(cfg.RabbitMQ, eventHandler, appLog)
	if err != nil {
		appLog.Warn("rabbitmq unavailable, event consumer disabled", "error", err)
	}
	if consumer != nil {
		defer consumer.Close()
		if err := consumer.Setup(); err != nil {
			appLog.Error("rabbitmq setup failed", "error", err)
		} else if err := consumer.Start(ctx); err != nil {
			appLog.Error("rabbitmq consumer start failed", "error", err)
		}
	}

	if cfg.App.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := router.New(router.Deps{
		UserHandler:  userH,
		AdminHandler: adminH,
		Log:          appLog,
		Cfg:          cfg,
	})

	srv := &http.Server{
		Addr:         ":" + cfg.App.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		appLog.Info("user-service started", "port", cfg.App.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			appLog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	appLog.Info("shutting down gracefully...")
	shutCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutCtx); err != nil {
		appLog.Error("shutdown error", "error", err)
	}
	appLog.Info("user-service stopped")
}
