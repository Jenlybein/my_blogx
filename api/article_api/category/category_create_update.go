package category

import (
	"errors"
	"fmt"
	"myblogx/common/res"
	"myblogx/global"
	"myblogx/middleware"
	"myblogx/models"
	"myblogx/utils/jwts"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// 创建或者编辑分类（传入ID则视为创建，不传入则视为编辑）
func (CategoryApi) CategoryCreateUpdateView(c *gin.Context) {
	cr := middleware.GetBindJson[CategoryRequest](c)
	claims := jwts.MustGetClaimsByGin(c)

	// 创建
	if cr.ID == 0 {
		// 分类创建改为直接创建新记录，不再恢复同名软删数据。
		if err := global.DB.Create(&models.CategoryModel{
			Title:  cr.Title,
			UserID: claims.UserID,
		}).Error; err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				res.FailWithMsg("分类名称重复", c)
				return
			}
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
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			res.FailWithMsg("分类名称重复", c)
			return
		}
		res.FailWithMsg(fmt.Sprintf("更新分类失败 %v", err), c)
		return
	}
	res.OkWithMsg("更新分类成功", c)
}
