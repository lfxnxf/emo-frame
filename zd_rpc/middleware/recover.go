package middleware

import (
	"context"
	"fmt"
	"github.com/lfxnxf/emo-frame/logging"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"runtime"
)

func recoverSysMW(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	defer func() {
		if r := recover(); r != nil {
			buf := make([]byte, 64<<10)
			buf = buf[:runtime.Stack(buf, false)]
			err := fmt.Errorf("errgroup: panic recovered: %s\n%s", r, buf)
			logging.Errorw("mw_sys_recover_happen", zap.Error(err))
			logging.CrashLog(err)
		}
	}()
	return handler(ctx, req)
}
