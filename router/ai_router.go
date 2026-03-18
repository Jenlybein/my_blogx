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

	authGroup.GET("article_metainfo", mw.BindJson[ai_api.AIArticleMetaInfoRequest], app.AIArticleMetaInfoView)
}
