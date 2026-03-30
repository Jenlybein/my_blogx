package router

import (
	"myblogx/api"
	global_notif_api "myblogx/api/global_msg_api"
	mw "myblogx/middleware"
	"myblogx/models"

	"github.com/gin-gonic/gin"
)

func GlobalNotifRouter(r *gin.RouterGroup) {
	Group := r.Group("global_notif")
	authGroup := Group.Group("", mw.AuthMiddleware)
	adminGroup := authGroup.Group("", mw.AdminMiddleware)

	app := api.App.GlobalNotifApi

	authGroup.GET("", mw.BindQuery[global_notif_api.GlobalNotifListRequest], app.GlobalNotifListView)
	authGroup.POST("read", mw.BindJson[models.IDListRequest], app.GlobalNotifReadView)
	authGroup.DELETE("user", mw.BindJson[models.IDListRequest], app.GlobalNotifUserRemoveView)

	adminGroup.POST("", mw.CaptureLog(mw.ReqBody|mw.ReqHeader), mw.BindJson[global_notif_api.GlobalNotifCreateRequest], app.GlobalNotifCreateView)
	adminGroup.DELETE("", mw.CaptureLog(mw.ReqBody|mw.ReqHeader), mw.BindJson[models.IDListRequest], app.GlobalNotifAdminRemoveView)
}
