package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/DoMinhHHung/user-service/internal/domain/entity"
	"github.com/DoMinhHHung/user-service/internal/domain/port"
	"github.com/DoMinhHHung/user-service/internal/infrastructure/postgres"
	"github.com/DoMinhHHung/user-service/pkg/apperr"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

type userRepo struct{ db *postgres.Pool }

func New(db *postgres.Pool) port.UserRepository {
	return &userRepo{db: db}
}

func (r *userRepo) Create(ctx context.Context, u *entity.User) error {
	_, err := r.db.Q(ctx).Exec(ctx, `
		INSERT INTO users (id, email, profile_completed, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (id) DO NOTHING
	`, u.ID, u.Email, false, string(entity.StatusActive), u.CreatedAt, u.UpdatedAt)
	if err != nil {
		if isUniqueViolation(err) {
			return apperr.ErrUserAlreadyExists
		}
		return fmt.Errorf("create user: %w", err)
	}

	for _, role := range u.Roles {
		if err := r.AddRole(ctx, u.ID, role); err != nil {
			return fmt.Errorf("add role %s: %w", role, err)
		}
	}
	return nil
}

func (r *userRepo) FindByID(ctx context.Context, id string) (*entity.User, error) {
	var u entity.User
	var gender pgtype.Text
	var status pgtype.Text
	var dob *time.Time

	err := r.db.Q(ctx).QueryRow(ctx, `
		SELECT id, email, COALESCE(full_name,''), COALESCE(phone_number,''),
		       gender, date_of_birth, COALESCE(avatar_url,''),
		       COALESCE(avatar_public_id,''), profile_completed, status,
		       created_at, updated_at
		FROM users WHERE id = $1 AND status != 'DELETED'
	`, id).Scan(
		&u.ID, &u.Email, &u.FullName, &u.PhoneNumber,
		&gender, &dob,
		&u.AvatarURL, &u.AvatarPublicID,
		&u.ProfileCompleted, &status,
		&u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperr.ErrUserNotFound
		}
		return nil, fmt.Errorf("find user by id: %w", err)
	}

	if gender.Valid {
		u.Gender = entity.UserGender(gender.String)
	}
	if status.Valid {
		u.Status = entity.UserStatus(status.String)
	}
	u.DateOfBirth = dob

	roles, err := r.loadRoles(ctx, u.ID)
	if err != nil {
		return nil, err
	}
	u.Roles = roles
	return &u, nil
}

func (r *userRepo) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	var id string
	err := r.db.Q(ctx).QueryRow(ctx,
		`SELECT id FROM users WHERE email = $1 AND status != 'DELETED'`, email,
	).Scan(&id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperr.ErrUserNotFound
		}
		return nil, fmt.Errorf("find user by email: %w", err)
	}
	return r.FindByID(ctx, id)
}

func (r *userRepo) Update(ctx context.Context, u *entity.User) error {
	tag, err := r.db.Q(ctx).Exec(ctx, `
		UPDATE users SET
			full_name        = $1,
			phone_number     = $2,
			gender           = $3,
			date_of_birth    = $4,
			profile_completed = $5,
			updated_at       = NOW()
		WHERE id = $6 AND status = 'ACTIVE'
	`,
		nullStr(u.FullName), nullStr(u.PhoneNumber),
		nullStr(string(u.Gender)), u.DateOfBirth,
		u.ProfileCompleted, u.ID,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return apperr.ErrPhoneExists
		}
		return fmt.Errorf("update user: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return apperr.ErrUserNotFound
	}
	return nil
}

