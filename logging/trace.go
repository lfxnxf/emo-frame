package logging

import (
	"github.com/lfxnxf/emo-frame/trace"
	"golang.org/x/net/context"
)

type contextFunc func(ctx context.Context) (string, string)

var contextList []contextFunc

func RegisterCtx(cb contextFunc) {
	contextList = append(contextList, cb)
}

func extraTraceID(ctx context.Context) string {
	var traceId string
	span, ok := ctx.Value(string(trace.CtxTraceSpanKey)).(trace.Span)
	if ok {
		traceId = span.Trace()
	}
	return traceId
}
