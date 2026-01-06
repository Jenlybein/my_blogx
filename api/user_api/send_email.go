package user_api

import (
	"fmt"
	"myblogx/common/res"
	"myblogx/global"
	"myblogx/models"
	"myblogx/service/email_service"

	"github.com/gin-gonic/gin"
	"github.com/mojocn/base64Captcha"
	"github.com/sirupsen/logrus"
)

type SendEmailRequest struct {
	Type  int8   `json:"type" oneof:"1,2"` // 1:注册 2:重置密码
	Email string `json:"email" binding:"required,email"`
}

type SendEmailResponse struct {
	ID string `json:"id"`
}

func (u *UserApi) SendEmailView(c *gin.Context) {
	if !global.Config.Site.Login.EmailLogin {
		res.FailWithMsg("站点未启用邮箱功能", c)
		return
	}

	var cr SendEmailRequest
	err := c.ShouldBindJSON(&cr)
	if err != nil {
		res.FailWithError(err, c)
		return
	}

	var user models.UserModel
	code := base64Captcha.RandText(4, "0123456789")
	timeout := 5 // 验证码有效期5分钟
	isEmailExist := global.DB.Take(&user, "email = ?", cr.Email).Error == nil

	switch cr.Type {
	case 1:
		// 注册逻辑
		if isEmailExist {
			res.FailWithMsg("注册失败：邮箱已被使用", c)
			return
		}
		err = email_service.SendRegisterCode(cr.Email, code, timeout)
	case 2:
		// 重置密码逻辑
		if !isEmailExist {
			res.FailWithMsg("重置密码失败：邮箱不存在", c)
			return
		}
		err = email_service.SendResetPwdCode(cr.Email, code, timeout)
	default:
		res.FailWithMsg("邮件发送失败：不存在的操作类型", c)
		return
	}
	if err != nil {
		fmt.Println(err)
		logrus.Errorf("邮件发送失败: %v", err)
		res.FailWithMsg("邮件发送失败", c)
		return
	}

	id := base64Captcha.RandomId()
	global.EmailVerifyStore.Store(id, cr.Email, code)

	res.OkWithData(SendEmailResponse{ID: id}, c)
}
