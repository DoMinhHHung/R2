package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DoMinhHHung/R2/internal/adapter/middleware"
	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestRequestID_GeneratesIDWhenAbsent(t *testing.T) {
	r := gin.New()
	r.GET("/test", middleware.RequestID(), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	rid := w.Header().Get(middleware.HeaderRequestID)
	if rid == "" {
		t.Error("RequestID() should generate an X-Request-ID when not provided")
	}
}

func TestRequestID_PreservesExistingID(t *testing.T) {
	const existingID = "my-custom-request-id-123"

	r := gin.New()
	r.GET("/test", middleware.RequestID(), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set(middleware.HeaderRequestID, existingID)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	rid := w.Header().Get(middleware.HeaderRequestID)
	if rid != existingID {
		t.Errorf("RequestID() = %q, want %q (existing ID should be preserved)", rid, existingID)
	}
}

func TestRequestID_SetsContextValue(t *testing.T) {
	var ctxRID interface{}

	r := gin.New()
	r.GET("/test", middleware.RequestID(), func(c *gin.Context) {
		ctxRID, _ = c.Get("request_id")
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if ctxRID == nil || ctxRID == "" {
		t.Error("RequestID() should set request_id in gin context")
	}
}

func TestRequestID_ContextMatchesResponseHeader(t *testing.T) {
	var ctxRID interface{}

	r := gin.New()
	r.GET("/test", middleware.RequestID(), func(c *gin.Context) {
		ctxRID, _ = c.Get("request_id")
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	headerRID := w.Header().Get(middleware.HeaderRequestID)
	if ctxRID != headerRID {
		t.Errorf("RequestID() context value %q != header value %q", ctxRID, headerRID)
	}
}

func TestRequestID_GeneratesUniqueIDs(t *testing.T) {
	r := gin.New()
	r.GET("/test", middleware.RequestID(), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	seen := make(map[string]struct{})
	for i := 0; i < 100; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		rid := w.Header().Get(middleware.HeaderRequestID)
		if _, dup := seen[rid]; dup {
			t.Errorf("RequestID() generated duplicate ID: %s", rid)
		}
		seen[rid] = struct{}{}
	}
}

func TestRequestID_HeaderName(t *testing.T) {
	if middleware.HeaderRequestID != "X-Request-ID" {
		t.Errorf("HeaderRequestID = %q, want %q", middleware.HeaderRequestID, "X-Request-ID")
	}
}