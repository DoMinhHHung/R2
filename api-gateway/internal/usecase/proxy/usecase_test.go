package proxy

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/DoMinhHHung/R2/internal/domain/entity"
	"github.com/DoMinhHHung/R2/internal/infrastructure/config"
)

// ---- Resolve tests ----

func buildTestConfig(serviceURL string) config.Config {
	return config.Config{
		Services: config.ServicesConfig{
			Auth:         serviceURL,
			User:         serviceURL,
			Property:     serviceURL,
			Booking:      serviceURL,
			Payment:      serviceURL,
			Notification: serviceURL,
		},
		CB: config.CircuitBreakerConfig{
			MaxRequests:  5,
			IntervalSecs: 10,
			TimeoutSecs:  30,
			FailureRatio: 0.6,
		},
		Retry: config.RetryConfig{MaxAttempts: 1},
	}
}

func TestResolve_MatchesPrefix(t *testing.T) {
	uc := New(buildTestConfig("http://localhost:9999"))

	tests := []struct {
		path  string
		found bool
	}{
		{"/api/v1/auth/login", true},
		{"/api/v1/users/123", true},
		{"/api/v1/properties", true},
		{"/api/v1/bookings/456", true},
		{"/api/v1/payments", true},
		{"/api/v1/notifications", true},
		{"/api/v1/unknown", false},
		{"/healthz/live", false},
		{"", false},
		{"/api/v1", false},
	}

	for _, tt := range tests {
		route, ok := uc.Resolve(tt.path)
		if ok != tt.found {
			t.Errorf("Resolve(%q) found=%v, want %v", tt.path, ok, tt.found)
		}
		if ok && route == nil {
			t.Errorf("Resolve(%q) returned nil route", tt.path)
		}
	}
}

func TestResolve_ReturnsCorrectRoute(t *testing.T) {
	uc := New(buildTestConfig("http://localhost:9999"))

	route, ok := uc.Resolve("/api/v1/bookings/new")
	if !ok {
		t.Fatal("Resolve(/api/v1/bookings/new) should find a route")
	}
	if route.Prefix != "/api/v1/bookings" {
		t.Errorf("route.Prefix = %q, want %q", route.Prefix, "/api/v1/bookings")
	}
	if !route.RequireAuth {
		t.Error("bookings route should require auth")
	}
}

func TestResolve_AuthRouteDoesNotRequireAuth(t *testing.T) {
	uc := New(buildTestConfig("http://localhost:9999"))

	route, ok := uc.Resolve("/api/v1/auth/login")
	if !ok {
		t.Fatal("Resolve(/api/v1/auth/login) should find a route")
	}
	if route.RequireAuth {
		t.Error("auth route should not require auth")
	}
}

func TestResolve_FirstMatchWins(t *testing.T) {
	uc := New(buildTestConfig("http://localhost:9999"))

	// /api/v1/auth is registered before /api/v1/users
	route, ok := uc.Resolve("/api/v1/auth/signup")
	if !ok {
		t.Fatal("should resolve /api/v1/auth/signup")
	}
	if route.Prefix != "/api/v1/auth" {
		t.Errorf("expected /api/v1/auth route, got %q", route.Prefix)
	}
}

// ---- NewWithRules tests ----

func TestNewWithRules_BuildsRoutes(t *testing.T) {
	cfg := buildTestConfig("http://svc:8080")
	rules := []config.RouteRule{
		{Prefix: "/api/v1/auth", Service: "user", Methods: []string{"POST"}, RequireAuth: false, Version: "v1"},
		{Prefix: "/api/v1/bookings", Service: "booking", Methods: []string{"*"}, RequireAuth: true, Version: "v1"},
	}

	uc := NewWithRules(cfg, rules)
	if uc == nil {
		t.Fatal("NewWithRules() returned nil")
	}

	route, ok := uc.Resolve("/api/v1/auth/login")
	if !ok {
		t.Error("NewWithRules() Resolve(/api/v1/auth/login) should match")
	}
	if route.Prefix != "/api/v1/auth" {
		t.Errorf("route.Prefix = %q, want /api/v1/auth", route.Prefix)
	}

	_, ok = uc.Resolve("/api/v1/unknown")
	if ok {
		t.Error("NewWithRules() Resolve(/api/v1/unknown) should not match")
	}
}

