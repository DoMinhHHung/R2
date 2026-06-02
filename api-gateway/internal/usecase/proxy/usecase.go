package proxy

import (
	"context"
	"fmt"
	"math/rand/v2"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/DoMinhHHung/Rental/internal/domain/entity"
	"github.com/DoMinhHHung/Rental/internal/infrastructure/config"
	"github.com/sony/gobreaker"
)

var downstreamHeaders = []string{
	"X-User-ID",
	"X-User-Roles",
	"X-API-Key",
	"X-Internal-Token",
}

type ServiceProxy struct {
	target  *url.URL
	proxy   *httputil.ReverseProxy
	breaker *gobreaker.CircuitBreaker
}

type UseCase struct {
	services map[string]*ServiceProxy
	routes   []entity.Route
	cfg      config.Config
}

func New(cfg config.Config) *UseCase {
	uc := &UseCase{
		cfg:      cfg,
		services: make(map[string]*ServiceProxy),
	}

	uc.registerRoutes()
	uc.initProxies()
	return uc
}

func (u *UseCase) registerRoutes() {
	u.routes = []entity.Route{
		{Prefix: "/api/v1/auth", ServiceURL: u.cfg.Services.User, Methods: []string{"POST"}, RequireAuth: false, Roles: []string{}, Version: "v1"},
		{Prefix: "/api/v1/users", ServiceURL: u.cfg.Services.User, Methods: []string{"*"}, RequireAuth: true, Roles: []string{"admin", "user"}, Version: "v1"},
		{Prefix: "/api/v1/properties", ServiceURL: u.cfg.Services.Property, Methods: []string{"*"}, RequireAuth: false, Roles: []string{}, Version: "v1"},
		{Prefix: "/api/v1/bookings", ServiceURL: u.cfg.Services.Booking, Methods: []string{"*"}, RequireAuth: true, Roles: []string{"user", "landlord", "admin"}, Version: "v1"},
		{Prefix: "/api/v1/payments", ServiceURL: u.cfg.Services.Payment, Methods: []string{"*"}, RequireAuth: true, Roles: []string{"user", "landlord", "admin"}, Version: "v1"},
		{Prefix: "/api/v1/notifications", ServiceURL: u.cfg.Services.Notification, Methods: []string{"*"}, RequireAuth: true, Roles: []string{"user", "landlord", "admin"}, Version: "v1"},
	}
}

func (u *UseCase) initProxies() {

	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   5 * time.Second,
		ResponseHeaderTimeout: 30 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	seen := map[string]bool{}
	for _, route := range u.routes {
		if seen[route.ServiceURL] {
			continue
		}
		seen[route.ServiceURL] = true

		target, err := url.Parse(route.ServiceURL)
		if err != nil {
			continue
		}

		svcName := serviceNameFromURL(route.ServiceURL)
		cbSettings := gobreaker.Settings{
			Name:        svcName,
			MaxRequests: u.cfg.CB.MaxRequests,
			Interval:    time.Duration(u.cfg.CB.IntervalSecs) * time.Second,
			Timeout:     time.Duration(u.cfg.CB.TimeoutSecs) * time.Second,
			ReadyToTrip: func(counts gobreaker.Counts) bool {
				if counts.Requests < 3 {
					return false
				}
				ratio := float64(counts.TotalFailures) / float64(counts.Requests)
				return ratio >= u.cfg.CB.FailureRatio
			},
		}

		proxy := httputil.NewSingleHostReverseProxy(target)
		proxy.Transport = transport
		proxy.Director = buildDirector(target)
		proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
			http.Error(w, `{"error":"upstream_unavailable"}`, http.StatusBadGateway)
		}

		u.services[route.ServiceURL] = &ServiceProxy{
			target:  target,
			proxy:   proxy,
			breaker: gobreaker.NewCircuitBreaker(cbSettings),
		}
	}
}

