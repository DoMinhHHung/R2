package middleware

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/DoMinhHHung/R2/internal/domain/entity"
	infraJWT "github.com/DoMinhHHung/R2/internal/infrastructure/jwt"
	"github.com/gin-gonic/gin"
	gojwt "github.com/golang-jwt/jwt/v5"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// generateTestRSAKey creates an RSA key pair and writes the public key to a temp file.
// Returns the private key and the path to the public key PEM file.
func generateTestRSAKey(t *testing.T) (*rsa.PrivateKey, string) {
	t.Helper()
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate RSA key: %v", err)
	}

	pubKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		t.Fatalf("marshal public key: %v", err)
	}

	pubPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubKeyBytes,
	})

	dir := t.TempDir()
	pubKeyPath := filepath.Join(dir, "public.pem")
	if err := os.WriteFile(pubKeyPath, pubPEM, 0644); err != nil {
		t.Fatalf("write public key: %v", err)
	}

	return privateKey, pubKeyPath
}

// buildTestJWT creates a signed JWT token for tests.
func buildTestJWT(t *testing.T, privateKey *rsa.PrivateKey, userID, role, issuer string, exp time.Time) string {
	t.Helper()
	type testClaims struct {
		UserID string `json:"uid"`
		Email  string `json:"email"`
		Role   string `json:"role"`
		gojwt.RegisteredClaims
	}

	c := testClaims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: gojwt.RegisteredClaims{
			Issuer:    issuer,
			ExpiresAt: gojwt.NewNumericDate(exp),
			IssuedAt:  gojwt.NewNumericDate(time.Now()),
		},
	}

	token := gojwt.NewWithClaims(gojwt.SigningMethodRS256, c)
	signed, err := token.SignedString(privateKey)
	if err != nil {
		t.Fatalf("sign JWT: %v", err)
	}
	return signed
}

func newTestRouter(handlers ...gin.HandlerFunc) (*gin.Engine, *httptest.ResponseRecorder) {
	r := gin.New()
	r.GET("/test", handlers...)
	w := httptest.NewRecorder()
	return r, w
}

// ---- resolveService tests (unexported, tested within package) ----

