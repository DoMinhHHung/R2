package entity_test

import (
	"testing"
	"time"

	"github.com/DoMinhHHung/R2/internal/domain/entity"
)

func TestGatewayRequest_ZeroValue(t *testing.T) {
	var r entity.GatewayRequest
	if r.RequestID != "" {
		t.Errorf("zero GatewayRequest.RequestID = %q, want empty string", r.RequestID)
	}
	if r.UserID != "" {
		t.Errorf("zero GatewayRequest.UserID = %q, want empty string", r.UserID)
	}
	if !r.StartTime.IsZero() {
		t.Error("zero GatewayRequest.StartTime should be zero time")
	}
	if r.Roles != nil {
		t.Errorf("zero GatewayRequest.Roles = %v, want nil", r.Roles)
	}
}

func TestGatewayRequest_FieldAssignment(t *testing.T) {
	now := time.Now()
	r := entity.GatewayRequest{
		RequestID: "req-123",
		UserID:    "user-456",
		APIKey:    "key-789",
		ClientIP:  "192.168.1.1",
		Method:    "POST",
		Path:      "/api/v1/users",
		Service:   "user",
		Roles:     []string{"admin", "user"},
		StartTime: now,
	}

	if r.RequestID != "req-123" {
		t.Errorf("RequestID = %q, want req-123", r.RequestID)
	}
	if r.UserID != "user-456" {
		t.Errorf("UserID = %q, want user-456", r.UserID)
	}
	if r.APIKey != "key-789" {
		t.Errorf("APIKey = %q, want key-789", r.APIKey)
	}
	if r.ClientIP != "192.168.1.1" {
		t.Errorf("ClientIP = %q, want 192.168.1.1", r.ClientIP)
	}
	if r.Method != "POST" {
		t.Errorf("Method = %q, want POST", r.Method)
	}
	if r.Path != "/api/v1/users" {
		t.Errorf("Path = %q, want /api/v1/users", r.Path)
	}
	if r.Service != "user" {
		t.Errorf("Service = %q, want user", r.Service)
	}
	if len(r.Roles) != 2 {
		t.Errorf("Roles len = %d, want 2", len(r.Roles))
	}
	if !r.StartTime.Equal(now) {
		t.Errorf("StartTime = %v, want %v", r.StartTime, now)
	}
}

func TestRateLimitKey_ZeroValue(t *testing.T) {
	var k entity.RateLimitKey
	if k.Type != "" {
		t.Errorf("zero RateLimitKey.Type = %q, want empty", k.Type)
	}
	if k.Value != "" {
		t.Errorf("zero RateLimitKey.Value = %q, want empty", k.Value)
	}
}

func TestRateLimitKey_FieldAssignment(t *testing.T) {
	tests := []struct {
		keyType string
		value   string
	}{
		{"ip", "10.0.0.1"},
		{"user", "user-uuid-123"},
		{"apikey", "secret-key-abc"},
	}

	for _, tt := range tests {
		k := entity.RateLimitKey{Type: tt.keyType, Value: tt.value}
		if k.Type != tt.keyType {
			t.Errorf("RateLimitKey.Type = %q, want %q", k.Type, tt.keyType)
		}
		if k.Value != tt.value {
			t.Errorf("RateLimitKey.Value = %q, want %q", k.Value, tt.value)
		}
	}
}

func TestPermission_ZeroValue(t *testing.T) {
	var p entity.Permission
	if p.Service != "" || p.Method != "" || p.Path != "" {
		t.Error("zero Permission fields should be empty strings")
	}
	if p.Roles != nil {
		t.Errorf("zero Permission.Roles = %v, want nil", p.Roles)
	}
}

func TestPermission_FieldAssignment(t *testing.T) {
	p := entity.Permission{
		Service: "property",
		Method:  "DELETE",
		Path:    "/api/v1/properties/123",
		Roles:   []string{"admin"},
	}

	if p.Service != "property" {
		t.Errorf("Permission.Service = %q, want property", p.Service)
	}
	if p.Method != "DELETE" {
		t.Errorf("Permission.Method = %q, want DELETE", p.Method)
	}
	if p.Path != "/api/v1/properties/123" {
		t.Errorf("Permission.Path = %q, want /api/v1/properties/123", p.Path)
	}
	if len(p.Roles) != 1 || p.Roles[0] != "admin" {
		t.Errorf("Permission.Roles = %v, want [admin]", p.Roles)
	}
}

func TestPermission_EmptyRoles(t *testing.T) {
	p := entity.Permission{
		Service: "svc",
		Method:  "GET",
		Path:    "/path",
		Roles:   []string{},
	}
	if len(p.Roles) != 0 {
		t.Errorf("Permission.Roles with empty slice len = %d, want 0", len(p.Roles))
	}
}