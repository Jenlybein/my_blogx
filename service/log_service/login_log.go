// 登录日志服务

package log_service

import (
	"myblogx/core"
	"myblogx/global"
	"myblogx/models"
	"myblogx/models/ctype"
	"myblogx/models/enum"
	"myblogx/utils/jwts"

	"github.com/gin-gonic/gin"
)

func NewLoginSuccess(c *gin.Context, loginType enum.LoginType) {
	ip := c.ClientIP()
	addr := core.GetIpAddr(ip)

	// 解析 jwt token 中的 userID
	userID := ctype.ID(0)
	username := ""
	claims, err := jwts.ParseTokenByGin(c)
	if err != nil {
		global.Logger.Errorf("解析 token 失败: %v", err)
	} else {
		username = claims.Username
		userID = claims.UserID
	}

	global.DB.Create(&models.LogModel{
		LogType:     enum.LoginLogType,
		Title:       "用户登录成功",
		Content:     "",
		UserID:      userID,
		IP:          ip,
		Addr:        addr,
		LoginStatus: true,
		Username:    username,
		LoginType:   loginType,
	})
}

func NewLoginFail(c *gin.Context, loginType enum.LoginType, msg string, username string, password string) {
	_ = password
	ip := c.ClientIP()
	addr := core.GetIpAddr(ip)

	global.DB.Create(&models.LogModel{
		LogType:     enum.LoginLogType,
		Title:       "用户登录失败",
		Content:     msg,
		IP:          ip,
		Addr:        addr,
		LoginStatus: false,
		Username:    username,
		LoginType:   loginType,
	})
}
