package entity

import "time"

type Session struct {
	ID               string
	UserID           string
	RefreshTokenHash string
	DeviceName       string
	DeviceType       string
	Browser          string
	OS               string
	IPAddress        string
	UserAgent        string
	LastActivity     time.Time
	ExpiresAt        time.Time
	CreatedAt        time.Time
}

func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}
