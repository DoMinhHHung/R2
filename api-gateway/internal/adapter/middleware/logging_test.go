package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DoMinhHHung/R2/internal/adapter/middleware"
	"github.com/DoMinhHHung/R2/internal/domain/port"
	"github.com/gin-gonic/gin"
)

type logEntry struct {
	msg    string
	fields []any
}

type spyLogger struct {
	infoCalls []logEntry
}

func (s *spyLogger) Info(msg string, fields ...any) {
	s.infoCalls = append(s.infoCalls, logEntry{msg: msg, fields: fields})
}
func (s *spyLogger) Error(_ string, _ ...any) {}
func (s *spyLogger) Warn(_ string, _ ...any)  {}
func (s *spyLogger) Debug(_ string, _ ...any) {}
func (s *spyLogger) With(_ ...any) port.Logger { return s }

func TestLogging_CallsLoggerInfo(t *testing.T) {
	logger := &spyLogger{}

	r := gin.New()
	r.GET("/hello", middleware.Logging(logger), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/hello", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if len(logger.infoCalls) == 0 {
		t.Fatal("Logging() should call logger.Info at least once")
	}

	call := logger.infoCalls[0]
	if call.msg != "request" {
		t.Errorf("Logging() log message = %q, want %q", call.msg, "request")
	}
}

func TestLogging_LogsHTTPMethod(t *testing.T) {
	logger := &spyLogger{}

	r := gin.New()
	r.POST("/submit", middleware.Logging(logger), func(c *gin.Context) {
		c.Status(http.StatusCreated)
	})

	req := httptest.NewRequest(http.MethodPost, "/submit", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if len(logger.infoCalls) == 0 {
		t.Fatal("Logging() should have called logger.Info")
	}

	fields := logger.infoCalls[0].fields
	found := false
	for i := 0; i+1 < len(fields); i += 2 {
		if fields[i] == "method" && fields[i+1] == http.MethodPost {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Logging() fields %v should contain method=POST", fields)
	}
}

func TestLogging_LogsStatusCode(t *testing.T) {
	logger := &spyLogger{}

	r := gin.New()
	r.GET("/notfound", middleware.Logging(logger), func(c *gin.Context) {
		c.Status(http.StatusNotFound)
	})

	req := httptest.NewRequest(http.MethodGet, "/notfound", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	fields := logger.infoCalls[0].fields
	found := false
	for i := 0; i+1 < len(fields); i += 2 {
		if fields[i] == "status" && fields[i+1] == http.StatusNotFound {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Logging() fields %v should contain status=404", fields)
	}
}

func TestLogging_LogsPath(t *testing.T) {
	logger := &spyLogger{}

	r := gin.New()
	r.GET("/api/v1/test", middleware.Logging(logger), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	fields := logger.infoCalls[0].fields
	found := false
	for i := 0; i+1 < len(fields); i += 2 {
		if fields[i] == "path" && fields[i+1] == "/api/v1/test" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Logging() fields %v should contain path=/api/v1/test", fields)
	}
}

func TestLogging_LogsLatency(t *testing.T) {
	logger := &spyLogger{}

	r := gin.New()
	r.GET("/fast", middleware.Logging(logger), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/fast", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	fields := logger.infoCalls[0].fields
	found := false
	for i := 0; i < len(fields); i++ {
		if fields[i] == "latency_ms" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Logging() fields %v should contain latency_ms", fields)
	}
}

func TestLogging_LogsUserAgent(t *testing.T) {
	logger := &spyLogger{}

	r := gin.New()
	r.GET("/test", middleware.Logging(logger), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("User-Agent", "test-agent/1.0")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	fields := logger.infoCalls[0].fields
	found := false
	for i := 0; i+1 < len(fields); i += 2 {
		if fields[i] == "user_agent" && fields[i+1] == "test-agent/1.0" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Logging() fields %v should contain user_agent=test-agent/1.0", fields)
	}
}

func TestLogging_LogsRequestID(t *testing.T) {
	logger := &spyLogger{}

	r := gin.New()
	r.GET("/test",
		middleware.RequestID(),
		middleware.Logging(logger),
		func(c *gin.Context) {
			c.Status(http.StatusOK)
		},
	)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set(middleware.HeaderRequestID, "test-req-id-999")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	fields := logger.infoCalls[0].fields
	found := false
	for i := 0; i+1 < len(fields); i += 2 {
		if fields[i] == "request_id" && fields[i+1] == "test-req-id-999" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Logging() fields %v should contain request_id=test-req-id-999", fields)
	}
}