func TestResolveService(t *testing.T) {
	tests := []struct {
		name string
		path string
		want string
	}{
		{"auth path", "/api/v1/auth/login", "auth"},
		{"users path", "/api/v1/users/123", "users"},
		{"properties path", "/api/v1/properties", "properties"},
		{"bookings path", "/api/v1/bookings/456", "bookings"},
		{"only two segments", "/api/v1", ""},
		{"empty path", "", ""},
		{"single segment", "/foo", ""},
		{"root", "/", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolveService(tt.path)
			if got != tt.want {
				t.Errorf("resolveService(%q) = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}

// ---- Auth middleware tests ----

func TestAuth_MissingAuthorizationHeader(t *testing.T) {
	privateKey, pubKeyPath := generateTestRSAKey(t)
	_ = privateKey // not needed for this test

	parser, err := infraJWT.NewParser(pubKeyPath, "test-issuer")
	if err != nil {
		t.Fatalf("NewParser: %v", err)
	}

	r := gin.New()
	r.GET("/test", Auth(parser), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Auth() with no header = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestAuth_NonBearerHeader(t *testing.T) {
	privateKey, pubKeyPath := generateTestRSAKey(t)
	_ = privateKey

	parser, err := infraJWT.NewParser(pubKeyPath, "test-issuer")
	if err != nil {
		t.Fatalf("NewParser: %v", err)
	}

	r := gin.New()
	r.GET("/test", Auth(parser), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Basic sometoken")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Auth() with non-Bearer header = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestAuth_InvalidToken(t *testing.T) {
	_, pubKeyPath := generateTestRSAKey(t)

	parser, err := infraJWT.NewParser(pubKeyPath, "test-issuer")
	if err != nil {
		t.Fatalf("NewParser: %v", err)
	}

	r := gin.New()
	r.GET("/test", Auth(parser), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer not.a.real.jwt")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Auth() with invalid token = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestAuth_ValidToken(t *testing.T) {
	privateKey, pubKeyPath := generateTestRSAKey(t)

	parser, err := infraJWT.NewParser(pubKeyPath, "test-issuer")
	if err != nil {
		t.Fatalf("NewParser: %v", err)
	}

	tokenStr := buildTestJWT(t, privateKey, "user-1", "admin", "test-issuer", time.Now().Add(1*time.Hour))

	var gotUserID string
	var gotRoles []string

	r := gin.New()
	r.GET("/test", Auth(parser), func(c *gin.Context) {
		gotUserID = c.GetString("user_id")
		rolesRaw, _ := c.Get("user_roles")
		gotRoles, _ = rolesRaw.([]string)
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Auth() with valid token = %d, want %d", w.Code, http.StatusOK)
	}
	if gotUserID != "user-1" {
		t.Errorf("Auth() user_id = %q, want %q", gotUserID, "user-1")
	}
	if len(gotRoles) != 1 || gotRoles[0] != "admin" {
		t.Errorf("Auth() user_roles = %v, want [admin]", gotRoles)
	}
}

func TestAuth_ExpiredToken(t *testing.T) {
	privateKey, pubKeyPath := generateTestRSAKey(t)

	parser, err := infraJWT.NewParser(pubKeyPath, "test-issuer")
	if err != nil {
		t.Fatalf("NewParser: %v", err)
	}

	tokenStr := buildTestJWT(t, privateKey, "user-1", "admin", "test-issuer", time.Now().Add(-1*time.Hour))

	r := gin.New()
	r.GET("/test", Auth(parser), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Auth() with expired token = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

// ---- OptionalAuth middleware tests ----

func TestOptionalAuth_NoHeader_PassesThrough(t *testing.T) {
	privateKey, pubKeyPath := generateTestRSAKey(t)
	_ = privateKey

	parser, err := infraJWT.NewParser(pubKeyPath, "test-issuer")
	if err != nil {
		t.Fatalf("NewParser: %v", err)
	}

	called := false
	r := gin.New()
	r.GET("/test", OptionalAuth(parser), func(c *gin.Context) {
		called = true
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("OptionalAuth() without header = %d, want %d", w.Code, http.StatusOK)
	}
	if !called {
		t.Error("OptionalAuth() should call next handler even without token")
	}
}

func TestOptionalAuth_InvalidToken_PassesThrough(t *testing.T) {
	_, pubKeyPath := generateTestRSAKey(t)

	parser, err := infraJWT.NewParser(pubKeyPath, "test-issuer")
	if err != nil {
		t.Fatalf("NewParser: %v", err)
	}

	called := false
	r := gin.New()
	r.GET("/test", OptionalAuth(parser), func(c *gin.Context) {
		called = true
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer bad.jwt.token")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("OptionalAuth() with invalid token = %d, want %d", w.Code, http.StatusOK)
	}
	if !called {
		t.Error("OptionalAuth() should call next handler even with invalid token")
	}
}

func TestOptionalAuth_ValidToken_SetsClaims(t *testing.T) {
	privateKey, pubKeyPath := generateTestRSAKey(t)

	parser, err := infraJWT.NewParser(pubKeyPath, "test-issuer")
	if err != nil {
		t.Fatalf("NewParser: %v", err)
	}

	tokenStr := buildTestJWT(t, privateKey, "user-42", "user", "test-issuer", time.Now().Add(1*time.Hour))

	var gotUserID interface{}
	r := gin.New()
	r.GET("/test", OptionalAuth(parser), func(c *gin.Context) {
		gotUserID, _ = c.Get("user_id")
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("OptionalAuth() with valid token = %d, want %d", w.Code, http.StatusOK)
	}
	if gotUserID != "user-42" {
		t.Errorf("OptionalAuth() user_id = %v, want %q", gotUserID, "user-42")
	}
}

// ---- RBAC middleware tests ----

type mockRBACChecker struct {
	result bool
}

func (m *mockRBACChecker) Check(_ context.Context, _ entity.Permission, _ []string) bool {
	return m.result
}

func TestRBAC_NoClaims_Forbidden(t *testing.T) {
	checker := &mockRBACChecker{result: true}

	r := gin.New()
	r.GET("/test", RBAC(checker), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("RBAC() without claims = %d, want %d", w.Code, http.StatusForbidden)
	}
}

func TestRBAC_InvalidClaimsType_Forbidden(t *testing.T) {
	checker := &mockRBACChecker{result: true}

	r := gin.New()
	r.GET("/test",
		func(c *gin.Context) {
			c.Set("claims", "not-a-claims-struct")
			c.Next()
		},
		RBAC(checker),
		func(c *gin.Context) {
			c.Status(http.StatusOK)
		},
	)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("RBAC() with invalid claims type = %d, want %d", w.Code, http.StatusForbidden)
	}
}

func TestRBAC_CheckDenied_Forbidden(t *testing.T) {
	checker := &mockRBACChecker{result: false}

	r := gin.New()
	r.GET("/api/v1/properties/new",
		func(c *gin.Context) {
			c.Set("claims", &entity.Claims{UserID: "u1", Roles: []string{"user"}})
			c.Next()
		},
		RBAC(checker),
		func(c *gin.Context) {
			c.Status(http.StatusOK)
		},
	)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/properties/new", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("RBAC() when checker denies = %d, want %d", w.Code, http.StatusForbidden)
	}
}

func TestRBAC_CheckAllowed_Proceeds(t *testing.T) {
	checker := &mockRBACChecker{result: true}

	r := gin.New()
	r.GET("/api/v1/properties/new",
		func(c *gin.Context) {
			c.Set("claims", &entity.Claims{UserID: "u1", Roles: []string{"admin"}})
			c.Next()
		},
		RBAC(checker),
		func(c *gin.Context) {
			c.Status(http.StatusOK)
		},
	)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/properties/new", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("RBAC() when checker allows = %d, want %d", w.Code, http.StatusOK)
	}
}