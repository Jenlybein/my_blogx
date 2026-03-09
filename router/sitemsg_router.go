package router

import (
	"myblogx/api"
	"myblogx/api/sitemsg_api"
	mw "myblogx/middleware"

	"github.com/gin-gonic/gin"
)

func SitemsgRouter(r *gin.RouterGroup) {
	Group := r.Group("sitemsg")
	authGroup := Group.Group("", mw.AuthMiddleware)
	// adminGroup := authGroup.Group("", mw.AdminMiddleware)

	app := api.App.SitemsgApi
	authGroup.GET("conf", app.UserMsgConfView)
	authGroup.PUT("conf", mw.BindJson[sitemsg_api.UserMsgConfResponseAndRequest], app.UserMsgConfUpdateView)
}
