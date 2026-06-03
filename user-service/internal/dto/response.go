package dto

import "time"

type UserResponse struct {
	ID               string     `json:"id"`
	Email            string     `json:"email"`
	FullName         string     `json:"full_name,omitempty"`
	PhoneNumber      string     `json:"phone_number,omitempty"`
	Gender           string     `json:"gender,omitempty"`
	DateOfBirth      *time.Time `json:"date_of_birth,omitempty"`
	AvatarURL        string     `json:"avatar_url,omitempty"`
	ProfileCompleted bool       `json:"profile_completed"`
	Roles            []string   `json:"roles"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

type AdminUserResponse struct {
	UserResponse
	Status string `json:"status"`
}

type PaginatedResponse struct {
	Items      any   `json:"items"`
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	TotalPages int   `json:"total_pages"`
}

func NewPaginated(items any, total int64, page, limit int) *PaginatedResponse {
	totalPages := int(total) / limit
	if int(total)%limit != 0 {
		totalPages++
	}
	return &PaginatedResponse{
		Items:      items,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}
}
