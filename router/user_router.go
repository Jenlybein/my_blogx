package router

import (
	"myblogx/api"
	"myblogx/middleware"

	"github.com/gin-gonic/gin"
)

func UserRouter(r *gin.RouterGroup) {
	app := api.App.UserApi
	r.POST("user/email/verify", middleware.CaptchaMiddleware, app.SendEmailView)
	r.POST("user/email/register", middleware.EmailVerifyMiddleware, app.RegisterEmailView)
	r.POST("user/qq", app.QQLoginView)
	r.POST("user/login", middleware.CaptchaMiddleware, app.PwdLoginView)
	r.GET("user/detail", middleware.AuthMiddleware, app.UserDetailView)
	r.GET("user/base", app.UserBaseInfoView)
}
