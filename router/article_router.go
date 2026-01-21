package router

import (
	"myblogx/api"
	"myblogx/api/article_api"
	"myblogx/middleware"

	"github.com/gin-gonic/gin"
)

func ArticleRouter(r *gin.RouterGroup) {
	Group := r.Group("articles")
	authGroup := Group.Group("", middleware.AuthMiddleware)
	// adminGroup := authGroup.Group("", middleware.AdminMiddleware)

	app := api.App.ArticleApi

	authGroup.POST("", middleware.BindJsonMiddleware[article_api.ArticleCreateRequest], app.ArticleCreateView)
}
