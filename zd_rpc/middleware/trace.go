package middleware

import (
	"context"
	"github.com/lfxnxf/emo-frame/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

var MdTraceIdKey = "x-trace-id"
var MdSpanIdKey = "x-span-id"

// todo 拆分微服务时再接入open-trace，单体时模拟生成trace_id用来查询日志
func getTrace(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	var traceId string
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		traceList := md.Get(MdTraceIdKey)
		if len(traceList) > 0 {
			traceId = traceList[0]
		}
	}
	if len(traceId) <= 0 {
		ctx = trace.GenTrace(ctx, info.FullMethod)
	}
	return handler(ctx, req)
}

func setTrace() grpc.DialOption {
	return grpc.WithUnaryInterceptor(func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		span, ok := ctx.Value(string(trace.CtxTraceSpanKey)).(trace.Span)
		if !ok || len(span.Trace()) <= 0 {
			ctx = trace.GenTrace(ctx, method)
			span, _ = ctx.Value(string(trace.CtxTraceSpanKey)).(trace.Span)
		}
		md := metadata.Pairs(MdTraceIdKey, span.Trace(), MdSpanIdKey, span.Span())
		ctx = metadata.NewOutgoingContext(ctx, md)
		return invoker(ctx, method, req, reply, cc, opts...)
	})
}
