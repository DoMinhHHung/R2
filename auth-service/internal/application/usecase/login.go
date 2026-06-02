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
	"github.com/mssola/useragent"
)

type LoginUseCase struct {
	authUserRepo port.AuthUserRepository
	sessionRepo  port.SessionRepository
	hasher       port.PasswordHasher
	tokenSvc     port.TokenService
	jwtCfg       config.JWTConfig
}

func NewLoginUseCase(
	authUserRepo port.AuthUserRepository,
	sessionRepo port.SessionRepository,
	hasher port.PasswordHasher,
	tokenSvc port.TokenService,
	jwtCfg config.JWTConfig,
) *LoginUseCase {
	return &LoginUseCase{
		authUserRepo: authUserRepo,
		sessionRepo:  sessionRepo,
		hasher:       hasher,
		tokenSvc:     tokenSvc,
		jwtCfg:       jwtCfg,
	}
}

type LoginContext struct {
	IPAddress string
	UserAgent string
}

func (uc *LoginUseCase) Login(ctx context.Context, req *dto.LoginRequest, lctx LoginContext) (*dto.TokenResponse, error) {

	user, err := uc.authUserRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, apperr.ErrUserNotFound) {
			uc.hasher.Verify("dummy", "$argon2id$v=19$m=65536,t=3,p=2$dummysalt$dummyhash")
			return nil, apperr.ErrInvalidCredentials
		}
		return nil, apperr.ErrInternal
	}

	match, err := uc.hasher.Verify(req.Password, user.PasswordHash)
	if err != nil || !match {
		return nil, apperr.ErrInvalidCredentials
	}

	if !user.IsVerified {
		return nil, apperr.ErrUserNotVerified
	}

	ua := useragent.New(lctx.UserAgent)
	browserName, _ := ua.Browser()
	osInfo := ua.OS()
	deviceType := "desktop"
	if ua.Mobile() {
		deviceType = "mobile"
	} else if ua.Bot() {
		deviceType = "bot"
	}

	sessionID := uuidv7.New()
	now := time.Now()
	expiresAt := now.Add(uc.jwtCfg.RefreshTokenExp)

	accessToken, err := uc.tokenSvc.GenerateAccessToken(user, sessionID)
	if err != nil {
		return nil, apperr.ErrInternal
	}

	refreshToken, err := uc.tokenSvc.GenerateRefreshToken(user, sessionID)
	if err != nil {
		return nil, apperr.ErrInternal
	}

	tokenHash := jwtinfra.HashToken(refreshToken)

	session := &entity.Session{
		ID:               sessionID,
		UserID:           user.ID,
		RefreshTokenHash: tokenHash,
		DeviceName:       req.DeviceName,
		DeviceType:       deviceType,
		Browser:          browserName,
		OS:               osInfo,
		IPAddress:        lctx.IPAddress,
		UserAgent:        lctx.UserAgent,
		LastActivity:     now,
		ExpiresAt:        expiresAt,
		CreatedAt:        now,
	}

	if err := uc.sessionRepo.Create(ctx, session); err != nil {
		return nil, apperr.ErrInternal
	}

	return &dto.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(uc.jwtCfg.AccessTokenExp.Seconds()),
	}, nil
}
