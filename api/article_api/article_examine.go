package article_api

import (
	"myblogx/common/res"
	"myblogx/global"
	"myblogx/middleware"
	"myblogx/models"

	"github.com/gin-gonic/gin"
)

func (ArticleApi) ArticleExamineView(c *gin.Context) {
	id := middleware.GetBindUri[models.IDRequest](c)
	cr := middleware.GetBindJson[ArticleExamineRequest](c)

	var article models.ArticleModel
	if err := global.DB.Take(&article, id.ID).Error; err != nil {
		res.FailWithMsg("文章不存在", c)
		return
	}

	if err := global.DB.Model(&article).Updates(models.ArticleModel{
		Status: cr.Status,
	}).Error; err != nil {
		res.FailWithMsg("文章审核失败", c)
		return
	}

	// 给文章创作者发送系统通知

	res.OkWithMsg("文章审核成功", c)
}
