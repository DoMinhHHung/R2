package usecase

import (
	"context"
	"errors"

	"github.com/DoMinhHHung/auth-service/internal/application/dto"
	"github.com/DoMinhHHung/auth-service/internal/domain/port"
	"github.com/DoMinhHHung/auth-service/pkg/apperr"
)

type SessionUseCase struct {
	sessionRepo port.SessionRepository
}

func NewSessionUseCase(sessionRepo port.SessionRepository) *SessionUseCase {
	return &SessionUseCase{sessionRepo: sessionRepo}
}

func (uc *SessionUseCase) Logout(ctx context.Context, sessionID string) error {
	if err := uc.sessionRepo.Delete(ctx, sessionID); err != nil {
		if errors.Is(err, apperr.ErrSessionNotFound) {
			return nil
		}
		return apperr.ErrInternal
	}
	return nil
}

func (uc *SessionUseCase) LogoutAll(ctx context.Context, userID string) error {
	if err := uc.sessionRepo.DeleteAllByUserID(ctx, userID); err != nil {
		return apperr.ErrInternal
	}
	return nil
}

func (uc *SessionUseCase) GetActiveSessions(ctx context.Context, userID string) ([]*dto.SessionResponse, error) {
	sessions, err := uc.sessionRepo.FindActiveByUserID(ctx, userID)
	if err != nil {
		return nil, apperr.ErrInternal
	}

	result := make([]*dto.SessionResponse, 0, len(sessions))
	for _, s := range sessions {
		result = append(result, &dto.SessionResponse{
			ID:           s.ID,
			DeviceName:   s.DeviceName,
			DeviceType:   s.DeviceType,
			Browser:      s.Browser,
			OS:           s.OS,
			IPAddress:    s.IPAddress,
			LastActivity: s.LastActivity,
			ExpiresAt:    s.ExpiresAt,
			CreatedAt:    s.CreatedAt,
		})
	}
	return result, nil
}

func (uc *SessionUseCase) RevokeSession(ctx context.Context, userID, sessionID string) error {
	session, err := uc.sessionRepo.FindByID(ctx, sessionID)
	if err != nil {
		return err
	}

	if session.UserID != userID {
		return apperr.ErrSessionNotFound
	}

	return uc.sessionRepo.Delete(ctx, sessionID)
}
