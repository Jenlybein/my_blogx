package log_service

import (
	"myblogx/models/ctype"
	"myblogx/models/enum"
	"myblogx/utils/requestmeta"

	"github.com/gin-gonic/gin"
)

// EmitLoginEventFromGin 基于 Gin 请求上下文补齐 IP、地域、UA 和 request_id 后记录登录事件。
func EmitLoginEventFromGin(c *gin.Context, eventName string, loginType enum.LoginType, success bool, username string, userID ctype.ID, reason string, extra map[string]any) {
	// 从 Gin 上下文补齐 IP、地域、UA。
	meta := requestmeta.BuildSessionMeta(c)

	// 记录登录事件。
	EmitLoginEvent(LoginEventInput{
		EventName: eventName,
		Username:  username,
		LoginType: loginType,
		Success:   success,
		Reason:    reason,
		UserID:    userID,
		IP:        meta.IP,
		Addr:      meta.Addr,
		UA:        meta.UA,
		RequestID: requestmeta.GetRequestID(c),
		Extra:     extra,
	})
}
