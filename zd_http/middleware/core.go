package middleware

import (
	"github.com/lfxnxf/emo-frame/zd_http/http_ctx"
	"net/http"
)

var corsDomains = []string{
	"http://app.cestc-yd.cn:18080",
	// 线上
	"http://39.101.130.183:18080",
	"https://39.101.130.183:18080",
	"http://web.cestc-yd.cn",
	"https://web.cestc-yd.cn",
	"https://webcec.cestc-yd.cn",

	// 测试
	"http://39.101.130.183:18090",
	"https://39.101.130.183:18090",
	"http://testweb.cestc-yd.cn",
	"https://testweb.cestc-yd.cn",
	"https://testwebcec.cestc-yd.cn",

	// 开发环境
	"http://47.92.30.21",
}

// 跨域
func crossDomain() http_ctx.HttpHandler {
	return func(c *http_ctx.HttpContext) {
		method := c.Request.Method
		origin := c.Request.Header.Get("Origin") //请求头部
		//判断origin是否允许跨域请求
		var isExist bool
		for _, val := range corsDomains {
			if val == origin {
				isExist = true
			}
		}
		if isExist {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Headers", "Content-Type,AccessToken,X-CSRF-Token,Authorization,Token,Sign,Sec-WebSocket-Protocol")
			c.Header("Access-Control-Allow-Methods", "PUT,DELETE,POST,GET,OPTIONS")
			c.Header("Access-Control-Expose-Headers", "Content-Length,Access-Control-Allow-Origin,Access-Control-Allow-Headers,Content-Type")
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Disposition")
		}
		//放行所有OPTIONS方法
		if method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
		}
		c.Next()
	}
}