func TestNewWithRules_MapsServiceURLs(t *testing.T) {
	cfg := config.Config{
		Services: config.ServicesConfig{
			User:    "http://user-service:8081",
			Booking: "http://booking-service:8083",
		},
		CB:    config.CircuitBreakerConfig{MaxRequests: 5, FailureRatio: 0.5, TimeoutSecs: 5},
		Retry: config.RetryConfig{MaxAttempts: 1},
	}

	rules := []config.RouteRule{
		{Prefix: "/api/v1/users", Service: "user", Methods: []string{"*"}, RequireAuth: true},
	}

	uc := NewWithRules(cfg, rules)
	route, ok := uc.Resolve("/api/v1/users/profile")
	if !ok {
		t.Fatal("Resolve should find users route")
	}
	if route.ServiceURL != "http://user-service:8081" {
		t.Errorf("ServiceURL = %q, want http://user-service:8081", route.ServiceURL)
	}
}

func TestNewWithRules_EmptyRules(t *testing.T) {
	cfg := buildTestConfig("http://svc:8080")
	uc := NewWithRules(cfg, []config.RouteRule{})
	if uc == nil {
		t.Fatal("NewWithRules() with empty rules returned nil")
	}
	_, ok := uc.Resolve("/api/v1/anything")
	if ok {
		t.Error("NewWithRules() with empty rules should not resolve any path")
	}
}

// ---- ValidatedClaimsKey tests ----

func TestValidatedClaimsKey_NotNil(t *testing.T) {
	key := ValidatedClaimsKey()
	if key == nil {
		t.Error("ValidatedClaimsKey() should not be nil")
	}
}

func TestValidatedClaimsKey_Consistent(t *testing.T) {
	k1 := ValidatedClaimsKey()
	k2 := ValidatedClaimsKey()
	if k1 != k2 {
		t.Error("ValidatedClaimsKey() should return consistent comparable values")
	}
}

func TestValidatedClaimsKey_UsableAsContextKey(t *testing.T) {
	claims := &entity.Claims{UserID: "u1", Roles: []string{"admin"}}
	ctx := context.WithValue(context.Background(), ValidatedClaimsKey(), claims)

	got, ok := ctx.Value(ValidatedClaimsKey()).(*entity.Claims)
	if !ok || got == nil {
		t.Fatal("context value should be retrievable using ValidatedClaimsKey()")
	}
	if got.UserID != "u1" {
		t.Errorf("retrieved claims UserID = %q, want %q", got.UserID, "u1")
	}
}

// ---- serviceNameFromURL tests ----

func TestServiceNameFromURL(t *testing.T) {
	tests := []struct {
		rawURL string
		want   string
	}{
		{"http://user-service:8082", "user-service"},
		{"http://localhost:9999", "localhost"},
		{"http://booking-service:8083/api", "booking-service"},
		{"", ""},
	}

	for _, tt := range tests {
		got := serviceNameFromURL(tt.rawURL)
		if got != tt.want {
			t.Errorf("serviceNameFromURL(%q) = %q, want %q", tt.rawURL, got, tt.want)
		}
	}
}

// ---- jitter tests ----

func TestJitter_InRange(t *testing.T) {
	for i := 0; i < 1000; i++ {
		d := jitter(100, 500)
		if d < 100*time.Millisecond || d >= 500*time.Millisecond {
			t.Errorf("jitter(100, 500) = %v, out of range [100ms, 500ms)", d)
		}
	}
}

func TestJitter_MinEqualsMax(t *testing.T) {
	d := jitter(200, 200)
	if d != 200*time.Millisecond {
		t.Errorf("jitter(200, 200) = %v, want 200ms", d)
	}
}

func TestJitter_MaxLessThanMin(t *testing.T) {
	d := jitter(500, 100)
	if d != 500*time.Millisecond {
		t.Errorf("jitter(500, 100) = %v, want 500ms (min when max < min)", d)
	}
}

func TestJitter_ZeroMin(t *testing.T) {
	for i := 0; i < 100; i++ {
		d := jitter(0, 100)
		if d < 0 || d >= 100*time.Millisecond {
			t.Errorf("jitter(0, 100) = %v, want in [0, 100ms)", d)
		}
	}
}

// ---- responseRecorder tests ----

func TestResponseRecorder_DefaultStatusCode(t *testing.T) {
	rec := newResponseRecorder()
	if rec.statusCode != http.StatusOK {
		t.Errorf("newResponseRecorder() statusCode = %d, want %d", rec.statusCode, http.StatusOK)
	}
}

func TestResponseRecorder_WriteHeader(t *testing.T) {
	rec := newResponseRecorder()
	rec.WriteHeader(http.StatusCreated)
	if rec.statusCode != http.StatusCreated {
		t.Errorf("WriteHeader(201) statusCode = %d, want %d", rec.statusCode, http.StatusCreated)
	}
}

