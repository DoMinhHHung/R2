package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DoMinhHHung/R2/internal/domain/entity"
	"github.com/gin-gonic/gin"
)

// mockRateLimiter implements port.RateLimiter.
type mockRateLimiter struct {
	allowed bool
	err     error
	calls   []entity.RateLimitKey
}

func (m *mockRateLimiter) Allow(_ context.Context, key entity.RateLimitKey) (bool, error) {
	m.calls = append(m.calls, key)
	return m.allowed, m.err
}

// ---- buildKeys tests ----

func TestBuildKeys_IPOnly(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	var keys []entity.RateLimitKey
	r.GET("/test", func(c *gin.Context) {
		keys = buildKeys(c)
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "10.0.0.1:1234"
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if len(keys) != 1 {
		t.Fatalf("buildKeys() len = %d, want 1", len(keys))
	}
	if keys[0].Type != "ip" {
		t.Errorf("buildKeys()[0].Type = %q, want %q", keys[0].Type, "ip")
	}
}

func TestBuildKeys_WithUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	var keys []entity.RateLimitKey
	r.GET("/test", func(c *gin.Context) {
		keys = buildKeys(c)
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-User-ID", "user-123")
	req.RemoteAddr = "10.0.0.1:1234"
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if len(keys) != 2 {
		t.Fatalf("buildKeys() len = %d, want 2 (ip + user)", len(keys))
	}
	found := false
	for _, k := range keys {
		if k.Type == "user" && k.Value == "user-123" {
			found = true
		}
	}
	if !found {
		t.Error("buildKeys() should include user key when X-User-ID header is set")
	}
}

func TestBuildKeys_WithAPIKey(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	var keys []entity.RateLimitKey
	r.GET("/test", func(c *gin.Context) {
		keys = buildKeys(c)
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-API-Key", "key-xyz")
	req.RemoteAddr = "10.0.0.1:1234"
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if len(keys) < 2 {
		t.Fatalf("buildKeys() len = %d, want at least 2", len(keys))
	}
	found := false
	for _, k := range keys {
		if k.Type == "apikey" && k.Value == "key-xyz" {
			found = true
		}
	}
	if !found {
		t.Error("buildKeys() should include apikey key when X-API-Key header is set")
	}
}

func TestBuildKeys_AllThree(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	var keys []entity.RateLimitKey
	r.GET("/test", func(c *gin.Context) {
		keys = buildKeys(c)
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-User-ID", "user-1")
	req.Header.Set("X-API-Key", "api-key-1")
	req.RemoteAddr = "10.0.0.1:1234"
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if len(keys) != 3 {
		t.Fatalf("buildKeys() len = %d, want 3 (ip + user + apikey)", len(keys))
	}
	types := map[string]string{}
	for _, k := range keys {
		types[k.Type] = k.Value
	}
	if _, ok := types["ip"]; !ok {
		t.Error("buildKeys() should contain ip key")
	}
	if types["user"] != "user-1" {
		t.Errorf("buildKeys() user value = %q, want %q", types["user"], "user-1")
	}
	if types["apikey"] != "api-key-1" {
		t.Errorf("buildKeys() apikey value = %q, want %q", types["apikey"], "api-key-1")
	}
}

// ---- RateLimit middleware tests ----

func TestRateLimit_AllowedRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rl := &mockRateLimiter{allowed: true}

	r := gin.New()
	r.GET("/test", RateLimit(rl), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "1.2.3.4:0"
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("RateLimit() with allowed = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestRateLimit_DeniedRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rl := &mockRateLimiter{allowed: false}

	r := gin.New()
	r.GET("/test", RateLimit(rl), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "1.2.3.4:0"
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("RateLimit() when denied = %d, want %d", w.Code, http.StatusTooManyRequests)
	}
}

func TestRateLimit_ErrorFailsOpen(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rl := &mockRateLimiter{allowed: false, err: errors.New("redis error")}

	r := gin.New()
	r.GET("/test", RateLimit(rl), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "1.2.3.4:0"
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// On error, middleware should fail open (pass the request through)
	if w.Code != http.StatusOK {
		t.Errorf("RateLimit() on error should fail open, got %d, want %d", w.Code, http.StatusOK)
	}
}

func TestRateLimit_ResponseBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rl := &mockRateLimiter{allowed: false}

	r := gin.New()
	r.GET("/test", RateLimit(rl), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "1.2.3.4:0"
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	body := w.Body.String()
	if body == "" {
		t.Error("RateLimit() 429 response should have a body")
	}
}

// ---- buildRLKey tests ----

func TestBuildRLKey(t *testing.T) {
	tests := []struct {
		key  entity.RateLimitKey
		want string
	}{
		{entity.RateLimitKey{Type: "ip", Value: "1.2.3.4"}, "rl:ip:1.2.3.4"},
		{entity.RateLimitKey{Type: "user", Value: "user-123"}, "rl:user:user-123"},
		{entity.RateLimitKey{Type: "apikey", Value: "abc"}, "rl:apikey:abc"},
		{entity.RateLimitKey{Type: "", Value: ""}, "rl::"},
	}

	for _, tt := range tests {
		got := buildRLKey(tt.key)
		if got != tt.want {
			t.Errorf("buildRLKey(%v) = %q, want %q", tt.key, got, tt.want)
		}
	}
}