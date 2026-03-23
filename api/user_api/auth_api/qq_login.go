package auth_api

import (
	"myblogx/common/res"
	"myblogx/global"
	"myblogx/middleware"
	"myblogx/models"
	"myblogx/models/enum"
	"myblogx/service/qq_service"
	"myblogx/service/user_service"
	"myblogx/utils/jwts"

	"github.com/gin-gonic/gin"
)

type QQLoginRequest struct {
	Code string `json:"code" binding:"required"`
}

func (AuthApi) QQLoginView(c *gin.Context) {
	if !global.Config.Site.Login.QQLogin {
		res.FailWithMsg("站点未启用qq登录功能", c)
		return
	}

	cr := middleware.GetBindJson[QQLoginRequest](c)

	userInfoResp, err := qq_service.GetUserInfo(cr.Code)
	if err != nil {
		res.FailWithError(err, c)
		return
	}

	var user models.UserModel
	if err = global.DB.Take(&user, "open_id = ?", userInfoResp.OpenID).Error; err != nil {
		username, usernameErr := user_service.NextAutoUsername()
		if usernameErr != nil {
			global.Logger.Errorf("qq 登录生成用户名失败: %v", usernameErr)
			res.FailWithMsg("qq登录失败", c)
			return
		}

		// 创建用户
		user = models.UserModel{
			Username:       username,
			Nickname:       userInfoResp.NickName,
			Avatar:         userInfoResp.Avatar,
			RegisterSource: enum.RegisterQQSourceType,
			OpenID:         userInfoResp.OpenID,
			Role:           enum.RoleUser,
		}
		if err = global.DB.Create(&user).Error; err != nil {
			res.FailWithMsg("qq登录失败 "+err.Error(), c)
			return
		}
	}

	// 签发token
	token, err := jwts.GetToken(jwts.Claims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
	})
	if err != nil {
		res.FailWithMsg("qq登录失败 "+err.Error(), c)
		return
	}
	// 登录日志
	user_service.NewUserService(user).UserLogin(c)

	res.OkWithData(token, c)
}
