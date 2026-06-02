package entity

type RecoverySession struct {
	Email      string `json:"email"`
	OTP        string `json:"otp"`
	RetryCount int    `json:"retry_count"`
}
