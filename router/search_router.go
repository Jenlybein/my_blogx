package router

import (
	"myblogx/api"
	"myblogx/api/search_api"
	mw "myblogx/middleware"

	"github.com/gin-gonic/gin"
)

func SearchRouter(r *gin.RouterGroup) {
	group := r.Group("search")
	// authGroup := group.Group("", mw.AuthMiddleware)
	// adminGroup := authGroup.Group("", mw.AdminMiddleware)

	app := api.App.SearchApi

	group.GET("articles", mw.BindQuery[search_api.ArticleSearchRequest], app.ArticleSearchView)
}
