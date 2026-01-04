package router

import (
	"myblogx/api"
	"myblogx/middleware"

	"github.com/gin-gonic/gin"
)

func UserRouter(r *gin.RouterGroup) {
	app := api.App.UserApi
	r.POST("user/email/verify", middleware.CaptchaMiddleware, app.SendEmailView)
	r.POST("user/email/register", app.RegisterEmailView)
}
