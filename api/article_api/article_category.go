package article_api

import (
	"fmt"
	"myblogx/common/res"
	"myblogx/global"
	"myblogx/middleware"
	"myblogx/models"
	"myblogx/utils/jwts"

	"github.com/gin-gonic/gin"
)

type CategoryRequest struct {
	ID    uint   `json:"id"`
	Title string `json:"title" binding:"required,min=2,max=20"`
}

// 创建或者编辑分类（传入ID则视为创建，不传入则视为编辑）
func (ArticleApi) CategoryCreateUpdateView(c *gin.Context) {
	cr := middleware.GetBindJson[CategoryRequest](c)
	claims := jwts.MustGetClaimsByGin(c)

	// 创建
	if cr.ID == 0 {
		if err := global.DB.Take(&models.CategoryModel{}, "title = ?", cr.Title).Error; err == nil {
			res.FailWithMsg("分类名称重复", c)
			return
		}

		if err := global.DB.Create(&models.CategoryModel{Title: cr.Title}).Error; err != nil {
			res.FailWithMsg(fmt.Sprintf("创建分类失败 %v", err), c)
			return
		}
		res.OkWithMsg("创建成功", c)
		return
	}

	// 编辑
	var category models.CategoryModel
	if err := global.DB.Take(&category, "user_id = ? and id = ?", claims.UserID, cr.ID).Error; err != nil {
		res.FailWithMsg("分类不存在", c)
		return
	}

	if err := global.DB.Model(&category).Update("title", cr.Title).Error; err != nil {
		res.FailWithMsg(fmt.Sprintf("更新分类失败 %v", err), c)
		return
	}
	res.OkWithMsg("更新分类成功", c)
}
