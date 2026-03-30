package log_service

import (
	"myblogx/models/ctype"
	"myblogx/utils/jwts"
	"myblogx/utils/requestmeta"

	"github.com/gin-gonic/gin"
)

// GinAuditInput 定义基于 Gin 上下文补充操作审计日志时的业务输入。
// 用于接收业务层传入的审计日志核心参数，由 Gin 上下文自动补充请求信息
type GinAuditInput struct {
	Level              string         // 日志级别（info/warn/error）
	Message            string         // 日志描述信息
	ActionName         string         // 操作名称（如：用户登录、文章发布）
	TargetType         string         // 操作目标类型（如：user/article/comment）
	TargetID           string         // 操作目标ID
	Success            bool           // 操作是否成功
	RequestBody        any            // 请求体数据
	ResponseBody       any            // 响应体数据
	UseRawRequestBody  bool           // 是否记录中间件采集到的原始请求体
	UseRawResponseBody bool           // 是否记录中间件采集到的原始响应体
	UseRawRequestHead  bool           // 是否记录中间件采集到的原始请求头
	UseRawResponseHead bool           // 是否记录中间件采集到的原始响应头
	Extra              map[string]any // 扩展字段，存储自定义信息
}

// EmitActionAuditFromGin 从 Gin 上下文提取请求元信息并写入操作审计日志。
func EmitActionAuditFromGin(c *gin.Context, input GinAuditInput) {
	// 处理 Gin 上下文为空的边界情况，直接记录基础审计日志
	if c == nil {
		EmitActionAudit(ActionAuditInput{
			Level:        input.Level,
			Message:      input.Message,
			ActionName:   input.ActionName,
			TargetType:   input.TargetType,
			TargetID:     input.TargetID,
			Success:      input.Success,
			RequestBody:  input.RequestBody,
			ResponseBody: input.ResponseBody,
			Extra:        input.Extra,
		})
		return
	}

	// 从 Gin 上下文解析 JWT Token，获取当前登录用户ID
	userID := ctype.ID(0)
	if c.Request != nil {
		if claims := jwts.GetClaimsByGin(c); claims != nil {
			userID = claims.UserID
		}
	}

	// 提取请求路由、请求方法、客户端IP、响应状态码
	// 获取 Gin 路由中定义的完整路径
	path := c.FullPath()
	method := ""
	ip := ""
	if c.Request != nil {
		// 获取 HTTP 请求方法（GET/POST/PUT/DELETE 等）
		method = c.Request.Method
		// 兼容处理：路由路径为空时，使用请求URL中的路径
		if path == "" && c.Request.URL != nil {
			path = c.Request.URL.Path
		}
		// 获取客户端真实IP
		ip = c.ClientIP()
	}
	// 获取 HTTP 响应状态码（200/400/500 等）
	statusCode := 0
	if c.Writer != nil {
		statusCode = c.Writer.Status()
	}

	// 组装完整审计日志参数，写入操作审计日志
	rawRequestBody := ""
	if input.UseRawRequestBody {
		rawRequestBody = GetRawRequestBody(c)
	}
	rawResponseBody := ""
	if input.UseRawResponseBody {
		rawResponseBody = GetRawResponseBody(c)
	}
	rawRequestHeader := ""
	if input.UseRawRequestHead {
		rawRequestHeader = GetRawRequestHeader(c)
	}
	rawResponseHeader := ""
	if input.UseRawResponseHead {
		rawResponseHeader = GetRawResponseHeader(c)
	}

	EmitActionAudit(ActionAuditInput{
		Level:             input.Level,
		Message:           input.Message,
		RequestID:         requestmeta.GetRequestID(c), // 从上下文获取请求ID，用于全链路追踪
		UserID:            userID,                      // 操作人用户ID
		IP:                ip,                          // 客户端IP
		Method:            method,                      // 请求方法
		Path:              path,                        // 请求路径
		StatusCode:        statusCode,                  // 响应状态码
		ActionName:        input.ActionName,
		TargetType:        input.TargetType,
		TargetID:          input.TargetID,
		Success:           input.Success,
		RequestBody:       input.RequestBody,
		ResponseBody:      input.ResponseBody,
		RequestBodyRaw:    rawRequestBody,
		ResponseBodyRaw:   rawResponseBody,
		RequestHeaderRaw:  rawRequestHeader,
		ResponseHeaderRaw: rawResponseHeader,
		Extra:             input.Extra,
	})
}
