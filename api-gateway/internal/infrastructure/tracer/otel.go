package tracer

import (
	"context"
	"fmt"

	"github.com/DoMinhHHung/R2/internal/domain/port"
	"github.com/DoMinhHHung/R2/internal/infrastructure/config"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type OtelTracer struct {
	tracer   trace.Tracer
	provider *sdktrace.TracerProvider
}

type otelSpan struct {
	span trace.Span
}

func New(cfg config.OTELConfig) (*OtelTracer, error) {
	conn, err := grpc.NewClient(cfg.Endpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("grpc dial otel: %w", err)
	}

	exporter, err := otlptracegrpc.New(context.Background(),
		otlptracegrpc.WithGRPCConn(conn),
	)
	if err != nil {
		return nil, fmt.Errorf("otlp exporter: %w", err)
	}

	res, err := resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(cfg.ServiceName),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("otel resource: %w", err)
	}

	provider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)

	otel.SetTracerProvider(provider)

	return &OtelTracer{
		tracer:   provider.Tracer(cfg.ServiceName),
		provider: provider,
	}, nil
}

func (t *OtelTracer) Start(ctx context.Context, spanName string) (context.Context, port.Span) {
	ctx, span := t.tracer.Start(ctx, spanName)
	return ctx, &otelSpan{span: span}
}

func (t *OtelTracer) Shutdown(ctx context.Context) error {
	return t.provider.Shutdown(ctx)
}

func (s *otelSpan) End() {
	s.span.End()
}

func (s *otelSpan) SetAttribute(key string, value any) {
	switch v := value.(type) {
	case string:
		s.span.SetAttributes(attribute.String(key, v))
	case int:
		s.span.SetAttributes(attribute.Int(key, v))
	case bool:
		s.span.SetAttributes(attribute.Bool(key, v))
	case float64:
		s.span.SetAttributes(attribute.Float64(key, v))
	}
}

func (s *otelSpan) RecordError(err error) {
	s.span.RecordError(err)
}
