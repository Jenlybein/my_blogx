package image_api

import (
	"fmt"
	"myblogx/common"
	"myblogx/common/res"
	"myblogx/global"
	"myblogx/middleware"
	"myblogx/models"
	"myblogx/service/log_service"

	"github.com/gin-gonic/gin"
)

type ImageApi struct {
}

type ImageListResponse struct {
	models.ImageModel
	WebPath string `json:"web_path"`
}

func (ImageApi) ImageListView(c *gin.Context) {
	cr := middleware.GetBindQuery[common.PageInfo](c)

	_list, count, _ := common.ListQuery(models.ImageModel{}, common.Options{
		PageInfo: cr,
		Likes:    []string{"filename"},
	})

	var respList []ImageListResponse
	for _, item := range _list {
		respList = append(respList, ImageListResponse{
			ImageModel: item,
			WebPath:    item.WebPath(),
		})
	}

	res.OkWithList(respList, count, c)
}

func (ImageApi) ImageRemoveView(c *gin.Context) {
	cr := middleware.GetBindJson[models.RemoveRequest](c)

	log := log_service.GetLog(c)
	log.SetShowRequest()
	log.SetShowResponse()
	log.ShowRequestHeader()
	log.ShowResponseHeader()

	var list []models.ImageModel
	global.DB.Find(&list, "id IN ?", cr.IDList)

	var successCount, errCount int64
	if len(list) > 0 {
		// 删除数据库记录
		successCount = global.DB.Delete(&list).RowsAffected
		errCount = int64(len(list)) - successCount
	} else {
		res.FailWithMsg("删除失败，图片不存在", c)
		return
	}

	msg := fmt.Sprintf("操作成功，删除了 %d 张图片，失败 %d 张", successCount, errCount)
	res.OkWithData(msg, c)
}
