package router

import (
	"myblogx/api"
	"myblogx/api/article_api"
	mw "myblogx/middleware"
	"myblogx/models"

	"github.com/gin-gonic/gin"
)

func ArticleRouter(r *gin.RouterGroup) {
	Group := r.Group("articles")
	authGroup := Group.Group("", mw.AuthMiddleware)
	adminGroup := authGroup.Group("", mw.AdminMiddleware)

	app := api.App.ArticleApi

	Group.GET("", mw.BindQuery[article_api.ArticleListRequest], app.ArticleListView)
	Group.GET("/:id", mw.BindUri[models.IDRequest], app.ArticleDetailView)
	Group.POST("view", mw.BindJson[article_api.ArticleViewCountRequest], app.ArticleVisitView)

	authGroup.POST("", mw.BindJson[article_api.ArticleCreateRequest], app.ArticleCreateView)
	authGroup.PUT(":id", mw.BindUri[models.IDRequest], mw.BindJson[article_api.ArticleUpdateRequest], app.ArticleUpdateView)
	authGroup.PUT(":id/digg", mw.BindUri[models.IDRequest], app.ArticleDiggView)
	authGroup.POST("favorite", mw.BindJson[article_api.ArticleFavoriteRequest], app.ArticleFavoriteSaveView)

	adminGroup.POST(":id/examine", mw.BindUri[models.IDRequest], mw.BindJson[article_api.ArticleExamineRequest], app.ArticleExamineView)
}
