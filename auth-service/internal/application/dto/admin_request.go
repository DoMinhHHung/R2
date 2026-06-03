package dto

type AdminLoginRequest struct {
	Email      string `json:"email"       validate:"required,email"`
	Password   string `json:"password"    validate:"required"`
	DeviceName string `json:"device_name"`
}

type CreateAdminRequest struct {
	Email    string `json:"email"    validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,min=12,max=72"`
}
