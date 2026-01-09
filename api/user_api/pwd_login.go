package user_api

import (
	"myblogx/common/res"
	"myblogx/global"
	"myblogx/models"
	"myblogx/service/user_service"
	"myblogx/utils/jwts"
	"myblogx/utils/pwd"

	"github.com/gin-gonic/gin"
)

type PwdLoginRequest struct {
	Username string `json:"username" binding:"required"` // 用户名或邮箱
	Password string `json:"password" binding:"required"`
}

func (u *UserApi) PwdLoginView(c *gin.Context) {
	if !global.Config.Site.Login.UsernamePwdLogin {
		res.FailWithMsg("站点未启用密码登录功能", c)
		return
	}

	var cr PwdLoginRequest
	err := c.ShouldBindJSON(&cr)
	if err != nil {
		res.FailWithError(err, c)
		return
	}

	var user models.UserModel
	if err = global.DB.Take(
		&user,
		"(username = ? OR email = ?) and (password <> '')",
		cr.Username, cr.Username,
	).Error; err != nil {
		res.FailWithMsg("用户名或邮箱不存在", c)
		return
	}

	// 校验密码
	if !pwd.CompareHashAndPassword(user.Password, cr.Password) {
		res.FailWithMsg("密码错误", c)
		return
	}

	// 签发token
	token, err := jwts.GetToken(jwts.Claims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
	})
	if err != nil {
		res.FailWithError(err, c)
		return
	}
	// 登录日志
	user_service.NewUserService(user).UserLogin(c)

	res.OkWithData(token, c)
}
