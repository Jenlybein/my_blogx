package router

import (
	"myblogx/api"
	"myblogx/api/data_api"
	mw "myblogx/middleware"

	"github.com/gin-gonic/gin"
)

func DataRouter(r *gin.RouterGroup) {
	app := api.App.DataApi

	group := r.Group("data", mw.AuthMiddleware)
	authGroup := group.Group("", mw.AuthMiddleware)
	adminGroup := authGroup.Group("", mw.AdminMiddleware)

	adminGroup.GET("sum", app.SumView)
	adminGroup.GET("growth", mw.BindQuery[data_api.GrowthDataRequest], app.GrowthDataView)
	adminGroup.GET("article-year", app.ArticleYearDataView)
}
