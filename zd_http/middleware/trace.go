package middleware

import (
	"github.com/lfxnxf/emo-frame/trace"
	"github.com/lfxnxf/emo-frame/zd_http/http_ctx"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	oteltrace "go.opentelemetry.io/otel/trace"
	"net/http"
)

var HeaderTraceIdKey = http.CanonicalHeaderKey("x-trace-id")
var HeaderSpanIdKey = http.CanonicalHeaderKey("x-span-id")

// todo 拆分微服务时再接入open-trace，单体时模拟生成trace_id用来查询日志
func setTrace() http_ctx.HttpHandler {
	return func(c *http_ctx.HttpContext) {
		propagator := otel.GetTextMapPropagator()
		ctx := propagator.Extract(c, propagation.HeaderCarrier(c.Request.Header))

		tracer := otel.GetTracerProvider().Tracer(c.FullPath())
		_, span := tracer.Start(
			ctx,
			c.Request.URL.Path,
			oteltrace.WithSpanKind(oteltrace.SpanKindServer),
			oteltrace.WithAttributes(semconv.HTTPServerAttributesFromHTTPRequest(
				"", c.Request.URL.Path, c.Request)...),
		)

		defer span.End()
		// 设置span到ctx
		c.Set(trace.CtxKeySpanContext, span)

		var (
			traceId string
			spanId  string
		)
		if sc := span.SpanContext(); sc.HasTraceID() {
			traceId = sc.TraceID().String()
			spanId = sc.SpanID().String()
			c.Set(string(trace.CtxTraceSpanKey), trace.NewSpan(traceId, spanId))
		}
		c.Writer.Header().Set(HeaderTraceIdKey, traceId)
		c.Writer.Header().Set(HeaderSpanIdKey, spanId)
		c.Next()
	}
}
