package port

import (
	"context"
	"github.com/DoMinhHHung/auth-service/internal/domain/entity"
)

type AuthUserRepository interface {
	Create(ctx context.Context, user *entity.AuthUser) error
	FindByEmail(ctx context.Context, email string) (*entity.AuthUser, error)
	FindByID(ctx context.Context, id string) (*entity.AuthUser, error)
	UpdatePassword(ctx context.Context, id string, passwordHash string) error
}

type SessionRepository interface {
	Create(ctx context.Context, session *entity.Session) error
	FindByID(ctx context.Context, id string) (*entity.Session, error)
	FindActiveByUserID(ctx context.Context, userID string) ([]*entity.Session, error)
	UpdateLastActivity(ctx context.Context, id string) error
	Delete(ctx context.Context, id string) error
	DeleteAllByUserID(ctx context.Context, userID string) error
}

type CacheRepository interface {
	// Signup flow
	SetSignupSession(ctx context.Context, session *entity.SignupSession) error
	GetSignupSession(ctx context.Context, email string) (*entity.SignupSession, error)
	UpdateSignupSession(ctx context.Context, session *entity.SignupSession) error
	DeleteSignupSession(ctx context.Context, email string) error

	// Recovery flow
	SetRecoverySession(ctx context.Context, session *entity.RecoverySession) error
	GetRecoverySession(ctx context.Context, email string) (*entity.RecoverySession, error)
	UpdateRecoverySession(ctx context.Context, session *entity.RecoverySession) error
	DeleteRecoverySession(ctx context.Context, email string) error

	// Reset token (one-time use, 10 min TTL)
	SetResetToken(ctx context.Context, token string, email string) error
	GetResetToken(ctx context.Context, token string) (string, error) // returns email
	DeleteResetToken(ctx context.Context, token string) error
}

type TransactionManager interface {
	WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}
