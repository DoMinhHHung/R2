package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/DoMinhHHung/user-service/internal/config"
	"github.com/redis/go-redis/v9"
)

const (
	userProfileTTL = 5 * time.Minute
	profileKeyFmt  = "user:profile:%s"
)

type Cache struct{ client *redis.Client }

func New(cfg config.RedisConfig) (*Cache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("ping redis: %w", err)
	}
	return &Cache{client: client}, nil
}

func (c *Cache) Close() error { return c.client.Close() }

func (c *Cache) GetUserProfile(ctx context.Context, userID string) ([]byte, error) {
	val, err := c.client.Get(ctx, fmt.Sprintf(profileKeyFmt, userID)).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, err
	}
	return val, nil
}

func (c *Cache) SetUserProfile(ctx context.Context, userID string, data any) error {
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, fmt.Sprintf(profileKeyFmt, userID), b, userProfileTTL).Err()
}

func (c *Cache) InvalidateUserProfile(ctx context.Context, userID string) {
	_ = c.client.Del(ctx, fmt.Sprintf(profileKeyFmt, userID)).Err()
}
