package trace

import (
	"context"
	"go.opentelemetry.io/otel"
)

const (
	CtxKeySystemPrefix    = "gin_ctx_key_prefix"
	CtxKeySpanContext     = CtxKeySystemPrefix + "SpanContext"
)

type ContextKey string

const (
	CtxTraceSpanKey ContextKey = "CtxTraceSpan"
	CtxMetadataKey  ContextKey = "CtxMetadata"
)

type Span struct {
	traceId string
	spanId  string
}

func NewSpan(traceId, spanId string) Span {
	return Span{
		traceId: traceId,
		spanId:  spanId,
	}
}

func (t *Span) Trace() string {
	return t.traceId
}

func (t *Span) Span() string {
	return t.spanId
}

func GetContext(ctx context.Context, traceId, spanId string) context.Context {
	ctx = context.WithValue(ctx, string(CtxTraceSpanKey), NewSpan(traceId, spanId))
	return ctx
}

// 设置trace_id
func GenTrace(ctx context.Context, name string) context.Context {
	tracer := otel.GetTracerProvider().Tracer(name)
	_, span := tracer.Start(
		ctx,
		name,
	)

	defer span.End()
	if sc := span.SpanContext(); sc.HasTraceID() {
		return GetContext(ctx, sc.TraceID().String(), sc.SpanID().String())
	}
	return ctx
}

func ExtraTraceID(ctx context.Context) string {
	var traceId string
	span, ok := ctx.Value(string(CtxTraceSpanKey)).(Span)
	if ok {
		traceId = span.Trace()
	}
	return traceId
}

