// 登录日志服务

package log_service

import (
	"fmt"
	"myblogx/core"
	"myblogx/global"
	"myblogx/models"
	"myblogx/models/enum"
	"myblogx/utils/jwts"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func NewLoginSuccess(c *gin.Context, loginType enum.LoginType) {
	ip := c.ClientIP()
	addr := core.GetIpAddr(ip)

	token := c.Request.Header.Get("token")
	fmt.Println(token)

	// 解析 jwt token 中的 userID
	userID := uint(0)
	username := ""
	claims, err := jwts.ParseTokenByGin(c)
	if err != nil {
		logrus.Errorf("解析 token 失败: %v", err)
	} else {
		username = claims.Username
		userID = uint(claims.UserID)
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
		Password:    "-",
		LoginType:   loginType,
	})
}

func NewLoginFail(c *gin.Context, loginType enum.LoginType, msg string, username string, password string) {
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
		Password:    password,
		LoginType:   loginType,
	})
}
