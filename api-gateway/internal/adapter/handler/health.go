package handler

import (
	"context"
	"time"

	"github.com/DoMinhHHung/Rental/internal/infrastructure/cache"
	"github.com/gin-gonic/gin"
)

type HealthHandler struct {
	redis *cache.RedisCache
}

func NewHealth(redis *cache.RedisCache) *HealthHandler {
	return &HealthHandler{redis: redis}
}

func (h *HealthHandler) Live(c *gin.Context) {
	c.JSON(200, gin.H{"status": "ok"})
}

func (h *HealthHandler) Ready(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	status := "ok"
	code := 200
	redisOK := true

	if err := h.redis.Ping(ctx); err != nil {
		redisOK = false
		status = "degraded"
		code = 503
	}

	c.JSON(code, gin.H{
		"status": status,
		"deps": gin.H{
			"redis": redisOK,
		},
	})
}
