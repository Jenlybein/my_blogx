package router

import (
	"myblogx/api"
	mw "myblogx/middleware"
	"myblogx/service/search_service"

	"github.com/gin-gonic/gin"
)

func SearchRouter(r *gin.RouterGroup) {
	group := r.Group("search")
	// authGroup := group.Group("", mw.AuthMiddleware)
	// adminGroup := authGroup.Group("", mw.AdminMiddleware)

	app := api.App.SearchApi

	group.GET("articles", mw.BindQuery[search_service.ArticleSearchRequest], app.ArticleSearchView)
}
