package auth_api

import (
	"myblogx/common/res"
	"myblogx/global"
	"myblogx/models"
	"myblogx/service/user_service"

	"github.com/gin-gonic/gin"
)

// EmailLoginView 邮箱验证码登录。
// 这里依赖 EmailVerifyMiddleware 先完成验证码校验，并把邮箱写入上下文。
func (AuthApi) EmailLoginView(c *gin.Context) {
	if !global.Config.Site.Login.EmailLogin {
		res.FailWithMsg("站点未启用邮箱登录功能", c)
		return
	}

	email := c.GetString("email")
	if email == "" {
		res.FailWithMsg("邮箱验证失败：邮箱不存在", c)
		return
	}

	var user models.UserModel
	if err := global.DB.Take(&user, "email = ?", email).Error; err != nil {
		// 这里不区分“邮箱不存在”和其他细节，避免把账号状态暴露给外部调用方。
		res.FailWithMsg("邮箱登录失败", c)
		return
	}
	if !user.CanLogin() {
		res.FailWithMsg(user.Status.String(), c)
		return
	}

	accessToken, refreshToken, _, err := user_service.CreateLoginTokens(&user, user_service.BuildSessionMetaFromGin(c))
	if err != nil {
		res.FailWithMsg("邮箱登录失败", c)
		return
	}

	user_service.SetRefreshTokenCookie(c, refreshToken)
	user_service.UserLoginLog(c, user.ID)

	res.OkWithData(accessToken, c)
}
