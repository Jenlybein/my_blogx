package image_api

import (
	"fmt"
	"myblogx/common/res"
	"myblogx/global"
	"myblogx/middleware"
	"myblogx/models"
	"myblogx/service/image_service"
	"myblogx/service/log_service"

	"github.com/gin-gonic/gin"
)

func (ImageApi) ImageRemoveView(c *gin.Context) {
	cr := middleware.GetBindJson[models.IDListRequest](c)

	log := log_service.GetLog(c)
	log.SetShowRequest()
	log.SetShowResponse()
	log.ShowRequestHeader()
	log.ShowResponseHeader()

	var list []models.ImageModel
	if err := global.DB.Find(&list, "id IN ?", cr.IDList).Error; err != nil {
		res.FailWithError(err, c)
		return
	}
	if len(list) == 0 {
		res.FailWithMsg("删除失败，图片不存在", c)
		return
	}

	for _, item := range list {
		if err := image_service.DeleteObject(item.Bucket, item.ObjectKey); err != nil {
			res.FailWithMsg(fmt.Sprintf("删除七牛对象失败: %v", err), c)
			return
		}
	}

	if err := global.DB.Unscoped().Delete(&list).Error; err != nil {
		res.FailWithError(err, c)
		return
	}

	msg := fmt.Sprintf("操作成功，删除了 %d 张图片", len(list))
	res.OkWithData(msg, c)
}
