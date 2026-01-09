package router

import (
	"myblogx/api"
	"myblogx/middleware"

	"github.com/gin-gonic/gin"
)

func UserRouter(r *gin.RouterGroup) {
	app := api.App.UserApi
	userGroup := r.Group("user")
	userGroup.POST("email/verify", middleware.CaptchaMiddleware, app.SendEmailView)
	userGroup.POST("email/register", middleware.EmailVerifyMiddleware, app.RegisterEmailView)
	userGroup.POST("qq", app.QQLoginView)
	userGroup.POST("login", middleware.CaptchaMiddleware, app.PwdLoginView)
	userGroup.GET("detail", middleware.AuthMiddleware, app.UserDetailView)
	userGroup.GET("base", app.UserBaseInfoView)
	userGroup.GET("login/log", middleware.AuthMiddleware, app.UserLoginLogList)
	userGroup.PUT("password/renewal/email", middleware.AuthMiddleware, app.UpdatePwdByEmailView)
	userGroup.PUT("password/recovery/email", middleware.EmailVerifyMiddleware, app.ResetPwdByEmailView)
	userGroup.PUT("email/bind", middleware.EmailVerifyMiddleware, middleware.AuthMiddleware, app.BindEmailView)
}
