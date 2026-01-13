package router

import (
	"myblogx/api"
	"myblogx/api/user_api/auth_api"
	"myblogx/api/user_api/log_api"
	"myblogx/api/user_api/profile_api"
	"myblogx/middleware"
	"myblogx/models"

	"github.com/gin-gonic/gin"
)

func UserRouter(r *gin.RouterGroup) {
	Group := r.Group("users")
	authGroup := Group.Group("", middleware.AuthMiddleware)
	adminGroup := authGroup.Group("", middleware.AdminMiddleware)

	auth := api.App.UserApi.AuthApi
	Group.POST("email/verify", middleware.CaptchaMiddleware, middleware.BindJsonMiddleware[auth_api.SendEmailRequest], auth.SendEmailView)
	Group.POST("email/register", middleware.EmailVerifyMiddleware, middleware.BindJsonMiddleware[auth_api.RegisterEmailRequest], auth.RegisterEmailView)
	Group.POST("qq", middleware.BindJsonMiddleware[auth_api.QQLoginRequest], auth.QQLoginView)
	Group.POST("login", middleware.CaptchaMiddleware, middleware.BindJsonMiddleware[auth_api.PwdLoginRequest], auth.PwdLoginView)
	Group.PUT("password/recovery/email", middleware.EmailVerifyMiddleware, middleware.BindJsonMiddleware[auth_api.ResetPasswordRequest], auth.ResetPwdByEmailView)
	authGroup.PUT("password/renewal/email", middleware.BindJsonMiddleware[auth_api.UpdatePasswordRequest], auth.UpdatePwdByEmailView)
	authGroup.PUT("email/bind", middleware.EmailVerifyMiddleware, auth.BindEmailView)

	profile := api.App.UserApi.ProfileApi
	authGroup.GET("detail", profile.UserDetailView)
	authGroup.GET("base", middleware.BindQueryMiddleware[models.IDRequest], profile.UserBaseInfoView)
	authGroup.PUT("info", middleware.BindJsonMiddleware[profile_api.UserInfoUpdateRequest], profile.UserInfoUpdateView)
	adminGroup.PUT("admin/info", middleware.BindJsonMiddleware[profile_api.AdminUserInfoUpdateRequest], profile.AdminUserInfoUpdateView)

	log := api.App.UserApi.LogApi
	authGroup.GET("login/log", middleware.BindQueryMiddleware[log_api.UserLoginListRequest], log.UserLoginLogList)
}
