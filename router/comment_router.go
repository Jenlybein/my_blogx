package router

import (
	"myblogx/api"
	"myblogx/api/comment_api"
	mw "myblogx/middleware"
	"myblogx/models"

	"github.com/gin-gonic/gin"
)

func CommentRouter(r *gin.RouterGroup) {
	app := api.App.CommentApi

	Group := r.Group("comments")
	authGroup := Group.Group("", mw.AuthMiddleware)
	// adminGroup := authGroup.Group("", mw.AdminMiddleware)

	Group.GET("", mw.BindQuery[comment_api.CommentRootListRequest], app.CommentRootListView)
	Group.GET("replies", mw.BindQuery[comment_api.CommentReplyListRequest], app.CommentReplyListView)
	authGroup.GET("man", mw.BindQuery[comment_api.CommentManListRequest], app.CommentManListView)
	authGroup.POST("", mw.BindJson[comment_api.CommentCreateRequest], app.CommentCreateView)
	authGroup.POST(":id/digg", mw.BindUri[models.IDRequest], app.CommentDiggView)
	authGroup.DELETE(":id", mw.BindUri[models.IDRequest], app.CommentRemoveView)
}
