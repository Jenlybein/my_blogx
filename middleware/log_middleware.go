package middleware

import (
	"myblogx/service/log_service"

	"github.com/gin-gonic/gin"
)

type ResponseReader struct {
	gin.ResponseWriter
	Body []byte
}

func (w *ResponseReader) Write(data []byte) (int, error) {
	w.Body = append(w.Body, data...)
	return w.ResponseWriter.Write(data)
}

func LogMiddleware(c *gin.Context) {
	log := log_service.NewActionLogByGin(c)

	// 请求中间件
	log.SetRequest(c)
	c.Set("log", log)

	// 为响应体内容添加记录载体
	res := &ResponseReader{
		ResponseWriter: c.Writer,
	}
	c.Writer = res

	c.Next()

	// 响应中间件
	log.SetResponse(res.Body)
	log.Save()
}
