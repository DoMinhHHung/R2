package port

import (
	"context"

	"github.com/DoMinhHHung/R2/internal/domain/entity"
)

type RateLimiter interface {
	Allow(ctx context.Context, key entity.RateLimitKey) (bool, error)
}
