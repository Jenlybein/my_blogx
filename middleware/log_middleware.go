// 日志中间件

package middleware

import (
	"myblogx/service/log_service"

	"github.com/gin-gonic/gin"
)

type ResponseWriter struct {
	gin.ResponseWriter
	Body []byte
}

// 重写 Write 方法，调用时将响应体内容保存到 Body 中，日志中间件需要记录响应体内容
func (w *ResponseWriter) Write(data []byte) (int, error) {
	w.Body = append(w.Body, data...)
	return w.ResponseWriter.Write(data)
}

// 重写Header()不必要，能直接读取到响应头
// func (w *ResponseWriter) Header() http.Header {
// 	return w.Head
// }

func LogMiddleware(c *gin.Context) {
	log := log_service.NewActionLogByGin(c)

	// 1.记录请求信息
	log.SetRequest(c) // 记录请求体
	c.Set("log", log) // 将日志对象存入context

	// 2. 包装ResponseWriter，为响应体添加记录载体，接管响应流
	res := &ResponseWriter{
		ResponseWriter: c.Writer,
		Body:           make([]byte, 0),
	}
	c.Writer = res

	// 3. 执行后续业务逻辑（控制器/其他中间件）
	c.Next()

	// 4. 业务执行完成，记录响应信息
	log.SetResponseHeader(c)  // 记录响应头
	log.SetResponse(res.Body) // 记录响应体

	log.MiddlewareSave() // 保存日志
}
