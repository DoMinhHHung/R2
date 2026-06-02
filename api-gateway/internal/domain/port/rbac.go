package port

import (
	"context"
	"github.com/DoMinhHHung/Rental/internal/domain/entity"
)

type RBACChecker interface {
	Check(ctx context.Context, perm entity.Permission, roles []string) bool
}
