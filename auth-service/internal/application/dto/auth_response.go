package dto

import "time"

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
}

type SessionResponse struct {
	ID           string    `json:"id"`
	DeviceName   string    `json:"device_name"`
	DeviceType   string    `json:"device_type"`
	Browser      string    `json:"browser"`
	OS           string    `json:"os"`
	IPAddress    string    `json:"ip_address"`
	LastActivity time.Time `json:"last_activity"`
	ExpiresAt    time.Time `json:"expires_at"`
	CreatedAt    time.Time `json:"created_at"`
}
