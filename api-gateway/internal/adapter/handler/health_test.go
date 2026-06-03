package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DoMinhHHung/R2/internal/adapter/handler"
	"github.com/DoMinhHHung/R2/internal/infrastructure/cache"
	"github.com/DoMinhHHung/R2/internal/infrastructure/config"
	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func newHealthTestRouter(redis *cache.RedisCache) *gin.Engine {
	r := gin.New()
	h := handler.NewHealth(redis)
	r.GET("/healthz/live", h.Live)
	r.GET("/healthz/ready", h.Ready)
	return r
}

func TestLive_Returns200(t *testing.T) {
	redis := cache.NewRedis(config.RedisConfig{Addr: "localhost:6379"})
	r := newHealthTestRouter(redis)

	req := httptest.NewRequest(http.MethodGet, "/healthz/live", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Live() = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestLive_ReturnsStatusOK(t *testing.T) {
	redis := cache.NewRedis(config.RedisConfig{Addr: "localhost:6379"})
	r := newHealthTestRouter(redis)

	req := httptest.NewRequest(http.MethodGet, "/healthz/live", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var body map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("Live() response is not valid JSON: %v", err)
	}

	status, ok := body["status"]
	if !ok {
		t.Fatal("Live() response missing 'status' field")
	}
	if status != "ok" {
		t.Errorf("Live() status = %q, want %q", status, "ok")
	}
}

func TestLive_ContentTypeJSON(t *testing.T) {
	redis := cache.NewRedis(config.RedisConfig{Addr: "localhost:6379"})
	r := newHealthTestRouter(redis)

	req := httptest.NewRequest(http.MethodGet, "/healthz/live", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	ct := w.Header().Get("Content-Type")
	if ct == "" {
		t.Error("Live() should set Content-Type header")
	}
}

func TestReady_DegradedWhenRedisFails(t *testing.T) {
	// Use an unreachable Redis address to simulate failure
	redis := cache.NewRedis(config.RedisConfig{Addr: "localhost:19999"})
	r := newHealthTestRouter(redis)

	req := httptest.NewRequest(http.MethodGet, "/healthz/ready", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("Ready() with failing Redis = %d, want %d", w.Code, http.StatusServiceUnavailable)
	}

	var body map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("Ready() response is not valid JSON: %v", err)
	}

	if body["status"] != "degraded" {
		t.Errorf("Ready() status = %v, want 'degraded'", body["status"])
	}

	deps, ok := body["deps"].(map[string]interface{})
	if !ok {
		t.Fatal("Ready() response missing 'deps' field")
	}
	if deps["redis"] != false {
		t.Errorf("Ready() deps.redis = %v, want false when Redis fails", deps["redis"])
	}
}

func TestReady_DegradedResponseShape(t *testing.T) {
	redis := cache.NewRedis(config.RedisConfig{Addr: "localhost:19999"})
	r := newHealthTestRouter(redis)

	req := httptest.NewRequest(http.MethodGet, "/healthz/ready", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var body map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("Ready() degraded response is not valid JSON: %v", err)
	}

	if _, ok := body["status"]; !ok {
		t.Error("Ready() response missing 'status' field")
	}
	if _, ok := body["deps"]; !ok {
		t.Error("Ready() response missing 'deps' field")
	}
}