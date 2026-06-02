package port

import (
	"context"

	"github.com/DoMinhHHung/R2/internal/domain/entity"
)

type RBACChecker interface {
	Check(ctx context.Context, perm entity.Permission, roles []string) bool
}
