package rbac_test

import (
	"context"
	"testing"

	"github.com/DoMinhHHung/R2/internal/domain/entity"
	"github.com/DoMinhHHung/R2/internal/infrastructure/config"
	"github.com/DoMinhHHung/R2/internal/usecase/rbac"
)

func buildRules() []config.RBACRuleItem {
	return []config.RBACRuleItem{
		{Service: "user", Method: "POST", Path: "/api/v1/auth/*", Roles: []string{"*"}},
		{Service: "user", Method: "*", Path: "/api/v1/users/*", Roles: []string{"admin", "user"}},
		{Service: "property", Method: "GET", Path: "/api/v1/properties/*", Roles: []string{"*"}},
		{Service: "property", Method: "POST", Path: "/api/v1/properties/*", Roles: []string{"landlord", "admin"}},
		{Service: "property", Method: "PUT", Path: "/api/v1/properties/*", Roles: []string{"landlord", "admin"}},
		{Service: "property", Method: "PATCH", Path: "/api/v1/properties/*", Roles: []string{"landlord", "admin"}},
		{Service: "property", Method: "DELETE", Path: "/api/v1/properties/*", Roles: []string{"admin"}},
		{Service: "booking", Method: "*", Path: "/api/v1/bookings/*", Roles: []string{"user", "landlord", "admin"}},
		{Service: "payment", Method: "*", Path: "/api/v1/payments/*", Roles: []string{"user", "landlord", "admin"}},
		{Service: "notification", Method: "*", Path: "/api/v1/notifications/*", Roles: []string{"user", "landlord", "admin"}},
	}
}

func TestNew_CreatesUseCase(t *testing.T) {
	uc := rbac.New(buildRules())
	if uc == nil {
		t.Fatal("New() returned nil")
	}
}

func TestNew_EmptyRules(t *testing.T) {
	uc := rbac.New(nil)
	if uc == nil {
		t.Fatal("New() with nil rules returned nil")
	}
	// With no rules, Check should deny everything
	perm := entity.Permission{Service: "user", Method: "GET", Path: "/api/v1/users/1"}
	if uc.Check(context.Background(), perm, []string{"admin"}) {
		t.Error("Check() with no rules should return false")
	}
}

