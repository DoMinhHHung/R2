package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	cacherepo "github.com/DoMinhHHung/auth-service/internal/infrastructure/cache_repo"
	"github.com/DoMinhHHung/auth-service/internal/infrastructure/email"
	"github.com/DoMinhHHung/auth-service/internal/infrastructure/hash"
	jwtinfra "github.com/DoMinhHHung/auth-service/internal/infrastructure/jwt"
	"github.com/DoMinhHHung/auth-service/internal/infrastructure/otp"
	"github.com/DoMinhHHung/auth-service/internal/infrastructure/repository"
	"github.com/DoMinhHHung/auth-service/internal/infrastructure/userclient"

	"github.com/DoMinhHHung/auth-service/internal/adapter/handler"
	"github.com/DoMinhHHung/auth-service/internal/adapter/router"
	"github.com/DoMinhHHung/auth-service/internal/application/usecase"
	"github.com/DoMinhHHung/auth-service/internal/infrastructure/cache"
	"github.com/DoMinhHHung/auth-service/internal/infrastructure/config"
	"github.com/DoMinhHHung/auth-service/internal/infrastructure/database"
	"github.com/DoMinhHHung/auth-service/internal/infrastructure/logger"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	appLog := logger.New(cfg.App.Env)

	ctx := context.Background()

	db, err := database.NewPool(ctx, cfg.Database)
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}
	defer db.Close()

	rdb, err := cache.NewRedisClient(cfg.Redis)
	if err != nil {
		log.Fatalf("connect redis: %v", err)
	}
	defer rdb.Close()

	tokenSvc, err := jwtinfra.New(cfg.JWT)
	if err != nil {
		log.Fatalf("init jwt service: %v", err)
	}

	authUserRepo := repository.NewAuthUserRepository(db)
	sessionRepo := repository.NewSessionRepository(db)
	cacheRepo := cacherepo.New(rdb)

	hasher := hash.NewArgon2idHasher()
	otpGen := otp.NewGenerator()
	emailSvc := email.NewSMTPService(cfg.Email)
	userClient := userclient.New(cfg.UserSvc.URL, cfg.UserSvc.InternalToken)

	signupUC := usecase.NewSignupUseCase(authUserRepo, cacheRepo, hasher, otpGen, emailSvc, userClient, cfg.OTP)
	loginUC := usecase.NewLoginUseCase(authUserRepo, sessionRepo, hasher, tokenSvc, cfg.JWT)
	tokenUC := usecase.NewTokenUseCase(authUserRepo, sessionRepo, tokenSvc, cfg.JWT)
	passwordUC := usecase.NewPasswordUseCase(authUserRepo, sessionRepo, cacheRepo, hasher, otpGen, emailSvc)
	sessionUC := usecase.NewSessionUseCase(sessionRepo)

	authH := handler.NewAuthHandler(signupUC, loginUC, tokenUC, passwordUC)
	sessionH := handler.NewSessionHandler(sessionUC)

	if cfg.App.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := router.New(router.Deps{
		AuthHandler:    authH,
		SessionHandler: sessionH,
		TokenSvc:       tokenSvc,
		Redis:          rdb,
		Logger:         appLog,
	})

	srv := &http.Server{
		Addr:         ":" + cfg.App.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		appLog.Info("auth service started", "port", cfg.App.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			appLog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	appLog.Info("shutting down gracefully...")
	shutdownCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		appLog.Error("shutdown error", "error", err)
	}
	appLog.Info("auth service stopped")
}
