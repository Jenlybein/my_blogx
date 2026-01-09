package user_api

import (
	"myblogx/common/res"
	"myblogx/global"
	"myblogx/models"
	"myblogx/utils/pwd"

	"github.com/gin-gonic/gin"
)

type ResetPasswordRequest struct {
	NewPassword string `json:"new_password" binding:"required"`
}

func (UserApi) ResetPwdByEmailView(c *gin.Context) {
	var cr ResetPasswordRequest
	err := c.ShouldBindJSON(&cr)
	if err != nil {
		res.FailWithError(err, c)
		return
	}

	email := c.GetString("email")

	var user models.UserModel
	if err = global.DB.Take(&user, "email = ?", email).Error; err != nil {
		res.FailWithMsg("用户不存在", c)
		return
	}

	// 校验旧密码
	if pwd.CompareHashAndPassword(user.Password, cr.NewPassword) {
		res.FailWithMsg("新密码不能与旧密码相同", c)
		return
	}

	// 设置新密码
	hashPwd, err := pwd.GenerateFromPassword(cr.NewPassword)
	if err != nil {
		res.FailWithError(err, c)
		return
	}
	if err := global.DB.Model(&user).Update("password", hashPwd).Error; err != nil {
		res.FailWithError(err, c)
		return
	}

	res.OkWithMsg("密码重置成功", c)
}
