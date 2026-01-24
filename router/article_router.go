package router

import (
	"myblogx/api"
	"myblogx/api/article_api"
	"myblogx/middleware"
	"myblogx/models"

	"github.com/gin-gonic/gin"
)

func ArticleRouter(r *gin.RouterGroup) {
	Group := r.Group("articles")
	authGroup := Group.Group("", middleware.AuthMiddleware)
	// adminGroup := authGroup.Group("", middleware.AdminMiddleware)

	app := api.App.ArticleApi

	Group.GET("", middleware.BindQueryMiddleware[article_api.ArticleListRequest], app.ArticleListView)
	Group.GET("/:id", middleware.BindUriMiddleware[models.IDRequest], app.ArticleDetailView)

	authGroup.POST("", middleware.BindJsonMiddleware[article_api.ArticleCreateRequest], app.ArticleCreateView)
	authGroup.PUT("/:id", middleware.BindUriMiddleware[models.IDRequest], middleware.BindJsonMiddleware[article_api.ArticleUpdateRequest], app.ArticleUpdateView)
}
