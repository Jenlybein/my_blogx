package router

import (
	"myblogx/api"
	"myblogx/api/article_api"
	"myblogx/api/article_api/category"
	"myblogx/api/article_api/favorite"
	"myblogx/api/article_api/tags"
	"myblogx/api/article_api/view_history"
	"myblogx/middleware"
	mw "myblogx/middleware"
	"myblogx/models"

	"github.com/gin-gonic/gin"
)

func ArticleRouter(r *gin.RouterGroup) {
	group := r.Group("articles")
	authGroup := group.Group("", mw.AuthMiddleware)
	adminGroup := authGroup.Group("", mw.AdminMiddleware)

	app := api.App.ArticleApi

	group.GET("", mw.BindQuery[article_api.ArticleListRequest], app.ArticleListView)
	group.GET("category", mw.BindQuery[category.CategoryListRequest], app.CategoryListView)
	group.GET("favorite", mw.BindQuery[favorite.FavoriteListRequest], app.FavoriteListView)
	group.GET("/:id", mw.BindUri[models.IDRequest], app.ArticleDetailView)
	group.POST("view", mw.BindJson[article_api.ArticleViewCountRequest], app.ArticleVisitView)
	authGroup.POST("", mw.BindJson[article_api.ArticleCreateRequest], app.ArticleCreateView)
	authGroup.PUT(":id", mw.BindUri[models.IDRequest], mw.BindJson[article_api.ArticleUpdateRequest], app.ArticleUpdateView)
	authGroup.DELETE(":id", middleware.BindUri[models.IDRequest], app.ArticleRemoveUserView)

	authGroup.PUT("favorite", mw.BindJson[favorite.FavoriteRequest], app.FavoriteCreateUpdateView)
	authGroup.DELETE("favorite", mw.BindJson[models.RemoveRequest], app.FavoriteDeleteView)
	authGroup.GET("tags/options", app.ArticleTagOptionsView)

	authGroup.PUT(":id/digg", mw.BindUri[models.IDRequest], app.ArticleDiggView)
	authGroup.POST("favorite", mw.BindJson[article_api.ArticleFavoriteRequest], app.ArticleFavoriteSaveView)

	authGroup.GET("history", mw.BindQuery[view_history.ArticleViewHistoryRequest], app.ArticleViewHistoryView)
	authGroup.DELETE("history", mw.BindJson[models.RemoveRequest], app.ArticleViewHistoryRemoveView)

	authGroup.POST("category", mw.BindJson[category.CategoryRequest], app.CategoryCreateUpdateView)
	authGroup.DELETE("category", mw.BindJson[models.RemoveRequest], app.CategoryDeleteView)
	authGroup.GET("category/options", app.CategoryOptionsView)

	adminGroup.POST(":id/examine", mw.BindUri[models.IDRequest], mw.BindJson[article_api.ArticleExamineRequest], app.ArticleExamineView)
	adminGroup.DELETE("", mw.BindJson[models.RemoveRequest], app.ArticleRemoveView)
	adminGroup.GET("tags", mw.BindQuery[tags.TagListRequest], app.TagListView)
	adminGroup.PUT("tags", mw.BindJson[tags.TagRequest], app.TagCreateUpdateView)
	adminGroup.DELETE("tags", mw.BindJson[models.RemoveRequest], app.TagDeleteView)
}
