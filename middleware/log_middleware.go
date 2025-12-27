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

// 重写 Write 方法，调用时将响应体内容保存到 Body 中，日志中间件需要记录响应体内容
func (w *ResponseWriter) Write(data []byte) (int, error) {
	w.Body = append(w.Body, data...)
	return w.ResponseWriter.Write(data)
}

// 重写 Header 方法，调用时将响应头内容保存到 Head 中，日志中间件需要记录响应头内容
func (w *ResponseWriter) Header() http.Header {
	return w.Head
}

func LogMiddleware(c *gin.Context) {
	log := log_service.NewActionLogByGin(c)

	// 请求中间件
	log.SetRequest(c) // 记录请求体
	c.Set("log", log) // 将日志对象存入context

	// 为响应体内容添加记录载体，包装ResponseWriter以捕获响应
	res := &ResponseWriter{
		ResponseWriter: c.Writer,
		Head:           make(http.Header),
	}
	// 为响应头添加记录载体
	c.Writer = res

	c.Next()

	// 请求处理完成后，使用响应中间件
	log.SetResponseHeader(res.Head)
	log.SetResponse(res.Body)

	log.MiddlewareSave() // 保存日志
}
