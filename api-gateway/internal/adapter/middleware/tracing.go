package middleware

import (
	"github.com/DoMinhHHung/R2/internal/domain/port"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

func Tracing(tracer port.Tracer) gin.HandlerFunc {
	propagator := otel.GetTextMapPropagator()

	return func(c *gin.Context) {
		ctx := propagator.Extract(c.Request.Context(), propagation.HeaderCarrier(c.Request.Header))

		rid, _ := c.Get("request_id")
		spanName := c.Request.Method + " " + c.FullPath()

		ctx, span := tracer.Start(ctx, spanName)
		defer span.End()

		span.SetAttribute("http.method", c.Request.Method)
		span.SetAttribute("http.path", c.Request.URL.Path)
		span.SetAttribute("http.client_ip", c.ClientIP())
		span.SetAttribute("request_id", rid)

		c.Request = c.Request.WithContext(ctx)
		c.Next()

		span.SetAttribute("http.status_code", c.Writer.Status())

		if c.Writer.Status() >= 500 {
			span.RecordError(nil)
		}

		propagator.Inject(ctx, propagation.HeaderCarrier(c.Writer.Header()))
	}
}
