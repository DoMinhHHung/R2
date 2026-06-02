package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/DoMinhHHung/auth-service/internal/application/dto"
	"github.com/DoMinhHHung/auth-service/internal/domain/entity"
	"github.com/DoMinhHHung/auth-service/internal/domain/port"
	"github.com/DoMinhHHung/auth-service/internal/infrastructure/config"
	"github.com/DoMinhHHung/auth-service/pkg/apperr"
	"github.com/DoMinhHHung/auth-service/pkg/uuidv7"
)

type SignupUseCase struct {
	authUserRepo port.AuthUserRepository
	cacheRepo    port.CacheRepository
	hasher       port.PasswordHasher
	otpGen       port.OTPGenerator
	emailSvc     port.EmailService
	userClient   port.UserServiceClient
	otpCfg       config.OTPConfig
}

func NewSignupUseCase(
	authUserRepo port.AuthUserRepository,
	cacheRepo port.CacheRepository,
	hasher port.PasswordHasher,
	otpGen port.OTPGenerator,
	emailSvc port.EmailService,
	userClient port.UserServiceClient,
	otpCfg config.OTPConfig,
) *SignupUseCase {
	return &SignupUseCase{
		authUserRepo: authUserRepo,
		cacheRepo:    cacheRepo,
		hasher:       hasher,
		otpGen:       otpGen,
		emailSvc:     emailSvc,
		userClient:   userClient,
		otpCfg:       otpCfg,
	}
}

func (uc *SignupUseCase) InitiateSignup(ctx context.Context, req *dto.SignupRequest) error {
	existing, err := uc.authUserRepo.FindByEmail(ctx, req.Email)
	if err != nil && !errors.Is(err, apperr.ErrUserNotFound) {
		return apperr.ErrInternal
	}
	if existing != nil {
		return apperr.ErrEmailExists
	}

	hash, err := uc.hasher.Hash(req.Password)
	if err != nil {
		return apperr.ErrInternal
	}

	otp, err := uc.otpGen.Generate()
	if err != nil {
		return apperr.ErrInternal
	}

	session := &entity.SignupSession{
		Email:        req.Email,
		PasswordHash: hash,
		Role:         entity.Role(req.Role),
		OTP:          otp,
		RetryCount:   0,
	}
	if err := uc.cacheRepo.SetSignupSession(ctx, session); err != nil {
		return apperr.ErrInternal
	}

	go func() {
		bgCtx := context.Background()
		if err := uc.emailSvc.SendSignupOTP(bgCtx, req.Email, otp); err != nil {
			// TODO: log error properly via injected logger
		}
	}()

	return nil
}

func (uc *SignupUseCase) VerifyOTP(ctx context.Context, req *dto.VerifyOTPRequest) error {
	session, err := uc.cacheRepo.GetSignupSession(ctx, req.Email)
	if err != nil {
		return err
	}

	if session.RetryCount >= uc.otpCfg.MaxRetry {
		return apperr.ErrMaxRetryExceeded
	}

	if session.OTP != req.OTP {
		session.RetryCount++
		_ = uc.cacheRepo.UpdateSignupSession(ctx, session)
		return apperr.ErrInvalidOTP
	}

	now := time.Now()
	user := &entity.AuthUser{
		ID:           uuidv7.New(),
		Email:        session.Email,
		PasswordHash: session.PasswordHash,
		Role:         session.Role,
		IsVerified:   true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := uc.authUserRepo.Create(ctx, user); err != nil {
		return err
	}

	go func() {
		bgCtx := context.Background()
		_ = uc.userClient.CreateProfile(bgCtx, user.ID, user.Email, string(user.Role))
	}()

	_ = uc.cacheRepo.DeleteSignupSession(ctx, req.Email)

	return nil
}

func (uc *SignupUseCase) ResendOTP(ctx context.Context, req *dto.ResendOTPRequest) error {
	session, err := uc.cacheRepo.GetSignupSession(ctx, req.Email)
	if err != nil {
		return err
	}

	otp, err := uc.otpGen.Generate()
	if err != nil {
		return apperr.ErrInternal
	}

	session.OTP = otp
	session.RetryCount = 0
	if err := uc.cacheRepo.SetSignupSession(ctx, session); err != nil {
		return apperr.ErrInternal
	}

	go func() {
		_ = uc.emailSvc.SendSignupOTP(context.Background(), req.Email, otp)
	}()

	return nil
}
