package router

import (
	"myblogx/api"
	"myblogx/api/comment_api"
	mw "myblogx/middleware"

	"github.com/gin-gonic/gin"
)

func CommentRouter(r *gin.RouterGroup) {
	app := api.App.CommentApi

	Group := r.Group("comments")
	authGroup := Group.Group("", mw.AuthMiddleware)
	// adminGroup := authGroup.Group("", mw.AdminMiddleware)

	authGroup.POST("", mw.BindJson[comment_api.CommentCreateRequest], app.CommentCreateView)
}
