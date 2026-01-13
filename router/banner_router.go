package router

import (
	"myblogx/api"
	"myblogx/api/banner_api"
	"myblogx/middleware"
	"myblogx/models"

	"github.com/gin-gonic/gin"
)

func BannerRouter(r *gin.RouterGroup) {
	Group := r.Group("banners")
	authGroup := Group.Group("", middleware.AuthMiddleware)
	adminGroup := authGroup.Group("", middleware.AdminMiddleware)

	app := api.App.BannerApi

	Group.GET("", middleware.BindQueryMiddleware[banner_api.BannerListRequest], app.BannerListView)
	adminGroup.POST("", middleware.BindJsonMiddleware[banner_api.BannerCreateRequest], app.BannerCreateView)
	adminGroup.PUT(":id", middleware.BindUriMiddleware[models.IDRequest], middleware.BindJsonMiddleware[banner_api.BannerCreateRequest], app.BannerUpdateView)
	adminGroup.DELETE("", middleware.BindJsonMiddleware[models.RemoveRequest], app.BannerRemoveView)
}
