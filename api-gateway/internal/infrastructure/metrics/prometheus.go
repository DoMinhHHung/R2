package metrics

import (
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Metrics struct {
	requestTotal    *prometheus.CounterVec
	requestDuration *prometheus.HistogramVec
	requestInFlight *prometheus.GaugeVec
	upstreamErrors  *prometheus.CounterVec
	cbStateChanges  *prometheus.CounterVec
}

func New(serviceName string) *Metrics {
	return &Metrics{
		requestTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Namespace: "gateway",
			Name:      "requests_total",
			Help:      "Total number of HTTP requests",
		}, []string{"method", "path", "status", "service"}),

		requestDuration: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: "gateway",
			Name:      "request_duration_seconds",
			Help:      "HTTP request duration in seconds",
			Buckets:   prometheus.DefBuckets,
		}, []string{"method", "path", "service"}),

		requestInFlight: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "gateway",
			Name:      "requests_in_flight",
			Help:      "Number of requests currently being processed",
		}, []string{"service"}),

		upstreamErrors: promauto.NewCounterVec(prometheus.CounterOpts{
			Namespace: "gateway",
			Name:      "upstream_errors_total",
			Help:      "Total upstream errors",
		}, []string{"service", "reason"}),

		cbStateChanges: promauto.NewCounterVec(prometheus.CounterOpts{
			Namespace: "gateway",
			Name:      "circuit_breaker_state_changes_total",
			Help:      "Circuit breaker state changes",
		}, []string{"service", "state"}),
	}
}

func (m *Metrics) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		svc := resolveServiceLabel(c.Request.URL.Path)

		m.requestInFlight.WithLabelValues(svc).Inc()
		defer m.requestInFlight.WithLabelValues(svc).Dec()

		c.Next()

		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Writer.Status())

		m.requestTotal.WithLabelValues(
			c.Request.Method,
			c.FullPath(),
			status,
			svc,
		).Inc()

		m.requestDuration.WithLabelValues(
			c.Request.Method,
			c.FullPath(),
			svc,
		).Observe(duration)
	}
}

func (m *Metrics) RecordUpstreamError(service, reason string) {
	m.upstreamErrors.WithLabelValues(service, reason).Inc()
}

func (m *Metrics) RecordCBStateChange(service, state string) {
	m.cbStateChanges.WithLabelValues(service, state).Inc()
}

func (m *Metrics) Handler() gin.HandlerFunc {
	h := promhttp.Handler()
	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

func resolveServiceLabel(path string) string {
	parts := strings.Split(strings.TrimPrefix(path, "/"), "/")
	if len(parts) >= 3 {
		return parts[2]
	}
	return "unknown"
}
