package middleware

import (
	"bytes"
	"github.com/gin-gonic/gin"
	"github.com/lfxnxf/emo-frame/logging"
	"github.com/lfxnxf/emo-frame/trace"
	"github.com/lfxnxf/emo-frame/utils"
	"github.com/lfxnxf/emo-frame/zd_http/http_ctx"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"time"
)

type responseWriter struct {
	gin.ResponseWriter
	b *bytes.Buffer
}

func (w responseWriter) Write(b []byte) (int, error) {
	w.b.Write(b)
	return w.ResponseWriter.Write(b)
}

func (w responseWriter) WriteString(s string) (int, error) {
	w.b.WriteString(s)
	return w.ResponseWriter.WriteString(s)
}

func loggingAccess() http_ctx.HttpHandler {
	return func(c *http_ctx.HttpContext) {
		if c.Request.Method == http.MethodOptions {
			c.Next()
		} else {
			// 当前时间
			nowTime := time.Now()

			// response
			writer := &responseWriter{
				c.Writer,
				bytes.NewBuffer([]byte{}),
			}
			c.Writer = writer

			// request
			var reqBody string
			if c.Request.Method == http.MethodPost {
				requestBody, _ := c.GetRawData()
				c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(requestBody))
				reqBody = string(requestBody)
			}

			c.Next()

			if c.Request.RequestURI != "/ping" {
				// 获取trace_id
				var traceId string
				span, ok := c.Value(string(trace.CtxTraceSpanKey)).(trace.Span)
				if ok {
					traceId = span.Trace()
				}
				// 服务名称
				hostname, _ := os.Hostname()

				if len(reqBody) >= 512 {
					reqBody = reqBody[0:512]
				}

				var replyBody = writer.b.String()
				if len(replyBody) >= 512 {
					replyBody = replyBody[0:512]
				}

				logItems := []interface{}{
					"start", nowTime.Format(utils.TimeFormatYYYYMMDDHHmmSS),
					"cost", math.Ceil(float64(time.Since(nowTime).Nanoseconds()) / 1e6),
					"trace_id", traceId,
					"host_ip", utils.GetHost(),
					"host_name", hostname,
					"req_method", c.Request.Method,
					"req_uri", c.Request.RequestURI,
					"real_ip", c.ClientIP(),
					"http_code", c.Writer.Status(),
					"req_body", reqBody,
					"resp_body", replyBody,
				}
				logging.DefaultKit.A().Debugw("http_server", logItems...)
			}
		}
	}
}
