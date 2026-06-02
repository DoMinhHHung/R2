package port

import "context"

type Tracer interface {
	Start(ctx context.Context, spanName string) (context.Context, Span)
}

type Span interface {
	End()
	SetAttribute(key string, value any)
	RecordError(err error)
}
