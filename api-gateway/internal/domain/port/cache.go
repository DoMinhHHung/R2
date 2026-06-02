package port

import (
	"context"
	"time"
)

type Cache interface {
	Incr(ctx context.Context, key string) (int64, error)
	Expire(ctx context.Context, key string, ttl time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value string, ttl time.Duration) error
	TTL(ctx context.Context, key string) (time.Duration, error)
}
