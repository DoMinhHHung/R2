package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/DoMinhHHung/auth-service/internal/application/dto"
	"github.com/DoMinhHHung/auth-service/internal/domain/entity"
	"github.com/DoMinhHHung/auth-service/internal/domain/port"
	"github.com/DoMinhHHung/auth-service/internal/infrastructure/config"
	jwtinfra "github.com/DoMinhHHung/auth-service/internal/infrastructure/jwt"
	"github.com/DoMinhHHung/auth-service/pkg/apperr"
	"github.com/DoMinhHHung/auth-service/pkg/uuidv7"
)

type TokenUseCase struct {
	authUserRepo port.AuthUserRepository
	sessionRepo  port.SessionRepository
	tokenSvc     port.TokenService
	jwtCfg       config.JWTConfig
}

func NewTokenUseCase(
	authUserRepo port.AuthUserRepository,
	sessionRepo port.SessionRepository,
	tokenSvc port.TokenService,
	jwtCfg config.JWTConfig,
) *TokenUseCase {
	return &TokenUseCase{
		authUserRepo: authUserRepo,
		sessionRepo:  sessionRepo,
		tokenSvc:     tokenSvc,
		jwtCfg:       jwtCfg,
	}
}

func (uc *TokenUseCase) RefreshToken(ctx context.Context, req *dto.RefreshTokenRequest) (*dto.TokenResponse, error) {
	claims, err := uc.tokenSvc.ParseToken(req.RefreshToken)
	if err != nil {
		return nil, apperr.ErrTokenInvalid
	}

	if claims.TokenType != "refresh" {
		return nil, apperr.ErrTokenInvalid
	}

	session, err := uc.sessionRepo.FindByID(ctx, claims.SessionID)
	if err != nil {
		if errors.Is(err, apperr.ErrSessionNotFound) {
			return nil, apperr.ErrSessionNotFound
		}
		return nil, apperr.ErrInternal
	}

	if session.IsExpired() {
		_ = uc.sessionRepo.Delete(ctx, session.ID)
		return nil, apperr.ErrTokenExpired
	}

	incomingHash := jwtinfra.HashToken(req.RefreshToken)
	if incomingHash != session.RefreshTokenHash {
		_ = uc.sessionRepo.Delete(ctx, session.ID)
		return nil, apperr.ErrTokenInvalid
	}

	user, err := uc.authUserRepo.FindByID(ctx, session.UserID)
	if err != nil {
		return nil, apperr.ErrInternal
	}

	newSessionID := uuidv7.New()
	now := time.Now()

	newAccessToken, err := uc.tokenSvc.GenerateAccessToken(user, newSessionID)
	if err != nil {
		return nil, apperr.ErrInternal
	}

	newRefreshToken, err := uc.tokenSvc.GenerateRefreshToken(user, newSessionID)
	if err != nil {
		return nil, apperr.ErrInternal
	}

	newSession := &entity.Session{
		ID:               newSessionID,
		UserID:           user.ID,
		RefreshTokenHash: jwtinfra.HashToken(newRefreshToken),
		DeviceName:       session.DeviceName,
		DeviceType:       session.DeviceType,
		Browser:          session.Browser,
		OS:               session.OS,
		IPAddress:        session.IPAddress,
		UserAgent:        session.UserAgent,
		LastActivity:     now,
		ExpiresAt:        now.Add(uc.jwtCfg.RefreshTokenExp),
		CreatedAt:        now,
	}

	if err := uc.sessionRepo.Delete(ctx, session.ID); err != nil {
		return nil, apperr.ErrInternal
	}
	if err := uc.sessionRepo.Create(ctx, newSession); err != nil {
		return nil, apperr.ErrInternal
	}

	return &dto.TokenResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(uc.jwtCfg.AccessTokenExp.Seconds()),
	}, nil
}
