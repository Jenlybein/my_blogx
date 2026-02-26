package article_api

import (
	"fmt"
	"myblogx/common/res"
	"myblogx/global"
	"myblogx/middleware"
	"myblogx/models"

	"github.com/gin-gonic/gin"
)

func (ArticleApi) ArticleRemoveView(c *gin.Context) {
	cr := middleware.GetBindJson[models.RemoveRequest](c)

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

}
