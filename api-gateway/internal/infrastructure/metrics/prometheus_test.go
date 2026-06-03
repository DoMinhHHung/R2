package metrics

import (
	"testing"
)

func TestResolveServiceLabel(t *testing.T) {
	tests := []struct {
		name string
		path string
		want string
	}{
		{
			name: "auth service path",
			path: "/api/v1/auth/login",
			want: "auth",
		},
		{
			name: "users service path",
			path: "/api/v1/users/123",
			want: "users",
		},
		{
			name: "properties service path",
			path: "/api/v1/properties",
			want: "properties",
		},
		{
			name: "bookings service path",
			path: "/api/v1/bookings/456",
			want: "bookings",
		},
		{
			name: "payments service path",
			path: "/api/v1/payments",
			want: "payments",
		},
		{
			name: "notifications service path",
			path: "/api/v1/notifications",
			want: "notifications",
		},
		{
			name: "short path returns unknown",
			path: "/short",
			want: "unknown",
		},
		{
			name: "empty path returns unknown",
			path: "",
			want: "unknown",
		},
		{
			name: "path too short returns unknown",
			path: "/api",
			want: "unknown",
		},
		{
			name: "health path only two segments",
			path: "/healthz/live",
			want: "unknown",
		},
		{
			name: "metrics path",
			path: "/metrics",
			want: "unknown",
		},
		{
			name: "exactly at version level returns unknown (only 2 segments)",
			path: "/api/v1",
			want: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolveServiceLabel(tt.path)
			if got != tt.want {
				t.Errorf("resolveServiceLabel(%q) = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}
