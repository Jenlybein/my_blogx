package middleware

import (
	"bytes"
	"io"
	"myblogx/common/res"
	"myblogx/global"
	io_utils "myblogx/utils/io_util"

	"github.com/gin-gonic/gin"
)

type CaptchaMiddlewareRequest struct {
	CaptchaID   string `json:"captcha_id" binding:"required"`
	CaptchaCode string `json:"captcha_code" binding:"required"`
}

func CaptchaMiddleware(c *gin.Context) {
	if !global.Config.Site.Login.Captcha {
		return
	}

	byteData, err := io_utils.GetBody(&c.Request.Body)
	if err != nil {
		res.FailWithMsg("获取请求体失败", c)
		c.Abort()
		return
	}

	var cr CaptchaMiddlewareRequest
	err = c.ShouldBindJSON(&cr)
	if err != nil {
		res.FailWithMsg("图形验证码参数错误", c)
		c.Abort()
		return
	}

	if !global.ImageCaptchaStore.Verify(cr.CaptchaID, cr.CaptchaCode, true) {
		res.FailWithMsg("图形验证码错误", c)
		c.Abort()
		return
	}

	// ShouldBindJSON 绑定参数后，请求体内容会被消耗，所以需要恢复
	c.Request.Body = io.NopCloser(bytes.NewBuffer(byteData))
}
