package router

import (
	"myblogx/api"
	"myblogx/api/article_api"
	"myblogx/api/article_api/category"
	"myblogx/api/article_api/favorite"
	"myblogx/api/article_api/tags"
	"myblogx/api/article_api/top"
	"myblogx/api/article_api/view_history"
	mw "myblogx/middleware"
	"myblogx/models"

	"github.com/gin-gonic/gin"
)

func ArticleRouter(r *gin.RouterGroup) {
	group := r.Group("articles")
	authGroup := group.Group("", mw.AuthMiddleware)
	adminGroup := authGroup.Group("", mw.AdminMiddleware)

	app := api.App.ArticleApi

	// 文章操作
	group.GET("", mw.BindQuery[article_api.ArticleListRequest], app.ArticleListView)
	group.GET(":id", mw.BindUri[models.IDRequest], app.ArticleDetailView)
	authGroup.POST("", mw.BindJson[article_api.ArticleCreateRequest], app.ArticleCreateView)
	authGroup.PUT(":id", mw.BindUri[models.IDRequest], mw.BindJson[article_api.ArticleUpdateRequest], app.ArticleUpdateView)
	authGroup.DELETE(":id", mw.BindUri[models.IDRequest], app.ArticleRemoveUserView)
	adminGroup.DELETE("", mw.CaptureLog(mw.ReqBody|mw.ReqHeader), mw.BindJson[models.IDListRequest], app.ArticleRemoveView)

	group.POST("view", mw.BindJson[article_api.ArticleViewCountRequest], app.ArticleVisitView)
	authGroup.PUT(":id/digg", mw.BindUri[models.IDRequest], app.ArticleDiggView)
	adminGroup.POST(":id/examine", mw.CaptureLog(mw.ReqBody|mw.ReqHeader), mw.BindUri[models.IDRequest], mw.BindJson[article_api.ArticleExamineRequest], app.ArticleExamineView)

	// 收藏
	group.GET("favorite", mw.BindQuery[favorite.FavoriteListRequest], app.FavoriteListView)
	group.GET("favorite/contents", mw.BindQuery[favorite.FavoriteArticlesRequest], app.FavoriteArticlesView)
	authGroup.PUT("favorite", mw.BindJson[favorite.FavoriteRequest], app.FavoriteCreateUpdateView)
	authGroup.DELETE("favorite/contents", mw.BindJson[favorite.FavoriteRemovePatchModel], app.FavoriteRemovePatchView)
	authGroup.DELETE("favorite", mw.BindJson[models.IDListRequest], app.FavoriteDeleteView)
	authGroup.POST("favorite", mw.BindJson[article_api.ArticleFavoriteRequest], app.ArticleFavoriteSaveView)

	// 置顶
	group.GET("top", mw.BindQuery[top.ArticleTopListRequest], app.ArticleTopListView)
	authGroup.POST("top", mw.BindJson[top.ArticleTopSetRequest], app.ArticleTopSetView)
	authGroup.DELETE("top", mw.BindJson[top.ArticleTopSetRequest], app.ArticleTopRemoveView)

	// 浏览历史
	authGroup.GET("history", mw.BindQuery[view_history.ArticleViewHistoryRequest], app.ArticleViewHistoryView)
	authGroup.DELETE("history", mw.BindJson[models.IDListRequest], app.ArticleViewHistoryRemoveView)

	// 分类
	group.GET("category", mw.BindQuery[category.CategoryListRequest], app.CategoryListView)
	authGroup.POST("category", mw.CaptureLog(mw.ReqBody), mw.BindJson[category.CategoryRequest], app.CategoryCreateUpdateView)
	authGroup.DELETE("category", mw.CaptureLog(mw.ReqBody|mw.ReqHeader), mw.BindJson[models.IDListRequest], app.CategoryDeleteView)
	authGroup.GET("category/options", app.CategoryOptionsView)

	// 标签
	group.GET("tags", mw.BindQuery[tags.TagListRequest], app.TagListView)
	authGroup.GET("tags/options", app.ArticleTagOptionsView)
	adminGroup.PUT("tags", mw.CaptureLog(mw.ReqBody), mw.BindJson[tags.TagRequest], app.TagCreateUpdateView)
	adminGroup.DELETE("tags", mw.CaptureLog(mw.ReqBody|mw.ReqHeader), mw.BindJson[models.IDListRequest], app.TagDeleteView)
}
