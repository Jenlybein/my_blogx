package auth_api

import (
	"myblogx/common/res"
	"myblogx/global"
	"myblogx/models"
	"myblogx/utils/jwts"

	"github.com/gin-gonic/gin"
)

func (AuthApi) BindEmailView(c *gin.Context) {
	email := c.GetString("email")
	if email == "" {
		res.FailWithMsg("邮箱验证失败：邮箱不存在", c)
		return
	}

	claims := jwts.MustGetClaimsByGin(c)

	var user models.UserModel
	if err := global.DB.Take(&user, claims.UserID).Error; err != nil {
		res.FailWithMsg("用户不存在", c)
		return
	}

	if err := global.DB.Model(&user).Update("email", email).Error; err != nil {
		res.FailWithError(err, c)
		return
	}

	res.OkWithMsg("邮箱绑定成功", c)
}