func TestResponseRecorder_WriteHeaderOnce(t *testing.T) {
	rec := newResponseRecorder()
	rec.WriteHeader(http.StatusCreated)
	rec.WriteHeader(http.StatusOK) // second call should be ignored
	if rec.statusCode != http.StatusCreated {
		t.Errorf("second WriteHeader should be ignored, got %d", rec.statusCode)
	}
}

func TestResponseRecorder_Write(t *testing.T) {
	rec := newResponseRecorder()
	n, err := rec.Write([]byte("hello"))
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	if n != 5 {
		t.Errorf("Write() n = %d, want 5", n)
	}
	if rec.body.String() != "hello" {
		t.Errorf("Write() body = %q, want %q", rec.body.String(), "hello")
	}
}

func TestResponseRecorder_Write_SetsDefaultStatus(t *testing.T) {
	rec := newResponseRecorder()
	rec.Write([]byte("data"))
	if rec.statusCode != http.StatusOK {
		t.Errorf("Write() without prior WriteHeader should default to 200, got %d", rec.statusCode)
	}
}

func TestResponseRecorder_Header(t *testing.T) {
	rec := newResponseRecorder()
	h := rec.Header()
	if h == nil {
		t.Fatal("Header() should not return nil")
	}
	h.Set("X-Test", "value")
	if rec.Header().Get("X-Test") != "value" {
		t.Error("Header() should return mutable header map")
	}
}

func TestResponseRecorder_FlushTo(t *testing.T) {
	rec := newResponseRecorder()
	rec.Header().Set("X-Custom", "yes")
	rec.WriteHeader(http.StatusAccepted)
	rec.Write([]byte(`{"ok":true}`))

	w := httptest.NewRecorder()
	if err := rec.FlushTo(w); err != nil {
		t.Fatalf("FlushTo() error = %v", err)
	}

	if w.Code != http.StatusAccepted {
		t.Errorf("FlushTo() status = %d, want %d", w.Code, http.StatusAccepted)
	}
	if w.Header().Get("X-Custom") != "yes" {
		t.Error("FlushTo() should copy headers")
	}
	if w.Body.String() != `{"ok":true}` {
		t.Errorf("FlushTo() body = %q, want %q", w.Body.String(), `{"ok":true}`)
	}
}

func TestResponseRecorder_FlushTo_EmptyBody(t *testing.T) {
	rec := newResponseRecorder()
	rec.WriteHeader(http.StatusNoContent)

	w := httptest.NewRecorder()
	if err := rec.FlushTo(w); err != nil {
		t.Fatalf("FlushTo() error = %v", err)
	}

	if w.Code != http.StatusNoContent {
		t.Errorf("FlushTo() status = %d, want %d", w.Code, http.StatusNoContent)
	}
	if w.Body.Len() != 0 {
		t.Error("FlushTo() with empty body should write no bytes")
	}
}

// ---- buildDirector tests ----

func TestBuildDirector_SetsTargetURLSchemeAndHost(t *testing.T) {
	target, _ := url.Parse("http://backend-service:8080")
	director := buildDirector(target)

	req, _ := http.NewRequest(http.MethodGet, "http://gateway/api/v1/users/1", nil)
	director(req)

	if req.URL.Scheme != "http" {
		t.Errorf("director scheme = %q, want http", req.URL.Scheme)
	}
	if req.URL.Host != "backend-service:8080" {
		t.Errorf("director host = %q, want backend-service:8080", req.URL.Host)
	}
	if req.Host != "backend-service:8080" {
		t.Errorf("director req.Host = %q, want backend-service:8080", req.Host)
	}
}

func TestBuildDirector_RemovesDownstreamHeaders(t *testing.T) {
	target, _ := url.Parse("http://backend:8080")
	director := buildDirector(target)

	req, _ := http.NewRequest(http.MethodGet, "http://gateway/api/v1/test", nil)
	req.Header.Set("X-User-ID", "should-be-removed")
	req.Header.Set("X-User-Roles", "admin")
	req.Header.Set("X-API-Key", "secret")
	req.Header.Set("X-Internal-Token", "internal")
	director(req)

	for _, h := range downstreamHeaders {
		if req.Header.Get(h) != "" {
			t.Errorf("director should remove header %q, but it's still set", h)
		}
	}
}

