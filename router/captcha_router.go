package router

import (
	"myblogx/api"

	"github.com/gin-gonic/gin"
)

func CaptchaRouter(r *gin.RouterGroup) {
	api := api.App.ImageCaptchaApi
	r.GET("/imagecaptcha", api.CaptchaView)
}
