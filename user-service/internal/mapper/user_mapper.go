package mapper

import (
	"github.com/DoMinhHHung/user-service/internal/domain/entity"
	"github.com/DoMinhHHung/user-service/internal/dto"
)

func ToUserResponse(u *entity.User) *dto.UserResponse {
	roles := make([]string, len(u.Roles))
	for i, r := range u.Roles {
		roles[i] = string(r)
	}

	return &dto.UserResponse{
		ID:               u.ID,
		Email:            u.Email,
		FullName:         u.FullName,
		PhoneNumber:      u.PhoneNumber,
		Gender:           string(u.Gender),
		DateOfBirth:      u.DateOfBirth,
		AvatarURL:        u.AvatarURL,
		ProfileCompleted: u.ProfileCompleted,
		Roles:            roles,
		CreatedAt:        u.CreatedAt,
		UpdatedAt:        u.UpdatedAt,
	}
}

func ToAdminUserResponse(u *entity.User) *dto.AdminUserResponse {
	return &dto.AdminUserResponse{
		UserResponse: *ToUserResponse(u),
		Status:       string(u.Status),
	}
}

func ToAdminUserResponseList(users []*entity.User) []*dto.AdminUserResponse {
	result := make([]*dto.AdminUserResponse, len(users))
	for i, u := range users {
		result[i] = ToAdminUserResponse(u)
	}
	return result
}
