package middleware

import (
	"fmt"
	"github.com/lfxnxf/emo-frame/logging"
	"github.com/lfxnxf/emo-frame/zd_http/http_ctx"
	"go.uber.org/zap"
	"runtime"
)

func recoverSysMW() http_ctx.HttpHandler {
	return func(c *http_ctx.HttpContext) {
		defer func() {
			if r := recover(); r != nil {
				buf := make([]byte, 64<<10)
				buf = buf[:runtime.Stack(buf, false)]
				err := fmt.Errorf("errgroup: panic recovered: %s\n%s", r, buf)
				logging.Errorw("mw_sys_recover_happen", zap.Error(err))
				logging.CrashLog(err)
			}
		}()
		c.Next()
	}
}
