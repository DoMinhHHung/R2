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

type authUserRepo struct {
	db *database.Pool
}

func NewAuthUserRepository(db *database.Pool) port.AuthUserRepository {
	return &authUserRepo{db: db}
}

func (r *authUserRepo) Create(ctx context.Context, user *entity.AuthUser) error {
	q := r.db.GetConn(ctx)
	_, err := q.Exec(ctx, `
        INSERT INTO auth_users (id, email, password_hash, role, is_verified, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
    `,
		user.ID, user.Email, user.PasswordHash,
		string(user.Role), user.IsVerified,
		user.CreatedAt, user.UpdatedAt,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return apperr.ErrEmailExists
		}
		return fmt.Errorf("create auth user: %w", err)
	}
	return nil
}

func (r *authUserRepo) FindByEmail(ctx context.Context, email string) (*entity.AuthUser, error) {
	q := r.db.GetConn(ctx)
	var user entity.AuthUser
	var role string

	err := q.QueryRow(ctx, `
        SELECT id, email, password_hash, role, is_verified, created_at, updated_at
        FROM auth_users
        WHERE email = $1
    `, email).Scan(
		&user.ID, &user.Email, &user.PasswordHash,
		&role, &user.IsVerified,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperr.ErrUserNotFound
		}
		return nil, fmt.Errorf("find user by email: %w", err)
	}

	user.Role = entity.Role(role)
	return &user, nil
}

func (r *authUserRepo) FindByID(ctx context.Context, id string) (*entity.AuthUser, error) {
	q := r.db.GetConn(ctx)
	var user entity.AuthUser
	var role string

	err := q.QueryRow(ctx, `
        SELECT id, email, password_hash, role, is_verified, created_at, updated_at
        FROM auth_users
        WHERE id = $1
    `, id).Scan(
		&user.ID, &user.Email, &user.PasswordHash,
		&role, &user.IsVerified,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperr.ErrUserNotFound
		}
		return nil, fmt.Errorf("find user by id: %w", err)
	}

	user.Role = entity.Role(role)
	return &user, nil
}

func (r *authUserRepo) UpdatePassword(ctx context.Context, id, passwordHash string) error {
	q := r.db.GetConn(ctx)
	tag, err := q.Exec(ctx, `
        UPDATE auth_users SET password_hash = $1, updated_at = NOW()
        WHERE id = $2
    `, passwordHash, id)
	if err != nil {
		return fmt.Errorf("update password: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return apperr.ErrUserNotFound
	}
	return nil
}

func isUniqueViolation(err error) bool {
	return err != nil && (containsStr(err.Error(), "23505") ||
		containsStr(err.Error(), "unique"))
}

func containsStr(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 &&
		func() bool {
			for i := 0; i <= len(s)-len(sub); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
			return false
		}())
}
