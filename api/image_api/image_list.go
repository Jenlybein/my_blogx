package image_api

import (
	"myblogx/common"
	"myblogx/common/res"
	"myblogx/middleware"
	"myblogx/models"

	"github.com/gin-gonic/gin"
)

func (ImageApi) ImageListView(c *gin.Context) {
	cr := middleware.GetBindQuery[common.PageInfo](c)

	list, count, err := common.ListQuery(models.ImageModel{}, common.Options{
		PageInfo: cr,
		Likes:    []string{"file_name", "object_key"},
	})
	if err != nil {
		res.FailWithError(err, c)
		return
	}
	res.OkWithList(list, count, c)
}
