package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/DoMinhHHung/R2/internal/infrastructure/config"
)

func writeTempRulesFile(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "rules.yaml")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write temp rules file: %v", err)
	}
	return path
}

func TestLoadRules_ValidFile(t *testing.T) {
	yaml := `
routes:
  - prefix: "/api/v1/auth"
    service: auth
    methods: ["POST"]
    require_auth: false
    version: "v1"
  - prefix: "/api/v1/users"
    service: user
    methods: ["*"]
    require_auth: true
    version: "v1"

rbac_rules:
  - service: "user"
    method: "POST"
    path: "/api/v1/auth/*"
    roles: ["*"]
  - service: "user"
    method: "*"
    path: "/api/v1/users/*"
    roles: ["admin", "user"]
`
	path := writeTempRulesFile(t, yaml)
	t.Setenv("GATEWAY_RULES_PATH", path)

	cfg, err := config.LoadRules()
	if err != nil {
		t.Fatalf("LoadRules() error = %v", err)
	}

	if len(cfg.Routes) != 2 {
		t.Errorf("Routes count = %d, want 2", len(cfg.Routes))
	}
	if len(cfg.RBACRules) != 2 {
		t.Errorf("RBACRules count = %d, want 2", len(cfg.RBACRules))
	}

	// Check first route
	r := cfg.Routes[0]
	if r.Prefix != "/api/v1/auth" {
		t.Errorf("Routes[0].Prefix = %q, want %q", r.Prefix, "/api/v1/auth")
	}
	if r.Service != "auth" {
		t.Errorf("Routes[0].Service = %q, want %q", r.Service, "auth")
	}
	if r.RequireAuth {
		t.Error("Routes[0].RequireAuth = true, want false")
	}
	if r.Version != "v1" {
		t.Errorf("Routes[0].Version = %q, want %q", r.Version, "v1")
	}

	// Check second route
	r2 := cfg.Routes[1]
	if !r2.RequireAuth {
		t.Error("Routes[1].RequireAuth = false, want true")
	}

	// Check RBAC rule
	rbac := cfg.RBACRules[0]
	if rbac.Service != "user" {
		t.Errorf("RBACRules[0].Service = %q, want %q", rbac.Service, "user")
	}
	if rbac.Method != "POST" {
		t.Errorf("RBACRules[0].Method = %q, want %q", rbac.Method, "POST")
	}
	if len(rbac.Roles) != 1 || rbac.Roles[0] != "*" {
		t.Errorf("RBACRules[0].Roles = %v, want [*]", rbac.Roles)
	}
}

func TestLoadRules_MissingFile(t *testing.T) {
	t.Setenv("GATEWAY_RULES_PATH", "/nonexistent/path/rules.yaml")

	_, err := config.LoadRules()
	if err == nil {
		t.Error("LoadRules() expected error for missing file, got nil")
	}
}

func TestLoadRules_InvalidYAML(t *testing.T) {
	path := writeTempRulesFile(t, "{ invalid yaml: [")
	t.Setenv("GATEWAY_RULES_PATH", path)

	_, err := config.LoadRules()
	if err == nil {
		t.Error("LoadRules() expected error for invalid YAML, got nil")
	}
}

func TestLoadRules_EmptyRoutes(t *testing.T) {
	yaml := `
routes: []
rbac_rules:
  - service: "user"
    method: "GET"
    path: "/api/v1/users/*"
    roles: ["admin"]
`
	path := writeTempRulesFile(t, yaml)
	t.Setenv("GATEWAY_RULES_PATH", path)

	_, err := config.LoadRules()
	if err == nil {
		t.Error("LoadRules() expected error when routes is empty, got nil")
	}
}

func TestLoadRules_RouteMethods(t *testing.T) {
	yaml := `
routes:
  - prefix: "/api/v1/properties"
    service: property
    methods: ["GET", "POST", "PUT", "DELETE"]
    require_auth: false
    version: "v1"
`
	path := writeTempRulesFile(t, yaml)
	t.Setenv("GATEWAY_RULES_PATH", path)

	cfg, err := config.LoadRules()
	if err != nil {
		t.Fatalf("LoadRules() error = %v", err)
	}

	if len(cfg.Routes[0].Methods) != 4 {
		t.Errorf("Methods count = %d, want 4", len(cfg.Routes[0].Methods))
	}
	want := []string{"GET", "POST", "PUT", "DELETE"}
	for i, m := range want {
		if cfg.Routes[0].Methods[i] != m {
			t.Errorf("Methods[%d] = %q, want %q", i, cfg.Routes[0].Methods[i], m)
		}
	}
}

func TestLoadRules_DefaultPath(t *testing.T) {
	// Clear the env var to trigger default path lookup; default path won't exist in test env
	t.Setenv("GATEWAY_RULES_PATH", "")

	// Should fail because configs/gateway-rules.yaml doesn't exist in test working dir
	_, err := config.LoadRules()
	if err == nil {
		// If it succeeds, that's fine (running from project root with the file)
		t.Log("LoadRules() with default path succeeded (file exists)")
	}
	// Either outcome is acceptable; we just verify no panic
}