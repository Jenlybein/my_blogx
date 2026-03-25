package auth_api

import (
	"myblogx/common/res"
	"myblogx/global"
	"myblogx/middleware"
	"myblogx/models"
	"myblogx/service/redis_service/redis_user"
	"myblogx/service/user_service"
	"myblogx/utils/pwd"
	"strings"

	"github.com/gin-gonic/gin"
)

type PwdLoginRequest struct {
	Username string `json:"username" binding:"required"` // 用户名或邮箱
	Password string `json:"password" binding:"required"`
}

func (AuthApi) PwdLoginView(c *gin.Context) {
	if !global.Config.Site.Login.UsernamePwdLogin {
		res.FailWithMsg("站点未启用密码登录功能", c)
		return
	}

	cr := middleware.GetBindJson[PwdLoginRequest](c)
	
	// 登录失败次数限制
	meta := user_service.BuildSessionMetaFromGin(c)
	account := strings.TrimSpace(cr.Username)
	if !redis_user.CheckLoginAllowed(account, meta.IP) {
		res.FailWithMsg(user_service.ErrLoginTooFrequent.Error(), c)
		return
	}

	var user models.UserModel
	if err := global.DB.Take(
		&user,
		"(username = ? OR email = ?) and (password <> '')",
		account, account,
	).Error; err != nil {
		redis_user.RecordLoginFailure(account, meta.IP)
		res.FailWithMsg("账号或密码错误", c)
		return
	}

	// 校验密码
	if !pwd.CompareHashAndPassword(user.Password, cr.Password) {
		redis_user.RecordLoginFailure(account, meta.IP)
		res.FailWithMsg("账号或密码错误", c)
		return
	}
	if !user.CanLogin() {
		res.FailWithMsg(user.Status.String(), c)
		return
	}
	redis_user.ResetLoginFailures(account, meta.IP)

	// 登录成功后创建服务端会话，再签发短期访问令牌。
	token, refreshToken, _, err := user_service.CreateLoginTokens(&user, meta)
	if err != nil {
		res.FailWithError(err, c)
		return
	}
	user_service.SetRefreshTokenCookie(c, refreshToken)
	// 登录日志
	user_service.UserLoginLog(c, user.ID)

	res.OkWithData(token, c)
}
