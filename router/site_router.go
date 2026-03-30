// 站点路由定义

package router

import (
	"myblogx/api"
	"myblogx/api/site_api"
	mw "myblogx/middleware"

	"github.com/gin-gonic/gin"
)

func SiteRouter(r *gin.RouterGroup) {
	Group := r.Group("site")
	authGroup := Group.Group("", mw.AuthMiddleware)
	adminGroup := authGroup.Group("", mw.AdminMiddleware)

	app := api.App.SiteApi

	Group.GET("qq_url", app.SiteInfoQQView)
	Group.GET("seo", app.SiteSEOView)
	Group.GET("ai_info", app.SiteInfoAIView)
	Group.GET(":name", mw.BindUri[site_api.SiteInfoRequest], app.SiteInfoView)

	adminGroup.GET("admin/:name", mw.BindUri[site_api.SiteInfoRequest], app.SiteInfoAdminView)
	adminGroup.PUT(":name", mw.CaptureLog(mw.ReqBody|mw.ReqHeader), mw.BindUri[site_api.SiteInfoRequest], app.SiteUpdateView)
}