func TestBuildDirector_InjectsClaims(t *testing.T) {
	target, _ := url.Parse("http://backend:8080")
	director := buildDirector(target)

	claims := &entity.Claims{
		UserID: "user-99",
		Roles:  []string{"landlord", "admin"},
		APIKey: "my-api-key",
	}
	ctx := context.WithValue(context.Background(), validatedClaimsKey{}, claims)
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "http://gateway/api/v1/test", nil)
	director(req)

	if req.Header.Get("X-User-ID") != "user-99" {
		t.Errorf("director X-User-ID = %q, want %q", req.Header.Get("X-User-ID"), "user-99")
	}
	if req.Header.Get("X-User-Roles") != "landlord,admin" {
		t.Errorf("director X-User-Roles = %q, want landlord,admin", req.Header.Get("X-User-Roles"))
	}
	if req.Header.Get("X-API-Key") != "my-api-key" {
		t.Errorf("director X-API-Key = %q, want my-api-key", req.Header.Get("X-API-Key"))
	}
}

func TestBuildDirector_InjectsClaims_NoAPIKey(t *testing.T) {
	target, _ := url.Parse("http://backend:8080")
	director := buildDirector(target)

	claims := &entity.Claims{
		UserID: "u1",
		Roles:  []string{"user"},
		APIKey: "", // no API key
	}
	ctx := context.WithValue(context.Background(), validatedClaimsKey{}, claims)
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "http://gateway/", nil)
	director(req)

	// X-API-Key should NOT be set when APIKey is empty
	if req.Header.Get("X-API-Key") != "" {
		t.Error("director should not inject X-API-Key when claims.APIKey is empty")
	}
}

func TestBuildDirector_NoClaims_NoInjection(t *testing.T) {
	target, _ := url.Parse("http://backend:8080")
	director := buildDirector(target)

	req, _ := http.NewRequest(http.MethodGet, "http://gateway/api/v1/test", nil)
	director(req)

	if req.Header.Get("X-User-ID") != "" {
		t.Error("director should not inject X-User-ID when no claims in context")
	}
}

func TestBuildDirector_PreservesRequestID(t *testing.T) {
	target, _ := url.Parse("http://backend:8080")
	director := buildDirector(target)

	req, _ := http.NewRequest(http.MethodGet, "http://gateway/api/v1/test", nil)
	req.Header.Set("X-Request-ID", "req-abc-123")
	director(req)

	if req.Header.Get("X-Request-ID") != "req-abc-123" {
		t.Errorf("director X-Request-ID = %q, want req-abc-123", req.Header.Get("X-Request-ID"))
	}
}

// ---- Forward integration test with test server ----

func TestForward_SuccessfulProxy(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"result": "ok"})
	}))
	defer upstream.Close()

	cfg := config.Config{
		Services: config.ServicesConfig{
			Auth: upstream.URL,
			User: upstream.URL,
		},
		CB:    config.CircuitBreakerConfig{MaxRequests: 5, FailureRatio: 0.6, TimeoutSecs: 5},
		Retry: config.RetryConfig{MaxAttempts: 1},
	}

	uc := New(cfg)
	route, ok := uc.Resolve("/api/v1/auth/login")
	if !ok {
		t.Fatal("Resolve should find auth route")
	}

	req, _ := http.NewRequest(http.MethodPost, upstream.URL+"/api/v1/auth/login",
		bytes.NewBufferString(`{"email":"test@test.com"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	err := uc.Forward(context.Background(), route, w, req)
	if err != nil {
		t.Fatalf("Forward() error = %v", err)
	}
	if w.Code != http.StatusOK {
		t.Errorf("Forward() status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestForward_NoProxy_Error(t *testing.T) {
	cfg := buildTestConfig("http://localhost:9999")
	uc := New(cfg)

	// Route with a service URL that wasn't registered in proxies
	route := &entity.Route{ServiceURL: "http://not-registered:1234"}
	req, _ := http.NewRequest(http.MethodGet, "http://gateway/api/v1/test", nil)
	w := httptest.NewRecorder()

	err := uc.Forward(context.Background(), route, w, req)
	if err == nil {
		t.Error("Forward() should return error when no proxy registered for service URL")
	}
}

func TestForward_Upstream5xx_ReturnsError(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer upstream.Close()

	cfg := config.Config{
		Services: config.ServicesConfig{
			Auth: upstream.URL,
			User: upstream.URL,
		},
		CB:    config.CircuitBreakerConfig{MaxRequests: 5, FailureRatio: 0.6, TimeoutSecs: 5},
		Retry: config.RetryConfig{MaxAttempts: 1, WaitMinMS: 10, WaitMaxMS: 50},
	}

	uc := New(cfg)
	route, _ := uc.Resolve("/api/v1/auth/login")

	req, _ := http.NewRequest(http.MethodGet, upstream.URL+"/api/v1/auth/login", nil)
	w := httptest.NewRecorder()

	// With a 5xx upstream, Forward should return an error
	err := uc.Forward(context.Background(), route, w, req)
	if err == nil {
		t.Error("Forward() should return error when upstream returns 5xx")
	}
}
