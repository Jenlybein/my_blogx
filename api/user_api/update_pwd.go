package user_api

import (
	"myblogx/common/res"
	"myblogx/global"
	"myblogx/models"
	"myblogx/utils/jwts"
	"myblogx/utils/pwd"

	"github.com/gin-gonic/gin"
)

type UpdatePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

func (UserApi) UpdatePwdByEmailView(c *gin.Context) {
	var cr UpdatePasswordRequest
	err := c.ShouldBindJSON(&cr)
	if err != nil {
		res.FailWithError(err, c)
		return
	}

	claims, err := jwts.GetClaimsByGin(c)
	if err != nil {
		res.FailWithError(err, c)
		return
	}

	var user models.UserModel
	if err = global.DB.Take(&user, claims.UserID).Error; err != nil {
		res.FailWithMsg("用户不存在", c)
		return
	}

	// 邮箱注册 or 已绑定邮箱的用户
	if user.Email == "" {
		res.FailWithMsg("用户未绑定邮箱", c)
		return
	}

	// 校验旧密码
	if !pwd.CompareHashAndPassword(user.Password, cr.OldPassword) {
		res.FailWithMsg("旧密码错误", c)
		return
	}

	// 校验新密码
	if cr.NewPassword == cr.OldPassword {
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

	res.OkWithMsg("密码更新成功", c)
}
