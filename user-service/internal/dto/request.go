package dto

type UpdateProfileRequest struct {
	FullName    string `json:"full_name"    validate:"required,min=2,max=255"`
	PhoneNumber string `json:"phone_number" validate:"required,min=9,max=20"`
	Gender      string `json:"gender"       validate:"required,oneof=MALE FEMALE OTHER"`
	DateOfBirth string `json:"date_of_birth" validate:"required,datetime=2006-01-02"`
}

type AdminBanRequest struct {
	Reason string `json:"reason" validate:"required,min=5,max=500"`
}

type InternalCreateUserRequest struct {
	UserID string `json:"user_id" validate:"required,uuid"`
	Email  string `json:"email"   validate:"required,email"`
	Role   string `json:"role"    validate:"required,oneof=TENANT LANDLORD ADMIN"`
}
