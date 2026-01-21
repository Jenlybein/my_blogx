package router

import (
	"myblogx/api"
	"myblogx/api/image_api"
	"myblogx/common"
	"myblogx/middleware"
	"myblogx/models"

	"github.com/gin-gonic/gin"
)

func ImageRouter(r *gin.RouterGroup) {
	Group := r.Group("images")
	authGroup := Group.Group("", middleware.AuthMiddleware)
	adminGroup := authGroup.Group("", middleware.AdminMiddleware)

	app := api.App.ImageApi

	authGroup.POST("", app.ImageUploadView)
	authGroup.POST("qiniu", app.GenUpToken)
	authGroup.POST("transfer", middleware.BindJsonMiddleware[image_api.TransferSaveRequest], app.TransferSaveView)

	adminGroup.GET("", middleware.BindQueryMiddleware[common.PageInfo], app.ImageListView)
	adminGroup.DELETE("", middleware.BindJsonMiddleware[models.RemoveRequest], app.ImageRemoveView)
}
