package entity

type SignupSession struct {
	Email        string `json:"email"`
	PasswordHash string `json:"password_hash"`
	Role         Role   `json:"role"`
	OTP          string `json:"otp"`
	RetryCount   int    `json:"retry_count"`
}
