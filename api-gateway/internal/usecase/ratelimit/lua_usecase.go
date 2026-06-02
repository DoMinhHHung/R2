package ratelimit

import (
	"context"
	"fmt"

	"github.com/DoMinhHHung/R2/internal/domain/entity"
	"github.com/DoMinhHHung/R2/internal/infrastructure/cache"
	"github.com/DoMinhHHung/R2/internal/infrastructure/config"
)

type LuaUseCase struct {
	limiter *cache.LuaRateLimiter
	cfg     config.RateLimitConfig
}

func NewLua(limiter *cache.LuaRateLimiter, cfg config.RateLimitConfig) *LuaUseCase {
	return &LuaUseCase{limiter: limiter, cfg: cfg}
}

func (u *LuaUseCase) Allow(ctx context.Context, key entity.RateLimitKey) (bool, int64, error) {
	limit := u.limitFor(key.Type)
	redisKey := fmt.Sprintf("rl:%s:%s", key.Type, key.Value)

	res, err := u.limiter.Allow(ctx, redisKey, limit, u.cfg.WindowSeconds)
	if err != nil {
		return true, 0, err
	}

	return res.Allowed, res.RetryAfterMS, nil
}

func (u *LuaUseCase) limitFor(t string) int {
	switch t {
	case "user":
		return u.cfg.User
	case "apikey":
		return u.cfg.APIKey
	default:
		return u.cfg.IP
	}
}
