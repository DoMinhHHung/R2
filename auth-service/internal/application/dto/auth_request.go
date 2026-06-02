package dto

type SignupRequest struct {
	Email    string `json:"email"    validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,min=8,max=72"`
	Role     string `json:"role"     validate:"required,oneof=TENANT LANDLORD"`
}

type VerifyOTPRequest struct {
	Email string `json:"email" validate:"required,email"`
	OTP   string `json:"otp"   validate:"required,len=6,numeric"`
}

type ResendOTPRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type LoginRequest struct {
	Email      string `json:"email"    validate:"required,email"`
	Password   string `json:"password" validate:"required"`
	DeviceName string `json:"device_name"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type VerifyRecoveryOTPRequest struct {
	Email string `json:"email" validate:"required,email"`
	OTP   string `json:"otp"   validate:"required,len=6,numeric"`
}

type ResetPasswordRequest struct {
	ResetToken  string `json:"reset_token"  validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8,max=72"`
}
