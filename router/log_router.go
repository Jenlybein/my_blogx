package router

import (
	"myblogx/api"
	"myblogx/middleware"

	"github.com/gin-gonic/gin"
)

func LogRouter(r *gin.RouterGroup) {
	app := api.App.LogApi
	logGroup := r.Group("logs")
	logGroup.Use(middleware.AdminMiddleware)
	logGroup.GET("", app.LogListView)
	logGroup.GET(":id", app.LogReadView)
	logGroup.DELETE("", app.LogRemoveView)
}
