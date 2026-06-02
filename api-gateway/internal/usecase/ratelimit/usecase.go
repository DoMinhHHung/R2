package ratelimit

import (
	"context"
	"fmt"
	"time"

	"github.com/DoMinhHHung/R2/internal/domain/entity"
	"github.com/DoMinhHHung/R2/internal/domain/port"
	"github.com/DoMinhHHung/R2/internal/infrastructure/config"
)

type UseCase struct {
	cache port.Cache
	cfg   config.RateLimitConfig
}

func New(cache port.Cache, cfg config.RateLimitConfig) *UseCase {
	return &UseCase{cache: cache, cfg: cfg}
}

func (u *UseCase) Allow(ctx context.Context, key entity.RateLimitKey) (bool, error) {
	limit := u.limitFor(key.Type)
	window := time.Duration(u.cfg.WindowSeconds) * time.Second
	redisKey := fmt.Sprintf("rl:%s:%s", key.Type, key.Value)

	count, err := u.cache.Incr(ctx, redisKey)
	if err != nil {
		return true, nil
	}

	if count == 1 {
		u.cache.Expire(ctx, redisKey, window)
	}

	return count <= int64(limit), nil
}

func (u *UseCase) limitFor(t string) int {
	switch t {
	case "user":
		return u.cfg.User
	case "apikey":
		return u.cfg.APIKey
	default:
		return u.cfg.IP
	}
}
