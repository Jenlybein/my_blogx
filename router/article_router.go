package router

import (
	"myblogx/api"
	"myblogx/api/article_api"
	"myblogx/middleware"
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
	authGroup.DELETE(":id", middleware.BindUri[models.IDRequest], app.ArticleRemoveUserView)
	authGroup.GET("history", mw.BindQuery[article_api.ArticleViewHistoryRequest], app.ArticleViewHistoryView)
	authGroup.DELETE("history", mw.BindJson[models.RemoveRequest], app.ArticleViewHistoryRemoveView)

	adminGroup.POST(":id/examine", mw.BindUri[models.IDRequest], mw.BindJson[article_api.ArticleExamineRequest], app.ArticleExamineView)
	adminGroup.DELETE("", mw.BindJson[models.RemoveRequest], app.ArticleRemoveView)

	Group.GET("category", mw.BindQuery[article_api.CategoryListRequest], app.CategoryListView)
	authGroup.POST("category", mw.BindJson[article_api.CategoryRequest], app.CategoryCreateUpdateView)
	authGroup.DELETE("category", mw.BindJson[models.RemoveRequest], app.CategoryDeleteView)

	Group.GET("favorite", mw.BindQuery[article_api.FavoriteListRequest], app.FavoriteListView)
	authGroup.PUT("favorite", mw.BindJson[article_api.FavoriteRequest], app.FavoriteCreateUpdateView)
	authGroup.DELETE("favorite", mw.BindJson[models.RemoveRequest], app.FavoriteDeleteView)

}
