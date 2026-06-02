package cache

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

type LuaRateLimiter struct {
	client *redis.Client
	script *redis.Script
}

func NewLuaRateLimiter(client *redis.Client) (*LuaRateLimiter, error) {
	luaBytes, err := os.ReadFile("scripts/rate_limit.lua")
	if err != nil {
		return nil, fmt.Errorf("read lua script: %w", err)
	}

	script := redis.NewScript(string(luaBytes))
	return &LuaRateLimiter{client: client, script: script}, nil
}

type LuaRLResult struct {
	Allowed      bool
	CurrentCount int64
	RetryAfterMS int64
}

func (l *LuaRateLimiter) Allow(ctx context.Context, key string, limit int, windowSeconds int) (*LuaRLResult, error) {
	now := time.Now().UnixMilli()

	res, err := l.script.Run(ctx, l.client,
		[]string{key},
		limit,
		windowSeconds,
		now,
	).Slice()
	if err != nil {
		return &LuaRLResult{Allowed: true}, nil
	}

	allowed := res[0].(int64) == 1
	count := res[1].(int64)
	retryAfter := res[2].(int64)

	return &LuaRLResult{
		Allowed:      allowed,
		CurrentCount: count,
		RetryAfterMS: retryAfter,
	}, nil
}
