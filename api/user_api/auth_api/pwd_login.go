package auth_api

import (
	"myblogx/common/res"
	"myblogx/global"
	"myblogx/middleware"
	"myblogx/models"
	"myblogx/models/ctype"
	"myblogx/models/enum"
	"myblogx/service/log_service"
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
		log_service.EmitLoginEventFromGin(c, "login_fail", enum.PasswordLoginType, false, "", 0, "站点未启用密码登录", nil)
		res.FailWithMsg("站点未启用密码登录功能", c)
		return
	}

	cr := middleware.GetBindJson[PwdLoginRequest](c)

	// 登录失败次数限制
	meta := user_service.BuildSessionMetaFromGin(c)
	account := strings.TrimSpace(cr.Username)
	if !redis_user.CheckLoginAllowed(account, meta.IP) {
		log_service.EmitLoginEventFromGin(c, "login_risk_control", enum.PasswordLoginType, false, account, 0, user_service.ErrLoginTooFrequent.Error(), map[string]any{
			"username": account,
		})
		res.FailWithMsg(user_service.ErrLoginTooFrequent.Error(), c)
		return
	}

	var user models.UserModel
	if err := global.DB.Take(
		&user,
		"(username = ? OR email = ?) and (password <> '')",
		account, account,
	).Error; err != nil {
		// Redis 记录登录失败
		redis_user.RecordLoginFailure(account, meta.IP)
		// 日志记录登录失败
		log_service.EmitLoginEventFromGin(c, "login_fail", enum.PasswordLoginType, false, account, 0, "账号或密码错误", map[string]any{
			"username": account,
		})
		res.FailWithMsg("账号或密码错误", c)
		return
	}

	// 校验密码
	if !pwd.CompareHashAndPassword(user.Password, cr.Password) {
		redis_user.RecordLoginFailure(account, meta.IP)
		log_service.EmitLoginEventFromGin(c, "login_fail", enum.PasswordLoginType, false, account, ctype.ID(user.ID), "账号或密码错误", map[string]any{
			"username": account,
		})
		res.FailWithMsg("账号或密码错误", c)
		return
	}

	// 账号被禁用
	if !user.CanLogin() {
		log_service.EmitLoginEventFromGin(c, "login_fail", enum.PasswordLoginType, false, user.Username, user.ID, user.Status.String(), map[string]any{
			"username": user.Username,
		})
		res.FailWithMsg(user.Status.String(), c)
		return
	}

	// Redis 记录登录成功
	redis_user.ResetLoginFailures(account, meta.IP)

	// 登录成功后创建服务端会话，再签发短期访问令牌。
	token, refreshToken, _, err := user_service.CreateLoginTokens(&user, meta)
	if err != nil {
		log_service.EmitLoginEventFromGin(c, "login_fail", enum.PasswordLoginType, false, user.Username, user.ID, err.Error(), map[string]any{
			"username": user.Username,
		})
		res.FailWithError(err, c)
		return
	}

	// 设置刷新令牌
	user_service.SetRefreshTokenCookie(c, refreshToken)

	// 记录登录日志
	log_service.EmitLoginEventFromGin(c, "login_success", enum.PasswordLoginType, true, user.Username, user.ID, "", map[string]any{
		"username": user.Username,
	})

	res.OkWithData(token, c)
}
