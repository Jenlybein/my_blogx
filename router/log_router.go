package router

import (
	"myblogx/api"
	"myblogx/api/log_api"
	"myblogx/middleware"
	"myblogx/models"

	"github.com/gin-gonic/gin"
)

func LogRouter(r *gin.RouterGroup) {
	Group := r.Group("logs")
	Group.Use(middleware.AuthMiddleware, middleware.AdminMiddleware)

	app := api.App.LogApi
	Group.GET("", middleware.BindQueryMiddleware[log_api.LogListRequest], app.LogListView)
	Group.GET(":id", middleware.BindUriMiddleware[models.IDRequest], app.LogReadView)
	Group.DELETE("", middleware.BindJsonMiddleware[models.RemoveRequest], app.LogRemoveView)
}
