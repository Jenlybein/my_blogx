// 站点路由定义

package router

import (
	"myblogx/api"

	"github.com/gin-gonic/gin"
)

func SiteRouter(r *gin.RouterGroup) {
	app := api.App.SiteApi
	r.GET("site/:name", app.SiteInfoView)
	r.GET("site/qq_url", app.SiteInfoQQView)
	r.PUT("site/:name", app.SiteUpdateView)
}
