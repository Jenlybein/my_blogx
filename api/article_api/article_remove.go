package article_api

import (
	"fmt"
	"myblogx/common/res"
	"myblogx/global"
	"myblogx/middleware"
	"myblogx/models"
	"myblogx/service/log_service"

	"github.com/gin-gonic/gin"
)

func (ArticleApi) ArticleRemoveView(c *gin.Context) {
	cr := middleware.GetBindJson[models.IDListRequest](c)

	var list []models.ArticleModel
	global.DB.Find(&list, "id in ?", cr.IDList)

	if len(list) == 0 {
		res.FailWithMsg("删除失败，文章不存在", c)
		return
	}
	if err := global.DB.Delete(&list).Error; err != nil {
		res.FailWithMsg("删除文章失败", c)
		return
	}
	res.OkWithMsg(fmt.Sprintf("文章删除成功, 成功删除%d条", len(list)), c)
	log_service.EmitActionAuditFromGin(c, log_service.GinAuditInput{
		ActionName:  "article_admin_remove",
		TargetType:  "article",
		Success:     true,
		Message:     fmt.Sprintf("文章删除成功, 成功删除%d条", len(list)),
		RequestBody: map[string]any{"id_list": cr.IDList},
		ResponseBody: map[string]any{
			"deleted_count": len(list),
		},
		UseRawRequestBody: true,
		UseRawRequestHead: true,
	})
}
