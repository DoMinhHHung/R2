package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/DoMinhHHung/R2/internal/adapter/handler"
	"github.com/DoMinhHHung/R2/internal/adapter/middleware"
	"github.com/DoMinhHHung/R2/internal/infrastructure/cache"
	"github.com/DoMinhHHung/R2/internal/infrastructure/config"
	infraJWT "github.com/DoMinhHHung/R2/internal/infrastructure/jwt"
	"github.com/DoMinhHHung/R2/internal/infrastructure/logger"
	"github.com/DoMinhHHung/R2/internal/infrastructure/metrics"
	"github.com/DoMinhHHung/R2/internal/infrastructure/tracer"
	"github.com/DoMinhHHung/R2/internal/usecase/proxy"
	"github.com/DoMinhHHung/R2/internal/usecase/ratelimit"
	"github.com/DoMinhHHung/R2/internal/usecase/rbac"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

const maxBodyBytes = 4 << 20

// @title           Rental Platform API Gateway
// @version         1.0
// @description     API Gateway for Rental Management Platform
// @host            localhost:8080
// @BasePath        /api/v1
// @securityDefinitions.apikey BearerAuth
// @in              header
// @name            Authorization
func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	rules, err := config.LoadRules()
	if err != nil {
		log.Fatalf("load gateway rules: %v", err)
	}

	appLog := logger.New(cfg.App.Env)

	otelTracer, err := tracer.New(cfg.OTEL)
	if err != nil {
		appLog.Warn("tracer init failed, continuing without tracing", "error", err)
	}

	rdb := cache.NewRedis(cfg.Redis)
	if err := rdb.Ping(context.Background()); err != nil {
		appLog.Error("redis connection failed", "error", err)
	}

	luaLimiter, err := cache.NewLuaRateLimiter(rdb.Client())
	if err != nil {
		appLog.Warn("lua rate limiter init failed, fallback to simple RL", "error", err)
	}

	met := metrics.New(cfg.OTEL.ServiceName)
	jwtParser, err := infraJWT.NewParser(cfg.JWT.PublicKeyPath, cfg.JWT.Issuer)
	if err != nil {
		log.Fatalf("init jwt parser: %v", err)
	}
	rbacUC := rbac.New(rules.RBACRules)
	proxyUC := proxy.NewWithRules(*cfg, rules.Routes)
	proxyHandler := handler.NewProxy(proxyUC)
	healthHandler := handler.NewHealth(rdb)

	if cfg.App.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.SecurityHeaders())
	r.Use(middleware.RequestID())
	r.Use(middleware.MaxBodySize(maxBodyBytes))
	r.Use(middleware.Logging(appLog))
	r.Use(met.Middleware())

	if otelTracer != nil {
		r.Use(otelgin.Middleware(cfg.OTEL.ServiceName))
		r.Use(middleware.Tracing(otelTracer))
	}

	if luaLimiter != nil {
		luaUC := ratelimit.NewLua(luaLimiter, cfg.RateLimit)
		r.Use(middleware.RateLimitLua(luaUC))
	} else {
		rlUC := ratelimit.New(rdb, cfg.RateLimit)
		r.Use(middleware.RateLimit(rlUC))
	}

	r.GET("/healthz/live", healthHandler.Live)
	r.GET("/healthz/ready", healthHandler.Ready)
	r.GET("/metrics", met.Handler())
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	v1 := r.Group("/api/v1")
	{
		v1.Any("/auth/*path",
			middleware.OptionalAuth(jwtParser),
			proxyHandler.Handle,
		)

		v1.GET("/properties/*path",
			middleware.OptionalAuth(jwtParser),
			proxyHandler.Handle,
		)
		v1.POST("/properties/*path",
			middleware.Auth(jwtParser),
			middleware.RBAC(rbacUC),
			proxyHandler.Handle,
		)
		v1.PUT("/properties/*path",
			middleware.Auth(jwtParser),
			middleware.RBAC(rbacUC),
			proxyHandler.Handle,
		)
		v1.PATCH("/properties/*path",
			middleware.Auth(jwtParser),
			middleware.RBAC(rbacUC),
			proxyHandler.Handle,
		)
		v1.DELETE("/properties/*path",
			middleware.Auth(jwtParser),
			middleware.RBAC(rbacUC),
			proxyHandler.Handle,
		)

		protected := v1.Group("")
		protected.Use(middleware.Auth(jwtParser), middleware.RBAC(rbacUC))
		protected.Any("/users/*path", proxyHandler.Handle)
		protected.Any("/bookings/*path", proxyHandler.Handle)
		protected.Any("/payments/*path", proxyHandler.Handle)
		protected.Any("/notifications/*path", proxyHandler.Handle)
	}

	httpSrv := &http.Server{
		Addr:         ":" + cfg.App.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		appLog.Info("gateway HTTP started", "port", cfg.App.Port)
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			appLog.Error("http server error", "error", err)
		}
	}()

	if cfg.App.TLSCert != "" && cfg.App.TLSKey != "" {
		tlsSrv := &http.Server{
			Addr:         ":8443",
			Handler:      r,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		}
		go func() {
			appLog.Info("gateway HTTPS started", "port", "8443")
			if err := tlsSrv.ListenAndServeTLS(cfg.App.TLSCert, cfg.App.TLSKey); err != nil && err != http.ErrServerClosed {
				appLog.Error("https server error", "error", err)
			}
		}()
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	appLog.Info("shutting down gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := httpSrv.Shutdown(ctx); err != nil {
		appLog.Error("shutdown error", "error", err)
	}

	if otelTracer != nil {
		otelTracer.Shutdown(ctx)
	}

	appLog.Info("gateway stopped")
}
