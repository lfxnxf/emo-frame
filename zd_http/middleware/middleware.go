package middleware

import "github.com/lfxnxf/emo-frame/zd_http/http_ctx"

func GetOpts() []http_ctx.HttpHandler {
	// todo max_connects
	// todo prometheus
	return []http_ctx.HttpHandler{
		loggingAccess(), // 生成access_log
		setTrace(),      // 设置trace
		recoverSysMW(),  // recover
		crossDomain(),   // 跨域设置
	}
}
