package user_service

import (
	"myblogx/global"
	"myblogx/models"
	"myblogx/models/ctype"

	"github.com/gin-gonic/gin"
)

func UserLoginLog(c *gin.Context, userID ctype.ID) {
	meta := BuildSessionMetaFromGin(c)

	if err := global.DB.Create(&models.UserLoginModel{
		UserID: userID,
		IP:     meta.IP,
		Addr:   meta.Addr,
		UA:     meta.UA,
	}).Error; err != nil {
		global.Logger.Errorf("用户登录日志创建失败: %v", err)
	}
}
