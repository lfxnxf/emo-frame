package middleware

import (
	"context"
	"encoding/json"
	"github.com/lfxnxf/emo-frame/logging"
	"github.com/lfxnxf/emo-frame/trace"
	"github.com/lfxnxf/emo-frame/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"math"
	"os"
	"time"
)

func loggingAccess(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	// 当前时间
	nowTime := time.Now()

	// 请求方法
	resp, err := handler(ctx, req)

	// 获取trace_id
	var traceId string
	span, ok := ctx.Value(string(trace.CtxTraceSpanKey)).(trace.Span)
	if ok {
		traceId = span.Trace()
	}
	// 服务名称
	hostname, _ := os.Hostname()

	var realIp string
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		rips := md.Get("x-real-ip")
		if len(rips) > 0 {
			realIp = rips[0]
		}
	}

	requestBody, _ := json.Marshal(req)

	respBody, _ := json.Marshal(resp)

	var rpcCode = 200
	if err != nil {
		rpcCode = 500
	}

	logItems := []interface{}{
		"start", nowTime.Format(utils.TimeFormatYYYYMMDDHHmmSS),
		"cost", math.Ceil(float64(time.Since(nowTime).Nanoseconds()) / 1e6),
		"trace_id", traceId,
		"host_ip", os.Getenv("FINAL_HOST"),
		"host_name", hostname,
		"req_method", info.FullMethod,
		"real_ip", realIp,
		"rpc_code", rpcCode,
		"req_body", string(requestBody),
		"resp_body", string(respBody),
	}
	logging.DefaultKit.A().Debugw("rpc_server", logItems...)
	return resp, err
}
