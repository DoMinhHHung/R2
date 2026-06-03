package port

import (
	"context"

	"github.com/DoMinhHHung/user-service/internal/domain/entity"
)

type UserRepository interface {
	Create(ctx context.Context, user *entity.User) error
	FindByID(ctx context.Context, id string) (*entity.User, error)
	FindByEmail(ctx context.Context, email string) (*entity.User, error)
	Update(ctx context.Context, user *entity.User) error
	UpdateAvatar(ctx context.Context, userID, avatarURL, publicID string) error
	DeleteAvatar(ctx context.Context, userID string) error
	UpdateStatus(ctx context.Context, userID string, status entity.UserStatus) error
	ListUsers(ctx context.Context, filter UserFilter) ([]*entity.User, int64, error)
	AddRole(ctx context.Context, userID string, role entity.UserRole) error
	ExistsByID(ctx context.Context, id string) (bool, error)
}

type UserFilter struct {
	Status *entity.UserStatus
	Role   *entity.UserRole
	Search string
	Page   int
	Limit  int
}
