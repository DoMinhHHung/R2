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

type AdminLoginUseCase struct {
	authUserRepo port.AuthUserRepository
	sessionRepo  port.SessionRepository
	hasher       port.PasswordHasher
	tokenSvc     port.TokenService
	jwtCfg       config.JWTConfig
}

func NewAdminLoginUseCase(
	authUserRepo port.AuthUserRepository,
	sessionRepo port.SessionRepository,
	hasher port.PasswordHasher,
	tokenSvc port.TokenService,
	jwtCfg config.JWTConfig,
) *AdminLoginUseCase {
	return &AdminLoginUseCase{
		authUserRepo: authUserRepo,
		sessionRepo:  sessionRepo,
		hasher:       hasher,
		tokenSvc:     tokenSvc,
		jwtCfg:       jwtCfg,
	}
}

func (uc *AdminLoginUseCase) Login(
	ctx context.Context,
	req *dto.AdminLoginRequest,
	lctx LoginContext,
) (*dto.TokenResponse, error) {

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

	if user.Role != entity.RoleAdmin {
		return nil, apperr.ErrInvalidCredentials
	}

	if !user.IsVerified {
		return nil, apperr.ErrInvalidCredentials
	}

	ua := useragent.New(lctx.UserAgent)
	browserName, _ := ua.Browser()
	deviceType := "desktop"
	if ua.Mobile() {
		deviceType = "mobile"
	}

	sessionID := uuidv7.New()
	now := time.Now()

	accessToken, err := uc.tokenSvc.GenerateAccessToken(user, sessionID)
	if err != nil {
		return nil, apperr.ErrInternal
	}

	refreshToken, err := uc.tokenSvc.GenerateRefreshToken(user, sessionID)
	if err != nil {
		return nil, apperr.ErrInternal
	}

	session := &entity.Session{
		ID:               sessionID,
		UserID:           user.ID,
		RefreshTokenHash: jwtinfra.HashToken(refreshToken),
		DeviceName:       req.DeviceName,
		DeviceType:       deviceType,
		Browser:          browserName,
		OS:               ua.OS(),
		IPAddress:        lctx.IPAddress,
		UserAgent:        lctx.UserAgent,
		LastActivity:     now,
		ExpiresAt:        now.Add(uc.jwtCfg.RefreshTokenExp),
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
