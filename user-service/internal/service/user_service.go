package service

import (
	"context"
	"fmt"
	"mime/multipart"
	"time"

	"github.com/DoMinhHHung/user-service/internal/domain/entity"
	"github.com/DoMinhHHung/user-service/internal/domain/port"
	"github.com/DoMinhHHung/user-service/internal/dto"
	"github.com/DoMinhHHung/user-service/internal/logger"
	"github.com/DoMinhHHung/user-service/internal/mapper"
	"github.com/DoMinhHHung/user-service/pkg/apperr"
)

type UserService struct {
	repo    port.UserRepository
	storage port.Storage
	log     *logger.Logger
}

func New(repo port.UserRepository, storage port.Storage, log *logger.Logger) *UserService {
	return &UserService{repo: repo, storage: storage, log: log}
}

func (s *UserService) CreateUserFromEvent(ctx context.Context, userID, email string, role entity.UserRole) error {
	exists, err := s.repo.ExistsByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("check exists: %w", err)
	}
	if exists {
		s.log.Info("user already exists, skipping", "user_id", userID)
		return nil
	}

	now := time.Now()
	user := &entity.User{
		ID:        userID,
		Email:     email,
		Status:    entity.StatusActive,
		Roles:     []entity.UserRole{role},
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return fmt.Errorf("create user: %w", err)
	}

	s.log.Info("user profile created", "user_id", userID, "email", email, "role", role)
	return nil
}

func (s *UserService) GetMyProfile(ctx context.Context, userID string) (*dto.UserResponse, error) {
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if err := s.checkActive(user); err != nil {
		return nil, err
	}

	return mapper.ToUserResponse(user), nil
}

func (s *UserService) UpdateMyProfile(ctx context.Context, userID string, req *dto.UpdateProfileRequest) (*dto.UserResponse, error) {
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if err := s.checkActive(user); err != nil {
		return nil, err
	}

	dob, err := time.Parse("2006-01-02", req.DateOfBirth)
	if err != nil {
		return nil, apperr.ErrInvalidInput
	}

	user.FullName = req.FullName
	user.PhoneNumber = req.PhoneNumber
	user.Gender = entity.UserGender(req.Gender)
	user.DateOfBirth = &dob

	user.ProfileCompleted = user.IsProfileComplete()

	if err := s.repo.Update(ctx, user); err != nil {
		return nil, err
	}

	updated, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return mapper.ToUserResponse(updated), nil
}

func (s *UserService) GetUserByID(ctx context.Context, id string) (*dto.UserResponse, error) {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if user.Status == entity.StatusBanned || user.Status == entity.StatusDeleted {
		return nil, apperr.ErrUserNotFound
	}
	resp := mapper.ToUserResponse(user)
	resp.PhoneNumber = ""
	resp.Email = ""
	return resp, nil
}

func (s *UserService) UploadAvatar(
	ctx context.Context,
	userID string,
	file multipart.File,
	header *multipart.FileHeader,
) (*dto.UserResponse, error) {
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if err := s.checkActive(user); err != nil {
		return nil, err
	}

	if user.AvatarPublicID != "" {
		if err := s.storage.Delete(ctx, user.AvatarPublicID); err != nil {
			s.log.Warn("failed to delete old avatar, continuing", "public_id", user.AvatarPublicID, "error", err)
		}
	}

	result, err := s.storage.Upload(ctx, file, header, userID)
	if err != nil {
		return nil, err
	}

	if err := s.repo.UpdateAvatar(ctx, userID, result.URL, result.PublicID); err != nil {
		_ = s.storage.Delete(ctx, result.PublicID)
		return nil, err
	}

	updated, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return mapper.ToUserResponse(updated), nil
}

func (s *UserService) DeleteAvatar(ctx context.Context, userID string) error {
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	if err := s.checkActive(user); err != nil {
		return err
	}
	if user.AvatarPublicID == "" {
		return apperr.ErrNoAvatar
	}

	if err := s.storage.Delete(ctx, user.AvatarPublicID); err != nil {
		s.log.Error("failed to delete from cloudinary", "public_id", user.AvatarPublicID, "error", err)
	}

	return s.repo.DeleteAvatar(ctx, userID)
}

// ─── Admin ────────────────────────────────────────────────────────────────────

func (s *UserService) AdminListUsers(ctx context.Context, filter port.UserFilter) (*dto.PaginatedResponse, error) {
	users, total, err := s.repo.ListUsers(ctx, filter)
	if err != nil {
		return nil, err
	}
	return dto.NewPaginated(mapper.ToAdminUserResponseList(users), total, filter.Page, filter.Limit), nil
}

func (s *UserService) AdminGetUser(ctx context.Context, id string) (*dto.AdminUserResponse, error) {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return mapper.ToAdminUserResponse(user), nil
}

func (s *UserService) BanUser(ctx context.Context, targetID, adminID string) error {
	if targetID == adminID {
		return apperr.New("SELF_BAN", "Cannot ban yourself", 400)
	}
	user, err := s.repo.FindByID(ctx, targetID)
	if err != nil {
		return err
	}
	if user.Status == entity.StatusBanned {
		return apperr.New("ALREADY_BANNED", "User is already banned", 400)
	}
	s.log.Info("banning user", "target_id", targetID, "admin_id", adminID)
	return s.repo.UpdateStatus(ctx, targetID, entity.StatusBanned)
}

func (s *UserService) UnbanUser(ctx context.Context, targetID string) error {
	user, err := s.repo.FindByID(ctx, targetID)
	if err != nil {
		return err
	}
	if user.Status != entity.StatusBanned {
		return apperr.New("NOT_BANNED", "User is not banned", 400)
	}
	return s.repo.UpdateStatus(ctx, targetID, entity.StatusActive)
}

func (s *UserService) checkActive(u *entity.User) error {
	switch u.Status {
	case entity.StatusBanned:
		return apperr.ErrUserBanned
	case entity.StatusDeleted:
		return apperr.ErrUserDeleted
	}
	return nil
}
