package router

import (
	"myblogx/api"
	"myblogx/api/chat_api"
	mw "myblogx/middleware"

	"github.com/gin-gonic/gin"
)

func ChatRouter(r *gin.RouterGroup) {
	group := r.Group("chat")
	authGroup := group.Group("", mw.AuthMiddleware)

	app := api.App.ChatApi
	authGroup.GET("sessions", mw.BindQuery[chat_api.ChatSessionListRequest], app.ChatSessionListView)
	authGroup.GET("messages", mw.BindQuery[chat_api.ChatMsgListRequest], app.ChatMsgListView)
}
