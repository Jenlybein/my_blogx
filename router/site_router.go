// 站点路由定义

package router

import (
	"myblogx/api"
	"myblogx/api/site_api"
	"myblogx/middleware"

	"github.com/gin-gonic/gin"
)

func SiteRouter(r *gin.RouterGroup) {
	Group := r.Group("site")
	authGroup := Group.Group("", middleware.AuthMiddleware)
	adminGroup := authGroup.Group("", middleware.AdminMiddleware)

	app := api.App.SiteApi
	Group.GET("qq_url", app.SiteInfoQQView)
	Group.GET(":name", middleware.BindUriMiddleware[site_api.SiteInfoRequest], app.SiteInfoView)
	adminGroup.GET("admin/:name", middleware.BindUriMiddleware[site_api.SiteInfoRequest], app.SiteInfoAdminView)
	adminGroup.PUT(":name", middleware.BindUriMiddleware[site_api.SiteInfoRequest], app.SiteUpdateView)
}