func (r *userRepo) UpdateAvatar(ctx context.Context, userID, avatarURL, publicID string) error {
	tag, err := r.db.Q(ctx).Exec(ctx, `
		UPDATE users SET avatar_url = $1, avatar_public_id = $2, updated_at = NOW()
		WHERE id = $3 AND status = 'ACTIVE'
	`, avatarURL, publicID, userID)
	if err != nil {
		return fmt.Errorf("update avatar: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return apperr.ErrUserNotFound
	}
	return nil
}

func (r *userRepo) DeleteAvatar(ctx context.Context, userID string) error {
	tag, err := r.db.Q(ctx).Exec(ctx, `
		UPDATE users SET avatar_url = NULL, avatar_public_id = NULL, updated_at = NOW()
		WHERE id = $1 AND status = 'ACTIVE'
	`, userID)
	if err != nil {
		return fmt.Errorf("delete avatar: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return apperr.ErrUserNotFound
	}
	return nil
}

func (r *userRepo) UpdateStatus(ctx context.Context, userID string, status entity.UserStatus) error {
	tag, err := r.db.Q(ctx).Exec(ctx,
		`UPDATE users SET status = $1, updated_at = NOW() WHERE id = $2`,
		string(status), userID,
	)
	if err != nil {
		return fmt.Errorf("update status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return apperr.ErrUserNotFound
	}
	return nil
}

func (r *userRepo) ListUsers(ctx context.Context, f port.UserFilter) ([]*entity.User, int64, error) {
	args := []any{}
	where := []string{"u.status != 'DELETED'"}
	i := 1

	if f.Status != nil {
		where = append(where, fmt.Sprintf("u.status = $%d", i))
		args = append(args, string(*f.Status))
		i++
	}
	if f.Search != "" {
		where = append(where, fmt.Sprintf("(u.email ILIKE $%d OR u.full_name ILIKE $%d)", i, i))
		args = append(args, "%"+f.Search+"%")
		i++
	}
	if f.Role != nil {
		where = append(where, fmt.Sprintf(
			"EXISTS (SELECT 1 FROM user_roles ur WHERE ur.user_id = u.id AND ur.role = $%d)", i,
		))
		args = append(args, string(*f.Role))
		i++
	}

	whereClause := "WHERE " + strings.Join(where, " AND ")

	var total int64
	if err := r.db.Q(ctx).QueryRow(ctx,
		`SELECT COUNT(*) FROM users u `+whereClause, args...,
	).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count users: %w", err)
	}

	if total == 0 {
		return []*entity.User{}, 0, nil
	}

	offset := (f.Page - 1) * f.Limit
	queryArgs := append(args, f.Limit, offset)

	rows, err := r.db.Q(ctx).Query(ctx, `
		SELECT u.id, u.email, COALESCE(u.full_name,''), COALESCE(u.phone_number,''),
		       u.gender, u.date_of_birth, COALESCE(u.avatar_url,''),
		       u.profile_completed, u.status, u.created_at, u.updated_at
		FROM users u
		`+whereClause+`
		ORDER BY u.created_at DESC
		LIMIT $`+fmt.Sprint(i)+` OFFSET $`+fmt.Sprint(i+1),
		queryArgs...,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	var users []*entity.User
	for rows.Next() {
		var u entity.User
		var gender, status pgtype.Text
		var dob *time.Time
		if err := rows.Scan(
			&u.ID, &u.Email, &u.FullName, &u.PhoneNumber,
			&gender, &dob, &u.AvatarURL,
			&u.ProfileCompleted, &status, &u.CreatedAt, &u.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan user row: %w", err)
		}
		if gender.Valid {
			u.Gender = entity.UserGender(gender.String)
		}
		if status.Valid {
			u.Status = entity.UserStatus(status.String)
		}
		u.DateOfBirth = dob
		users = append(users, &u)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows error: %w", err)
	}

	if len(users) == 0 {
		return users, total, nil
	}

	userIDs := make([]string, len(users))
	for i, u := range users {
		userIDs[i] = u.ID
	}
	rolesMap, err := r.loadRolesBatch(ctx, userIDs)
	if err != nil {
		return nil, 0, err
	}
	for _, u := range users {
		u.Roles = rolesMap[u.ID]
		if u.Roles == nil {
			u.Roles = []entity.UserRole{}
		}
	}

	return users, total, nil
}

func (r *userRepo) AddRole(ctx context.Context, userID string, role entity.UserRole) error {
	_, err := r.db.Q(ctx).Exec(ctx, `
		INSERT INTO user_roles (user_id, role)
		VALUES ($1, $2)
		ON CONFLICT (user_id, role) DO NOTHING
	`, userID, string(role))
	return err
}

func (r *userRepo) ExistsByID(ctx context.Context, id string) (bool, error) {
	var exists bool
	err := r.db.Q(ctx).QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)`, id,
	).Scan(&exists)
	return exists, err
}

func (r *userRepo) loadRoles(ctx context.Context, userID string) ([]entity.UserRole, error) {
	rows, err := r.db.Q(ctx).Query(ctx,
		`SELECT role FROM user_roles WHERE user_id = $1 ORDER BY granted_at`, userID,
	)
	if err != nil {
		return nil, fmt.Errorf("load roles: %w", err)
	}
	defer rows.Close()

	var roles []entity.UserRole
	for rows.Next() {
		var role string
		if err := rows.Scan(&role); err != nil {
			return nil, err
		}
		roles = append(roles, entity.UserRole(role))
	}
	return roles, rows.Err()
}

func (r *userRepo) loadRolesBatch(ctx context.Context, userIDs []string) (map[string][]entity.UserRole, error) {
	if len(userIDs) == 0 {
		return map[string][]entity.UserRole{}, nil
	}

	rows, err := r.db.Q(ctx).Query(ctx,
		`SELECT user_id, role FROM user_roles WHERE user_id = ANY($1) ORDER BY granted_at`,
		userIDs,
	)
	if err != nil {
		return nil, fmt.Errorf("load roles batch: %w", err)
	}
	defer rows.Close()

	result := make(map[string][]entity.UserRole, len(userIDs))
	for rows.Next() {
		var userID, role string
		if err := rows.Scan(&userID, &role); err != nil {
			return nil, fmt.Errorf("scan role row: %w", err)
		}
		result[userID] = append(result[userID], entity.UserRole(role))
	}
	return result, rows.Err()
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func nullStr(s string) any {
	if s == "" {
		return nil
	}
	return s
}
