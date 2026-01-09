package middleware

import (
	"myblogx/common/res"
	"myblogx/global"

	"myblogx/utils/io_util"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type EmailVerifyMiddlewareRequest struct {
	EmailID   string `json:"email_id" binding:"required"`
	EmailCode string `json:"email_code" binding:"required"`
}

func EmailVerifyMiddleware(c *gin.Context) {
	// 读取并恢复请求体
	var cr EmailVerifyMiddlewareRequest
	if err := io_util.ShouldBindJSONWithRecover(c, &cr); err != nil {
		logrus.Errorf("邮箱验证失败：请求体绑定失败：%v", err)
		res.FailWithMsg("邮箱验证失败：请求体读取失败", c)
		c.Abort()
		return
	}

	// 进行邮箱验证
	info, ok := global.EmailVerifyStore.Verify(cr.EmailID, cr.EmailCode)
	if !ok {
		res.FailWithMsg("邮箱验证失败：验证码不存在", c)
		c.Abort()
		return
	}

	c.Set("email", info.Email)
}
