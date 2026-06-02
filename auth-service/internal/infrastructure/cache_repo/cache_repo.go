package cacherepo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/DoMinhHHung/auth-service/internal/domain/entity"
	"github.com/DoMinhHHung/auth-service/internal/domain/port"
	"github.com/DoMinhHHung/auth-service/internal/infrastructure/cache"
	"github.com/DoMinhHHung/auth-service/pkg/apperr"
	"github.com/redis/go-redis/v9"
)

const (
	signupKeyPrefix  = "signup:"
	forgotKeyPrefix  = "forgot:"
	resetTokenPrefix = "reset_token:"
	signupTTL        = 5 * time.Minute
	recoveryTTL      = 5 * time.Minute
	resetTokenTTL    = 10 * time.Minute
)

type cacheRepo struct {
	rdb *cache.RedisClient
}

func New(rdb *cache.RedisClient) port.CacheRepository {
	return &cacheRepo{rdb: rdb}
}

// ─── Signup Session ───────────────────────────────────────────────────────────

func (r *cacheRepo) SetSignupSession(ctx context.Context, s *entity.SignupSession) error {
	data, err := json.Marshal(s)
	if err != nil {
		return fmt.Errorf("marshal signup session: %w", err)
	}
	return r.rdb.Set(ctx, signupKeyPrefix+s.Email, string(data), signupTTL)
}

func (r *cacheRepo) GetSignupSession(ctx context.Context, email string) (*entity.SignupSession, error) {
	val, err := r.rdb.Get(ctx, signupKeyPrefix+email)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, apperr.ErrOTPExpired
		}
		return nil, fmt.Errorf("get signup session: %w", err)
	}

	var s entity.SignupSession
	if err := json.Unmarshal([]byte(val), &s); err != nil {
		return nil, fmt.Errorf("unmarshal signup session: %w", err)
	}
	return &s, nil
}

func (r *cacheRepo) UpdateSignupSession(ctx context.Context, s *entity.SignupSession) error {
	data, err := json.Marshal(s)
	if err != nil {
		return fmt.Errorf("marshal signup session: %w", err)
	}
	return r.rdb.Set(ctx, signupKeyPrefix+s.Email, string(data), 0)
}

func (r *cacheRepo) DeleteSignupSession(ctx context.Context, email string) error {
	return r.rdb.Del(ctx, signupKeyPrefix+email)
}

// ─── Recovery Session ─────────────────────────────────────────────────────────

func (r *cacheRepo) SetRecoverySession(ctx context.Context, s *entity.RecoverySession) error {
	data, err := json.Marshal(s)
	if err != nil {
		return fmt.Errorf("marshal recovery session: %w", err)
	}
	return r.rdb.Set(ctx, forgotKeyPrefix+s.Email, string(data), recoveryTTL)
}

func (r *cacheRepo) GetRecoverySession(ctx context.Context, email string) (*entity.RecoverySession, error) {
	val, err := r.rdb.Get(ctx, forgotKeyPrefix+email)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, apperr.ErrOTPExpired
		}
		return nil, fmt.Errorf("get recovery session: %w", err)
	}

	var s entity.RecoverySession
	if err := json.Unmarshal([]byte(val), &s); err != nil {
		return nil, fmt.Errorf("unmarshal recovery session: %w", err)
	}
	return &s, nil
}

func (r *cacheRepo) UpdateRecoverySession(ctx context.Context, s *entity.RecoverySession) error {
	data, err := json.Marshal(s)
	if err != nil {
		return fmt.Errorf("marshal recovery session: %w", err)
	}
	return r.rdb.Set(ctx, forgotKeyPrefix+s.Email, string(data), 0)
}

func (r *cacheRepo) DeleteRecoverySession(ctx context.Context, email string) error {
	return r.rdb.Del(ctx, forgotKeyPrefix+email)
}

// ─── Reset Token ──────────────────────────────────────────────────────────────

func (r *cacheRepo) SetResetToken(ctx context.Context, token, email string) error {
	return r.rdb.Set(ctx, resetTokenPrefix+token, email, resetTokenTTL)
}

func (r *cacheRepo) GetResetToken(ctx context.Context, token string) (string, error) {
	email, err := r.rdb.Get(ctx, resetTokenPrefix+token)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", apperr.ErrResetTokenInvalid
		}
		return "", err
	}
	return email, nil
}

func (r *cacheRepo) DeleteResetToken(ctx context.Context, token string) error {
	return r.rdb.Del(ctx, resetTokenPrefix+token)
}
