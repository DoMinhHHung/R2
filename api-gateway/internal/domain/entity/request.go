package entity

import "time"

type GatewayRequest struct {
	RequestID string
	UserID    string
	APIKey    string
	ClientIP  string
	Method    string
	Path      string
	Service   string
	Roles     []string
	StartTime time.Time
}

type RateLimitKey struct {
	Type  string
	Value string
}

type Permission struct {
	Service string
	Method  string
	Path    string
	Roles   []string
}
