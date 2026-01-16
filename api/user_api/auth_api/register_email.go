package auth_api

import (
	"fmt"
	"myblogx/common/res"
	"myblogx/global"
	"myblogx/middleware"
	"myblogx/models"
	"myblogx/models/enum"
	"myblogx/service/user_service"
	"myblogx/utils/jwts"
	"myblogx/utils/pwd"

	"github.com/gin-gonic/gin"
)

type RegisterEmailRequest struct {
	Pwd string `json:"pwd" binding:"required"`
}

func (AuthApi) RegisterEmailView(c *gin.Context) {
	if !global.Config.Site.Login.EmailLogin {
		res.FailWithMsg("站点未启用邮箱注册功能", c)
		return
	}

	cr := middleware.GetBindJson[RegisterEmailRequest](c)

	email := c.GetString("email")
	if email == "" {
		res.FailWithMsg("邮箱验证失败：邮箱不存在", c)
		return
	}

	// 注册用户
	hashedPassword, err := pwd.GenerateFromPassword(cr.Pwd)
	if err != nil {
		res.FailWithMsg("邮箱注册失败", c)
		return
	}
	var maxID uint64
	global.DB.Model(&models.UserModel{}).Select("MAX(id)").Scan(&maxID)

	user := models.UserModel{
		Username:       fmt.Sprintf("%d", maxID+1+10000),
		Password:       hashedPassword,
		Nickname:       email,
		Avatar:         "xxx.png",
		RegisterSource: enum.RegisterEmailSourceType,
		Email:          email,
		Role:           enum.RoleUser,
	}
	if err = global.DB.Create(&user).Error; err != nil {
		res.FailWithMsg("邮箱注册失败", c)
		global.Logger.Errorf("邮箱注册失败 %v", err)
		return
	}

	// 签发token
	jwtToken, err := jwts.GetToken(jwts.Claims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
	})
	if err != nil {
		res.FailWithMsg("邮箱登录失败", c)
		return
	}
	// 登录日志
	user_service.NewUserService(user).UserLogin(c)

	// 返回token
	res.OkWithData(jwtToken, c)
}
