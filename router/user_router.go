package router

import (
	"myblogx/api"
	"myblogx/middleware"

	"github.com/gin-gonic/gin"
)

func UserRouter(r *gin.RouterGroup) {
	userGroup := r.Group("user")

	auth := api.App.UserApi.AuthApi
	userGroup.POST("email/verify", middleware.CaptchaMiddleware, auth.SendEmailView)
	userGroup.POST("email/register", middleware.EmailVerifyMiddleware, auth.RegisterEmailView)
	userGroup.POST("qq", auth.QQLoginView)
	userGroup.POST("login", middleware.CaptchaMiddleware, auth.PwdLoginView)
	userGroup.PUT("password/renewal/email", middleware.AuthMiddleware, auth.UpdatePwdByEmailView)
	userGroup.PUT("password/recovery/email", middleware.EmailVerifyMiddleware, auth.ResetPwdByEmailView)
	userGroup.PUT("email/bind", middleware.EmailVerifyMiddleware, middleware.AuthMiddleware, auth.BindEmailView)

	profile := api.App.UserApi.ProfileApi
	userGroup.GET("detail", middleware.AuthMiddleware, profile.UserDetailView)
	userGroup.GET("base", profile.UserBaseInfoView)
	userGroup.PUT("info", middleware.AuthMiddleware, profile.UserInfoUpdateView)

	log := api.App.UserApi.LogApi
	userGroup.GET("login/log", middleware.AuthMiddleware, log.UserLoginLogList)
}
