package port

import (
	"context"

	"github.com/DoMinhHHung/auth-service/internal/domain/entity"
)

type PasswordHasher interface {
	Hash(password string) (string, error)
	Verify(password, hash string) (bool, error)
}

type OTPGenerator interface {
	Generate() (string, error)
}

type TokenClaims struct {
	UserID    string
	Email     string
	Role      string
	SessionID string
	TokenType string
}

type TokenService interface {
	GenerateAccessToken(user *entity.AuthUser, sessionID string) (string, error)
	GenerateRefreshToken(user *entity.AuthUser, sessionID string) (string, error)
	ParseToken(tokenStr string) (*TokenClaims, error)
}

type EmailService interface {
	SendSignupOTP(ctx context.Context, to, otp string) error
	SendRecoveryOTP(ctx context.Context, to, otp string) error
}

type UserServiceClient interface {
	CreateProfile(ctx context.Context, userID, email, role string) error
}
