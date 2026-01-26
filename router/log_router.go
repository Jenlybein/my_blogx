package router

import (
	"myblogx/api"
	"myblogx/api/log_api"
	mw "myblogx/middleware"
	"myblogx/models"

	"github.com/gin-gonic/gin"
)

func LogRouter(r *gin.RouterGroup) {
	Group := r.Group("logs")
	Group.Use(mw.AuthMiddleware, mw.AdminMiddleware)

	app := api.App.LogApi
	Group.GET("", mw.BindQuery[log_api.LogListRequest], app.LogListView)
	Group.GET(":id", mw.BindUri[models.IDRequest], app.LogReadView)
	Group.DELETE("", mw.BindJson[models.RemoveRequest], app.LogRemoveView)
}
