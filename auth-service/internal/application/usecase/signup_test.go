package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/DoMinhHHung/auth-service/internal/application/dto"
	"github.com/DoMinhHHung/auth-service/internal/application/usecase"
	"github.com/DoMinhHHung/auth-service/internal/domain/entity"
	"github.com/DoMinhHHung/auth-service/internal/infrastructure/config"
	"github.com/DoMinhHHung/auth-service/pkg/apperr"
)

// Mock implementations
type mockAuthUserRepo struct {
	users map[string]*entity.AuthUser
}

func newMockAuthUserRepo() *mockAuthUserRepo {
	return &mockAuthUserRepo{users: make(map[string]*entity.AuthUser)}
}

func (m *mockAuthUserRepo) Create(_ context.Context, user *entity.AuthUser) error {
	if _, exists := m.users[user.Email]; exists {
		return apperr.ErrEmailExists
	}
	m.users[user.Email] = user
	return nil
}

func (m *mockAuthUserRepo) FindByEmail(_ context.Context, email string) (*entity.AuthUser, error) {
	if u, ok := m.users[email]; ok {
		return u, nil
	}
	return nil, apperr.ErrUserNotFound
}

func (m *mockAuthUserRepo) FindByID(_ context.Context, id string) (*entity.AuthUser, error) {
	return nil, apperr.ErrUserNotFound
}

func (m *mockAuthUserRepo) UpdatePassword(_ context.Context, id, hash string) error {
	return nil
}

type mockCacheRepo struct {
	sessions map[string]*entity.SignupSession
}

func newMockCacheRepo() *mockCacheRepo {
	return &mockCacheRepo{sessions: make(map[string]*entity.SignupSession)}
}

func (m *mockCacheRepo) SetSignupSession(_ context.Context, s *entity.SignupSession) error {
	m.sessions[s.Email] = s
	return nil
}

func (m *mockCacheRepo) GetSignupSession(_ context.Context, email string) (*entity.SignupSession, error) {
	if s, ok := m.sessions[email]; ok {
		return s, nil
	}
	return nil, apperr.ErrOTPExpired
}

func (m *mockCacheRepo) UpdateSignupSession(_ context.Context, s *entity.SignupSession) error {
	m.sessions[s.Email] = s
	return nil
}

func (m *mockCacheRepo) DeleteSignupSession(_ context.Context, email string) error {
	delete(m.sessions, email)
	return nil
}

func (m *mockCacheRepo) DeleteRecoverySession(_ context.Context, email string) error {
	return nil
}

func (m *mockCacheRepo) SetRecoverySession(_ context.Context, s *entity.RecoverySession) error {
	return nil
}

func (m *mockCacheRepo) GetRecoverySession(_ context.Context, email string) (*entity.RecoverySession, error) {
	return nil, apperr.ErrOTPExpired
}

func (m *mockCacheRepo) UpdateRecoverySession(_ context.Context, s *entity.RecoverySession) error {
	return nil
}

func (m *mockCacheRepo) SetResetToken(_ context.Context, token string, email string) error {
	return nil
}

func (m *mockCacheRepo) GetResetToken(_ context.Context, token string) (string, error) {
	return "", apperr.ErrOTPExpired
}

func (m *mockCacheRepo) DeleteResetToken(_ context.Context, token string) error {
	return nil
}

type mockHasher struct{}

func (m *mockHasher) Hash(p string) (string, error)    { return "hashed:" + p, nil }
func (m *mockHasher) Verify(p, h string) (bool, error) { return h == "hashed:"+p, nil }

type mockOTPGen struct{ otp string }

func (m *mockOTPGen) Generate() (string, error) { return m.otp, nil }

type mockEmailSvc struct{ sent []string }

func (m *mockEmailSvc) SendSignupOTP(_ context.Context, to, _ string) error {
	m.sent = append(m.sent, to)
	return nil
}

func (m *mockEmailSvc) SendRecoveryOTP(_ context.Context, to, _ string) error { return nil }

type mockUserClient struct{}

func (m *mockUserClient) CreateProfile(_ context.Context, _, _, _ string) error { return nil }

func TestSignupUseCase_InitiateSignup(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		uc := usecase.NewSignupUseCase(
			newMockAuthUserRepo(),
			newMockCacheRepo(),
			&mockHasher{},
			&mockOTPGen{otp: "123456"},
			&mockEmailSvc{},
			&mockUserClient{},
			config.OTPConfig{MaxRetry: 5, TTLMinutes: 5},
		)

		err := uc.InitiateSignup(context.Background(), &dto.SignupRequest{
			Email:    "test@example.com",
			Password: "password123",
			Role:     "TENANT",
		})

		if err != nil {
			t.Errorf("expected nil error, got: %v", err)
		}
	})

	t.Run("email_exists", func(t *testing.T) {
		repo := newMockAuthUserRepo()
		repo.users["existing@example.com"] = &entity.AuthUser{Email: "existing@example.com"}

		uc := usecase.NewSignupUseCase(
			repo, newMockCacheRepo(),
			&mockHasher{}, &mockOTPGen{otp: "123456"},
			&mockEmailSvc{}, &mockUserClient{},
			config.OTPConfig{MaxRetry: 5},
		)

		err := uc.InitiateSignup(context.Background(), &dto.SignupRequest{
			Email: "existing@example.com", Password: "password123", Role: "TENANT",
		})

		if !errors.Is(err, apperr.ErrEmailExists) {
			t.Errorf("expected ErrEmailExists, got: %v", err)
		}
	})
}

func TestSignupUseCase_VerifyOTP(t *testing.T) {
	t.Run("invalid_otp_increments_retry", func(t *testing.T) {
		cache := newMockCacheRepo()
		session := &entity.SignupSession{
			Email: "test@example.com", OTP: "123456", RetryCount: 0,
			Role: entity.RoleTenant, PasswordHash: "hash",
		}
		cache.sessions["test@example.com"] = session

		uc := usecase.NewSignupUseCase(
			newMockAuthUserRepo(), cache,
			&mockHasher{}, &mockOTPGen{},
			&mockEmailSvc{}, &mockUserClient{},
			config.OTPConfig{MaxRetry: 5},
		)

		err := uc.VerifyOTP(context.Background(), &dto.VerifyOTPRequest{
			Email: "test@example.com", OTP: "000000",
		})

		if !errors.Is(err, apperr.ErrInvalidOTP) {
			t.Errorf("expected ErrInvalidOTP, got: %v", err)
		}

		if cache.sessions["test@example.com"].RetryCount != 1 {
			t.Error("retry count should be incremented")
		}
	})
}
