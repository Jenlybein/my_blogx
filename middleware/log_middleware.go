package middleware

import (
	"strconv"
	"time"

	"myblogx/global"
	"myblogx/service/db_service"
	"myblogx/service/log_service"
	"myblogx/utils/jwts"
	"myblogx/utils/requestmeta"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// LogMiddleware 运行日志中间件
func LogMiddleware(c *gin.Context) {
	// 生成全局唯一的请求ID，用于全链路日志追踪
	requestID := newRequestID()
	// 将请求ID存入 Gin 上下文，方便后续日志/业务使用
	requestmeta.SetRequestID(c, requestID)
	// 将请求ID写入响应头，前端/网关可用于问题排查
	c.Writer.Header().Set("X-Request-Id", requestID)

	// 记录请求开始时间，用于计算接口耗时
	start := time.Now()

	// 执行后续中间件/接口处理逻辑
	c.Next()

	// 记录请求结束时间，用于计算接口耗时
	if !global.Config.Log.RequestLogEnabled {
		return
	}

	fields := logrus.Fields{
		"request_id":  requestID,                        // 请求唯一ID
		"event_name":  "http_request",                   // 事件类型标识
		"method":      c.Request.Method,                 // HTTP 请求方法
		"path":        c.FullPath(),                     // Gin 路由定义路径
		"status_code": c.Writer.Status(),                // HTTP 响应状态码
		"latency_ms":  time.Since(start).Milliseconds(), // 请求耗时（毫秒）
		"ip":          c.ClientIP(),                     // 客户端真实IP
	}
	// 兼容处理：路由路径为空时，使用原始请求URL路径
	if rawPath := c.Request.URL.Path; rawPath != "" && fields["path"] == "" {
		fields["path"] = rawPath
	}

	// 尝试解析 JWT Token，存在则记录操作用户ID
	if claims, err := jwts.ParseTokenByGin(c); err == nil {
		fields["user_id"] = uint64(claims.UserID)
	}

	// 获取运行时日志实例并绑定字段
	entry := log_service.RuntimeEntry(fields)
	switch {
	case c.Writer.Status() >= 500:
		entry.Error("请求执行失败")
	case c.Writer.Status() >= 400:
		entry.Warn("请求执行完成")
	default:
		entry.Info("请求执行完成")
	}
}

// newRequestID 生成唯一请求ID
func newRequestID() string {
	// 尝试从雪花算法生成唯一ID
	if id, err := db_service.NextSnowflakeID(); err == nil {
		return strconv.FormatUint(uint64(id), 10)
	}
	// 降级方案：使用当前时间纳秒值作为请求ID
	return strconv.FormatInt(time.Now().UnixNano(), 10)
}
