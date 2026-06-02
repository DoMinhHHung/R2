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

func TestSplitPath(t *testing.T) {
	tests := []struct {
		name  string
		path  string
		wantN int
		want  []string
	}{
		{
			name:  "standard API path",
			path:  "/api/v1/auth",
			wantN: 3,
			want:  []string{"api", "v1", "auth"},
		},
		{
			name:  "root slash only",
			path:  "/",
			wantN: 0,
			want:  []string{},
		},
		{
			name:  "empty string",
			path:  "",
			wantN: 0,
			want:  []string{},
		},
		{
			name:  "no leading slash",
			path:  "api/v1/users",
			wantN: 3,
			want:  []string{"api", "v1", "users"},
		},
		{
			name:  "trailing slash",
			path:  "/api/v1/",
			wantN: 2,
			want:  []string{"api", "v1"},
		},
		{
			name:  "double slash",
			path:  "/api//v1",
			wantN: 2,
			want:  []string{"api", "v1"},
		},
		{
			name:  "single segment",
			path:  "/metrics",
			wantN: 1,
			want:  []string{"metrics"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := splitPath(tt.path)
			if len(got) != tt.wantN {
				t.Errorf("splitPath(%q) len = %d, want %d; got %v", tt.path, len(got), tt.wantN, got)
				return
			}
			for i, seg := range tt.want {
				if got[i] != seg {
					t.Errorf("splitPath(%q)[%d] = %q, want %q", tt.path, i, got[i], seg)
				}
			}
		})
	}
}