func TestCheck_PublicEndpoints(t *testing.T) {
	uc := rbac.New(buildRules())
	ctx := context.Background()

	tests := []struct {
		name   string
		perm   entity.Permission
		roles  []string
		want   bool
	}{
		{
			name:  "auth POST allowed for anyone (wildcard role)",
			perm:  entity.Permission{Service: "user", Method: "POST", Path: "/api/v1/auth/login"},
			roles: []string{},
			want:  true,
		},
		{
			name:  "auth POST allowed for authenticated user",
			perm:  entity.Permission{Service: "user", Method: "POST", Path: "/api/v1/auth/login"},
			roles: []string{"user"},
			want:  true,
		},
		{
			name:  "property GET allowed for anyone",
			perm:  entity.Permission{Service: "property", Method: "GET", Path: "/api/v1/properties/123"},
			roles: []string{},
			want:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := uc.Check(ctx, tt.perm, tt.roles)
			if got != tt.want {
				t.Errorf("Check() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCheck_ProtectedEndpoints(t *testing.T) {
	uc := rbac.New(buildRules())
	ctx := context.Background()

	tests := []struct {
		name  string
		perm  entity.Permission
		roles []string
		want  bool
	}{
		{
			name:  "users: admin allowed",
			perm:  entity.Permission{Service: "user", Method: "GET", Path: "/api/v1/users/123"},
			roles: []string{"admin"},
			want:  true,
		},
		{
			name:  "users: user role allowed",
			perm:  entity.Permission{Service: "user", Method: "PUT", Path: "/api/v1/users/123"},
			roles: []string{"user"},
			want:  true,
		},
		{
			name:  "users: landlord denied",
			perm:  entity.Permission{Service: "user", Method: "GET", Path: "/api/v1/users/123"},
			roles: []string{"landlord"},
			want:  false,
		},
		{
			name:  "property POST: landlord allowed",
			perm:  entity.Permission{Service: "property", Method: "POST", Path: "/api/v1/properties/new"},
			roles: []string{"landlord"},
			want:  true,
		},
		{
			name:  "property POST: admin allowed",
			perm:  entity.Permission{Service: "property", Method: "POST", Path: "/api/v1/properties/new"},
			roles: []string{"admin"},
			want:  true,
		},
		{
			name:  "property POST: user denied",
			perm:  entity.Permission{Service: "property", Method: "POST", Path: "/api/v1/properties/new"},
			roles: []string{"user"},
			want:  false,
		},
		{
			name:  "property DELETE: only admin allowed",
			perm:  entity.Permission{Service: "property", Method: "DELETE", Path: "/api/v1/properties/123"},
			roles: []string{"admin"},
			want:  true,
		},
		{
			name:  "property DELETE: landlord denied",
			perm:  entity.Permission{Service: "property", Method: "DELETE", Path: "/api/v1/properties/123"},
			roles: []string{"landlord"},
			want:  false,
		},
		{
			name:  "booking: user allowed",
			perm:  entity.Permission{Service: "booking", Method: "POST", Path: "/api/v1/bookings/new"},
			roles: []string{"user"},
			want:  true,
		},
		{
			name:  "payment: landlord allowed",
			perm:  entity.Permission{Service: "payment", Method: "GET", Path: "/api/v1/payments/456"},
			roles: []string{"landlord"},
			want:  true,
		},
		{
			name:  "notification: user allowed",
			perm:  entity.Permission{Service: "notification", Method: "GET", Path: "/api/v1/notifications"},
			roles: []string{"user"},
			want:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := uc.Check(ctx, tt.perm, tt.roles)
			if got != tt.want {
				t.Errorf("Check() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCheck_MethodWildcard(t *testing.T) {
	uc := rbac.New([]config.RBACRuleItem{
		{Service: "booking", Method: "*", Path: "/api/v1/bookings/*", Roles: []string{"user"}},
	})
	ctx := context.Background()

	methods := []string{"GET", "POST", "PUT", "PATCH", "DELETE"}
	for _, m := range methods {
		perm := entity.Permission{Service: "booking", Method: m, Path: "/api/v1/bookings/1"}
		if !uc.Check(ctx, perm, []string{"user"}) {
			t.Errorf("Check() with method wildcard should allow method %s", m)
		}
	}
}

func TestCheck_ServiceMismatch(t *testing.T) {
	uc := rbac.New([]config.RBACRuleItem{
		{Service: "booking", Method: "GET", Path: "/api/v1/bookings/*", Roles: []string{"user"}},
	})
	ctx := context.Background()

	perm := entity.Permission{Service: "payment", Method: "GET", Path: "/api/v1/bookings/1"}
	if uc.Check(ctx, perm, []string{"user"}) {
		t.Error("Check() should deny when service does not match")
	}
}

func TestCheck_ExactPathMatch(t *testing.T) {
	uc := rbac.New([]config.RBACRuleItem{
		{Service: "svc", Method: "GET", Path: "/exact/path", Roles: []string{"user"}},
	})
	ctx := context.Background()

	// Exact match
	perm := entity.Permission{Service: "svc", Method: "GET", Path: "/exact/path"}
	if !uc.Check(ctx, perm, []string{"user"}) {
		t.Error("Check() should allow exact path match")
	}

	// Non-matching path
	perm2 := entity.Permission{Service: "svc", Method: "GET", Path: "/exact/path/extra"}
	if uc.Check(ctx, perm2, []string{"user"}) {
		t.Error("Check() should deny when path does not exactly match (no wildcard)")
	}
}

func TestCheck_MultipleRoles(t *testing.T) {
	uc := rbac.New([]config.RBACRuleItem{
		{Service: "svc", Method: "GET", Path: "/api/*", Roles: []string{"admin", "mod"}},
	})
	ctx := context.Background()

	// User with one allowed role
	if !uc.Check(ctx, entity.Permission{Service: "svc", Method: "GET", Path: "/api/foo"}, []string{"mod"}) {
		t.Error("Check() should allow user with 'mod' role")
	}
	// User with no matching role
	if uc.Check(ctx, entity.Permission{Service: "svc", Method: "GET", Path: "/api/foo"}, []string{"user"}) {
		t.Error("Check() should deny user with no matching role")
	}
	// User with both roles
	if !uc.Check(ctx, entity.Permission{Service: "svc", Method: "GET", Path: "/api/foo"}, []string{"user", "admin"}) {
		t.Error("Check() should allow user with 'admin' role in list")
	}
}

func TestCheck_WildcardPath(t *testing.T) {
	uc := rbac.New([]config.RBACRuleItem{
		{Service: "svc", Method: "GET", Path: "*", Roles: []string{"user"}},
	})
	ctx := context.Background()

	paths := []string{"/anything", "/api/v1/foo", "/", ""}
	for _, p := range paths {
		perm := entity.Permission{Service: "svc", Method: "GET", Path: p}
		if !uc.Check(ctx, perm, []string{"user"}) {
			t.Errorf("Check() with path wildcard '*' should allow path %q", p)
		}
	}
}

func TestCheck_NoRoles(t *testing.T) {
	uc := rbac.New([]config.RBACRuleItem{
		{Service: "svc", Method: "GET", Path: "/api/*", Roles: []string{"admin"}},
	})
	ctx := context.Background()

	perm := entity.Permission{Service: "svc", Method: "GET", Path: "/api/foo"}
	if uc.Check(ctx, perm, []string{}) {
		t.Error("Check() should deny user with empty roles when rule requires admin")
	}
	if uc.Check(ctx, perm, nil) {
		t.Error("Check() should deny user with nil roles when rule requires admin")
	}
}