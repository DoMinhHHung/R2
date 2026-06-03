package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/DoMinhHHung/auth-service/internal/application/dto"
	"github.com/DoMinhHHung/auth-service/internal/domain/entity"
	"github.com/DoMinhHHung/auth-service/internal/domain/port"
	"github.com/DoMinhHHung/auth-service/pkg/apperr"
	"github.com/DoMinhHHung/auth-service/pkg/uuidv7"
)

type AdminUseCase struct {
	authUserRepo port.AuthUserRepository
	hasher       port.PasswordHasher
}

func NewAdminUseCase(
	authUserRepo port.AuthUserRepository,
	hasher port.PasswordHasher,
) *AdminUseCase {
	return &AdminUseCase{
		authUserRepo: authUserRepo,
		hasher:       hasher,
	}
}

func (uc *AdminUseCase) CreateAdmin(ctx context.Context, req *dto.CreateAdminRequest, createdByID string) error {
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

	now := time.Now()
	admin := &entity.AuthUser{
		ID:           uuidv7.New(),
		Email:        req.Email,
		PasswordHash: hash,
		Role:         entity.RoleAdmin,
		IsVerified:   true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := uc.authUserRepo.Create(ctx, admin); err != nil {
		return err
	}

	// TODO: Emit audit event "admin.created" với createdByID
	// TODO: Gửi welcome email với temporary password

	return nil
}
