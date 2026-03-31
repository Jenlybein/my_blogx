package router

import (
	"myblogx/api"
	"myblogx/api/ai_api"
	mw "myblogx/middleware"

	"github.com/gin-gonic/gin"
)

func AIRouter(r *gin.RouterGroup) {
	app := api.App.AIApi

	group := r.Group("ai", mw.AuthMiddleware)
	authGroup := group.Group("", mw.AuthMiddleware)
	//adminGroup := authGroup.Group("", mw.AdminMiddleware)

	authGroup.POST("metainfo", mw.BindJson[ai_api.AIBaseRequest], app.AIArticleMetaInfoView)
	authGroup.POST("scoring/article", mw.BindJson[ai_api.AIArticleScoringRequest], app.AIArticleScoringView)
	authGroup.POST("overwrite", mw.BindJson[ai_api.AIOverwriteRequest], app.AIOverwriteView)
	authGroup.POST("diagnose", mw.BindJson[ai_api.AIDiagnoseRequest], app.AIDiagnoseView)
	authGroup.POST("search/list", mw.BindJson[ai_api.AIBaseRequest], app.AIArticleSearchListView)
	authGroup.POST("search/llm", mw.BindJson[ai_api.AIBaseRequest], app.AIArticleSearchLLMView)
}
