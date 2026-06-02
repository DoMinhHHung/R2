package usecase

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"

	"github.com/DoMinhHHung/auth-service/internal/application/dto"
	"github.com/DoMinhHHung/auth-service/internal/domain/entity"
	"github.com/DoMinhHHung/auth-service/internal/domain/port"
	"github.com/DoMinhHHung/auth-service/pkg/apperr"
)

type PasswordUseCase struct {
	authUserRepo port.AuthUserRepository
	sessionRepo  port.SessionRepository
	cacheRepo    port.CacheRepository
	hasher       port.PasswordHasher
	otpGen       port.OTPGenerator
	emailSvc     port.EmailService
}

func NewPasswordUseCase(
	authUserRepo port.AuthUserRepository,
	sessionRepo port.SessionRepository,
	cacheRepo port.CacheRepository,
	hasher port.PasswordHasher,
	otpGen port.OTPGenerator,
	emailSvc port.EmailService,
) *PasswordUseCase {
	return &PasswordUseCase{
		authUserRepo: authUserRepo,
		sessionRepo:  sessionRepo,
		cacheRepo:    cacheRepo,
		hasher:       hasher,
		otpGen:       otpGen,
		emailSvc:     emailSvc,
	}
}

func (uc *PasswordUseCase) ForgotPassword(ctx context.Context, req *dto.ForgotPasswordRequest) error {
	user, err := uc.authUserRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, apperr.ErrUserNotFound) {
			return nil
		}
		return apperr.ErrInternal
	}

	otp, err := uc.otpGen.Generate()
	if err != nil {
		return apperr.ErrInternal
	}

	session := &entity.RecoverySession{
		Email:      user.Email,
		OTP:        otp,
		RetryCount: 0,
	}
	if err := uc.cacheRepo.SetRecoverySession(ctx, session); err != nil {
		return apperr.ErrInternal
	}

	go func() {
		_ = uc.emailSvc.SendRecoveryOTP(context.Background(), req.Email, otp)
	}()

	return nil
}

func (uc *PasswordUseCase) VerifyRecoveryOTP(ctx context.Context, req *dto.VerifyRecoveryOTPRequest) (string, error) {
	session, err := uc.cacheRepo.GetRecoverySession(ctx, req.Email)
	if err != nil {
		return "", err
	}

	if session.RetryCount >= 5 {
		return "", apperr.ErrMaxRetryExceeded
	}

	if session.OTP != req.OTP {
		session.RetryCount++
		_ = uc.cacheRepo.UpdateRecoverySession(ctx, session)
		return "", apperr.ErrInvalidOTP
	}

	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", apperr.ErrInternal
	}
	resetToken := hex.EncodeToString(b)

	if err := uc.cacheRepo.SetResetToken(ctx, resetToken, req.Email); err != nil {
		return "", apperr.ErrInternal
	}

	_ = uc.cacheRepo.DeleteRecoverySession(ctx, req.Email)

	return resetToken, nil
}

func (uc *PasswordUseCase) ResetPassword(ctx context.Context, req *dto.ResetPasswordRequest) error {
	email, err := uc.cacheRepo.GetResetToken(ctx, req.ResetToken)
	if err != nil {
		return err
	}

	user, err := uc.authUserRepo.FindByEmail(ctx, email)
	if err != nil {
		return apperr.ErrInternal
	}

	hash, err := uc.hasher.Hash(req.NewPassword)
	if err != nil {
		return apperr.ErrInternal
	}

	if err := uc.authUserRepo.UpdatePassword(ctx, user.ID, hash); err != nil {
		return apperr.ErrInternal
	}

	_ = uc.sessionRepo.DeleteAllByUserID(ctx, user.ID)

	_ = uc.cacheRepo.DeleteResetToken(ctx, req.ResetToken)

	return nil
}
