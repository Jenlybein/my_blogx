package router

import (
	"myblogx/api"
	"myblogx/api/user_api/auth_api"
	"myblogx/api/user_api/log_api"
	"myblogx/api/user_api/profile_api"
	"myblogx/api/user_api/user_man_api"
	mw "myblogx/middleware"
	"myblogx/models"

	"github.com/gin-gonic/gin"
)

func UserRouter(r *gin.RouterGroup) {
	Group := r.Group("users")
	authGroup := Group.Group("", mw.AuthMiddleware)
	adminGroup := authGroup.Group("", mw.AdminMiddleware)

	auth := api.App.UserApi.AuthApi
	Group.POST("email/verify", mw.CaptchaMiddleware, mw.BindJson[auth_api.SendEmailRequest], auth.SendEmailView)
	Group.POST("email/login", mw.EmailVerifyMiddleware, auth.EmailLoginView)
	Group.POST("email/register", mw.EmailVerifyMiddleware, mw.BindJson[auth_api.RegisterEmailRequest], auth.RegisterEmailView)
	Group.POST("qq", mw.BindJson[auth_api.QQLoginRequest], auth.QQLoginView)
	Group.POST("login", mw.CaptchaMiddleware, mw.BindJson[auth_api.PwdLoginRequest], auth.PwdLoginView)
	Group.POST("refresh", auth.RefreshTokenView)
	Group.PUT("password/recovery/email", mw.EmailVerifyMiddleware, mw.BindJson[auth_api.ResetPasswordRequest], auth.ResetPwdByEmailView)
	authGroup.PUT("password/renewal/email", mw.BindJson[auth_api.UpdatePasswordRequest], auth.UpdatePwdByEmailView)
	authGroup.POST("logout", auth.UserLogoutView)
	authGroup.POST("logout/all", auth.UserLogoutAllView)
	authGroup.PUT("email/bind", mw.EmailVerifyMiddleware, auth.BindEmailView)

	profile := api.App.UserApi.ProfileApi
	authGroup.GET("detail", profile.UserDetailView)
	authGroup.GET("base", mw.BindQuery[models.IDRequest], profile.UserBaseInfoView)
	authGroup.PUT("info", mw.BindJson[profile_api.UserInfoUpdateRequest], profile.UserInfoUpdateView)
	adminGroup.PUT("admin/info", mw.CaptureLog(mw.ReqBody|mw.ReqHeader), mw.BindJson[profile_api.AdminUserInfoUpdateRequest], profile.AdminUserInfoUpdateView)

	log := api.App.UserApi.LogApi
	authGroup.GET("login/log", mw.BindQuery[log_api.UserLoginListRequest], log.UserLoginLogList)

	man := api.App.UserApi.UserManApi
	adminGroup.GET("admin/list", mw.BindQuery[user_man_api.UserListRequest], man.UserListView)
}
