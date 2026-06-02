package port

import (
	"context"
	"github.com/DoMinhHHung/Rental/internal/domain/entity"
)

type RateLimiter interface {
	Allow(ctx context.Context, key entity.RateLimitKey) (bool, error)
}
