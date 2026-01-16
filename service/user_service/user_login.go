package user_service

import (
	"myblogx/core"
	"myblogx/global"
	"myblogx/models"

	"github.com/gin-gonic/gin"
)

func (u *UserService) UserLogin(c *gin.Context) {
	ip := c.ClientIP()
	addr := core.GetIpAddr(ip)
	ua := c.Request.UserAgent()

	if err := global.DB.Create(&models.UserLoginModel{
		UserID: u.userModel.ID,
		IP:     ip,
		Addr:   addr,
		UA:     ua,
	}).Error; err != nil {
		global.Logger.Errorf("用户登录日志创建失败: %v", err)
	}
}