func (u *UseCase) Resolve(path string) (*entity.Route, bool) {
	for i := range u.routes {
		if strings.HasPrefix(path, u.routes[i].Prefix) {
			return &u.routes[i], true
		}
	}
	return nil, false
}

func NewWithRules(cfg config.Config, routeRules []config.RouteRule) *UseCase {
	serviceURLFor := map[string]string{
		"user":         cfg.Services.User,
		"property":     cfg.Services.Property,
		"booking":      cfg.Services.Booking,
		"payment":      cfg.Services.Payment,
		"notification": cfg.Services.Notification,
	}

	routes := make([]entity.Route, 0, len(routeRules))
	for _, r := range routeRules {
		routes = append(routes, entity.Route{
			Prefix:      r.Prefix,
			ServiceURL:  serviceURLFor[r.Service],
			Methods:     r.Methods,
			RequireAuth: r.RequireAuth,
			Version:     r.Version,
		})
	}

	uc := &UseCase{cfg: cfg, services: make(map[string]*ServiceProxy), routes: routes}
	uc.initProxies()
	return uc
}

func (u *UseCase) Forward(ctx context.Context, route *entity.Route, w http.ResponseWriter, r *http.Request) error {
	sp, ok := u.services[route.ServiceURL]
	if !ok {
		return fmt.Errorf("no proxy for %s", route.ServiceURL)
	}

	var lastErr error
	maxAttempts := u.cfg.Retry.MaxAttempts
	if maxAttempts < 1 {
		maxAttempts = 1
	}

	for attempt := 0; attempt < maxAttempts; attempt++ {
		if attempt > 0 {
			wait := jitter(u.cfg.Retry.WaitMinMS, u.cfg.Retry.WaitMaxMS)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(wait):
			}
		}

		rec := &responseRecorder{ResponseWriter: w, statusCode: 200}

		_, err := sp.breaker.Execute(func() (interface{}, error) {
			sp.proxy.ServeHTTP(rec, r)
			if rec.statusCode >= 500 {
				return nil, fmt.Errorf("upstream error: %d", rec.statusCode)
			}
			return nil, nil
		})

		if err == nil {
			return nil
		}

		if err == gobreaker.ErrOpenState {
			http.Error(w, `{"error":"circuit_open"}`, http.StatusServiceUnavailable)
			return err
		}

		lastErr = err
	}

	return lastErr
}

func buildDirector(target *url.URL) func(*http.Request) {
	return func(req *http.Request) {
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.Host = target.Host

		for _, h := range downstreamHeaders {
			req.Header.Del(h)
		}

		if claims, ok := req.Context().Value(validatedClaimsKey{}).(*entity.Claims); ok && claims != nil {
			req.Header.Set("X-User-ID", claims.UserID)
			req.Header.Set("X-User-Roles", strings.Join(claims.Roles, ","))
			if claims.APIKey != "" {
				req.Header.Set("X-API-Key", claims.APIKey)
			}
		}

		if rid := req.Header.Get("X-Request-ID"); rid != "" {
			req.Header.Set("X-Request-ID", rid)
		}
	}
}

type validatedClaimsKey struct{}

func ValidatedClaimsKey() interface{} { return validatedClaimsKey{} }

func serviceNameFromURL(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	return u.Hostname()
}

func jitter(minMS, maxMS int) time.Duration {
	if maxMS <= minMS {
		return time.Duration(minMS) * time.Millisecond
	}
	n := minMS + rand.IntN(maxMS-minMS)
	return time.Duration(n) * time.Millisecond
}

type responseRecorder struct {
	http.ResponseWriter
	statusCode int
	once       sync.Once
}

func (r *responseRecorder) WriteHeader(code int) {
	r.once.Do(func() {
		r.statusCode = code
		r.ResponseWriter.WriteHeader(code)
	})
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	r.once.Do(func() {
		r.statusCode = http.StatusOK
	})
	return r.ResponseWriter.Write(b)
}
