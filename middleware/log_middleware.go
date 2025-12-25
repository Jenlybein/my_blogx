// 日志中间件

package middleware

import (
	"myblogx/service/log_service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ResponseWriter struct {
	gin.ResponseWriter
	Body []byte
	Head http.Header
}

func (w *ResponseWriter) Write(data []byte) (int, error) {
	w.Body = append(w.Body, data...)
	return w.ResponseWriter.Write(data)
}

func (w *ResponseWriter) Header() http.Header {
	return w.Head
}

func LogMiddleware(c *gin.Context) {
	log := log_service.NewActionLogByGin(c)

	// 请求中间件
	log.SetRequest(c)
	c.Set("log", log)

	// 为响应体内容添加记录载体
	res := &ResponseWriter{
		ResponseWriter: c.Writer,
		Head:           make(http.Header),
	}
	c.Writer = res

	c.Next()

	// 响应中间件
	log.SetResponseHeader(res.Head)
	log.SetResponse(res.Body)

	log.MiddlewareSave()
}
