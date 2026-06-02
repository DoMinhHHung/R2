package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/DoMinhHHung/auth-service/internal/domain/entity"
	"github.com/DoMinhHHung/auth-service/internal/domain/port"
	"github.com/DoMinhHHung/auth-service/internal/infrastructure/database"
	"github.com/DoMinhHHung/auth-service/pkg/apperr"
	"github.com/jackc/pgx/v5"
)

type sessionRepo struct {
	db *database.Pool
}

func NewSessionRepository(db *database.Pool) port.SessionRepository {
	return &sessionRepo{db: db}
}

func (r *sessionRepo) Create(ctx context.Context, s *entity.Session) error {
	q := r.db.GetConn(ctx)
	_, err := q.Exec(ctx, `
        INSERT INTO user_sessions (
            id, user_id, refresh_token_hash, device_name, device_type,
            browser, os, ip_address, user_agent, last_activity, expires_at, created_at
        ) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
    `,
		s.ID, s.UserID, s.RefreshTokenHash, s.DeviceName, s.DeviceType,
		s.Browser, s.OS, s.IPAddress, s.UserAgent,
		s.LastActivity, s.ExpiresAt, s.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("create session: %w", err)
	}
	return nil
}

func (r *sessionRepo) FindByID(ctx context.Context, id string) (*entity.Session, error) {
	q := r.db.GetConn(ctx)
	var s entity.Session

	err := q.QueryRow(ctx, `
        SELECT id, user_id, refresh_token_hash, device_name, device_type,
               browser, os, ip_address::text, user_agent, last_activity, expires_at, created_at
        FROM user_sessions
        WHERE id = $1
    `, id).Scan(
		&s.ID, &s.UserID, &s.RefreshTokenHash, &s.DeviceName, &s.DeviceType,
		&s.Browser, &s.OS, &s.IPAddress, &s.UserAgent,
		&s.LastActivity, &s.ExpiresAt, &s.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperr.ErrSessionNotFound
		}
		return nil, fmt.Errorf("find session by id: %w", err)
	}
	return &s, nil
}

func (r *sessionRepo) FindActiveByUserID(ctx context.Context, userID string) ([]*entity.Session, error) {
	q := r.db.GetConn(ctx)
	rows, err := q.Query(ctx, `
        SELECT id, user_id, device_name, device_type, browser, os,
               ip_address::text, last_activity, expires_at, created_at
        FROM user_sessions
        WHERE user_id = $1 AND expires_at > NOW()
        ORDER BY last_activity DESC
    `, userID)
	if err != nil {
		return nil, fmt.Errorf("find active sessions: %w", err)
	}
	defer rows.Close()

	var sessions []*entity.Session
	for rows.Next() {
		var s entity.Session
		if err := rows.Scan(
			&s.ID, &s.UserID, &s.DeviceName, &s.DeviceType,
			&s.Browser, &s.OS, &s.IPAddress,
			&s.LastActivity, &s.ExpiresAt, &s.CreatedAt,
		); err != nil {
			return nil, err
		}
		sessions = append(sessions, &s)
	}
	return sessions, rows.Err()
}

func (r *sessionRepo) UpdateLastActivity(ctx context.Context, id string) error {
	q := r.db.GetConn(ctx)
	_, err := q.Exec(ctx, `
        UPDATE user_sessions SET last_activity = NOW() WHERE id = $1
    `, id)
	return err
}

func (r *sessionRepo) Delete(ctx context.Context, id string) error {
	q := r.db.GetConn(ctx)
	tag, err := q.Exec(ctx, `DELETE FROM user_sessions WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return apperr.ErrSessionNotFound
	}
	return nil
}

func (r *sessionRepo) DeleteAllByUserID(ctx context.Context, userID string) error {
	q := r.db.GetConn(ctx)
	_, err := q.Exec(ctx, `DELETE FROM user_sessions WHERE user_id = $1`, userID)
	return err
}
