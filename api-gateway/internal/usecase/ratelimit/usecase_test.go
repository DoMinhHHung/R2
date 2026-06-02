package ratelimit_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DoMinhHHung/R2/internal/domain/entity"
	"github.com/DoMinhHHung/R2/internal/infrastructure/config"
	"github.com/DoMinhHHung/R2/internal/usecase/ratelimit"
)

// mockCache implements port.Cache for testing.
type mockCache struct {
	counts  map[string]int64
	incrErr error
}

func newMockCache() *mockCache {
	return &mockCache{counts: make(map[string]int64)}
}

func (m *mockCache) Incr(_ context.Context, key string) (int64, error) {
	if m.incrErr != nil {
		return 0, m.incrErr
	}
	m.counts[key]++
	return m.counts[key], nil
}

func (m *mockCache) Expire(_ context.Context, _ string, _ time.Duration) error {
	return nil
}

func (m *mockCache) Get(_ context.Context, _ string) (string, error) {
	return "", nil
}

func (m *mockCache) Set(_ context.Context, _ string, _ string, _ time.Duration) error {
	return nil
}

func (m *mockCache) TTL(_ context.Context, _ string) (time.Duration, error) {
	return 0, nil
}

func defaultRLConfig() config.RateLimitConfig {
	return config.RateLimitConfig{
		IP:            5,
		User:          10,
		APIKey:        20,
		WindowSeconds: 60,
	}
}

func TestAllow_BelowLimit(t *testing.T) {
	cache := newMockCache()
	uc := ratelimit.New(cache, defaultRLConfig())

	key := entity.RateLimitKey{Type: "ip", Value: "1.2.3.4"}
	for i := 0; i < 5; i++ {
		allowed, err := uc.Allow(context.Background(), key)
		if err != nil {
			t.Fatalf("Allow() error = %v", err)
		}
		if !allowed {
			t.Errorf("Allow() call %d should be allowed (below limit 5)", i+1)
		}
	}
}

func TestAllow_AtLimit(t *testing.T) {
	cache := newMockCache()
	uc := ratelimit.New(cache, defaultRLConfig())

	key := entity.RateLimitKey{Type: "ip", Value: "1.2.3.4"}
	// Use up all 5 allowed requests
	for i := 0; i < 5; i++ {
		uc.Allow(context.Background(), key)
	}
	// 6th request should be denied (count=6 > limit=5)
	allowed, err := uc.Allow(context.Background(), key)
	if err != nil {
		t.Fatalf("Allow() error = %v", err)
	}
	if allowed {
		t.Error("Allow() 6th call should be denied (exceeds limit 5)")
	}
}

func TestAllow_CacheError_FailsOpen(t *testing.T) {
	cache := newMockCache()
	cache.incrErr = errors.New("redis unavailable")
	uc := ratelimit.New(cache, defaultRLConfig())

	key := entity.RateLimitKey{Type: "ip", Value: "1.2.3.4"}
	allowed, err := uc.Allow(context.Background(), key)
	if err != nil {
		t.Fatalf("Allow() returned unexpected error = %v", err)
	}
	if !allowed {
		t.Error("Allow() should fail open (return true) when cache errors")
	}
}

func TestAllow_UserLimit(t *testing.T) {
	cache := newMockCache()
	cfg := config.RateLimitConfig{IP: 1, User: 3, APIKey: 5, WindowSeconds: 60}
	uc := ratelimit.New(cache, cfg)

	key := entity.RateLimitKey{Type: "user", Value: "user-123"}
	for i := 0; i < 3; i++ {
		allowed, _ := uc.Allow(context.Background(), key)
		if !allowed {
			t.Errorf("Allow() call %d should be allowed for user (limit=3)", i+1)
		}
	}
	allowed, _ := uc.Allow(context.Background(), key)
	if allowed {
		t.Error("Allow() 4th call should be denied for user (limit=3)")
	}
}

func TestAllow_APIKeyLimit(t *testing.T) {
	cache := newMockCache()
	cfg := config.RateLimitConfig{IP: 1, User: 2, APIKey: 4, WindowSeconds: 60}
	uc := ratelimit.New(cache, cfg)

	key := entity.RateLimitKey{Type: "apikey", Value: "key-abc"}
	for i := 0; i < 4; i++ {
		allowed, _ := uc.Allow(context.Background(), key)
		if !allowed {
			t.Errorf("Allow() call %d should be allowed for apikey (limit=4)", i+1)
		}
	}
	allowed, _ := uc.Allow(context.Background(), key)
	if allowed {
		t.Error("Allow() 5th call should be denied for apikey (limit=4)")
	}
}

func TestAllow_IPLimit_Default(t *testing.T) {
	cache := newMockCache()
	cfg := config.RateLimitConfig{IP: 2, User: 10, APIKey: 20, WindowSeconds: 60}
	uc := ratelimit.New(cache, cfg)

	// Unknown type falls back to IP limit
	key := entity.RateLimitKey{Type: "unknown_type", Value: "x"}
	for i := 0; i < 2; i++ {
		allowed, _ := uc.Allow(context.Background(), key)
		if !allowed {
			t.Errorf("Allow() call %d should be allowed for default IP limit", i+1)
		}
	}
	allowed, _ := uc.Allow(context.Background(), key)
	if allowed {
		t.Error("Allow() 3rd call should be denied (default IP limit = 2)")
	}
}

func TestAllow_DifferentKeysIndependent(t *testing.T) {
	cache := newMockCache()
	cfg := config.RateLimitConfig{IP: 1, User: 10, APIKey: 20, WindowSeconds: 60}
	uc := ratelimit.New(cache, cfg)

	key1 := entity.RateLimitKey{Type: "ip", Value: "1.1.1.1"}
	key2 := entity.RateLimitKey{Type: "ip", Value: "2.2.2.2"}

	// Exhaust key1
	uc.Allow(context.Background(), key1)
	denied, _ := uc.Allow(context.Background(), key1)
	if denied {
		t.Error("key1 should be denied on 2nd call")
	}

	// key2 should still be allowed
	allowed, _ := uc.Allow(context.Background(), key2)
	if !allowed {
		t.Error("key2 should be allowed (independent from key1)")
	}
}