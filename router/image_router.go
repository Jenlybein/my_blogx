package router

import (
	"myblogx/api"
	"myblogx/api/image_api"
	"myblogx/common"
	mw "myblogx/middleware"
	"myblogx/models"

	"github.com/gin-gonic/gin"
)

func ImageRouter(r *gin.RouterGroup) {
	Group := r.Group("images")
	authGroup := Group.Group("", mw.AuthMiddleware)
	adminGroup := authGroup.Group("", mw.AdminMiddleware)

	app := api.App.ImageApi

	authGroup.POST("", app.ImageUploadView)
	authGroup.POST("qiniu", app.GenUpToken)
	authGroup.POST("transfer", mw.BindJson[image_api.TransferSaveRequest], app.TransferSaveView)

	adminGroup.GET("", mw.BindQuery[common.PageInfo], app.ImageListView)
	adminGroup.DELETE("", mw.BindJson[models.RemoveRequest], app.ImageRemoveView)
}
