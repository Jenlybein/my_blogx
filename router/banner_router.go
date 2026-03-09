package router

import (
	"myblogx/api"
	"myblogx/api/banner_api"
	mw "myblogx/middleware"
	"myblogx/models"

	"github.com/gin-gonic/gin"
)

func BannerRouter(r *gin.RouterGroup) {
	Group := r.Group("banners")
	authGroup := Group.Group("", mw.AuthMiddleware)
	adminGroup := authGroup.Group("", mw.AdminMiddleware)

	app := api.App.BannerApi

	Group.GET("", mw.BindQuery[banner_api.BannerListRequest], app.BannerListView)
	adminGroup.POST("", mw.BindJson[banner_api.BannerCreateRequest], app.BannerCreateView)
	adminGroup.PUT(":id", mw.BindUri[models.IDRequest], mw.BindJson[banner_api.BannerCreateRequest], app.BannerUpdateView)
	adminGroup.DELETE("", mw.BindJson[models.IDListRequest], app.BannerRemoveView)
}
