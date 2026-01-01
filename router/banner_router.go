package router

import (
	"myblogx/api"
	"myblogx/middleware"

	"github.com/gin-gonic/gin"
)

func BannerRouter(r *gin.RouterGroup) {
	app := api.App.BannerApi

	r.GET("banner", app.BannerListView)
	r.POST("banner", middleware.AdminMiddleware, app.BannerCreateView)
	r.PUT("banner/:id", middleware.AdminMiddleware, app.BannerUpdateView)
	r.DELETE("banner", middleware.AdminMiddleware, app.BannerRemoveView)
}
