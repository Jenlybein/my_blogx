package category

import (
	"errors"
	"fmt"
	"myblogx/common/res"
	"myblogx/global"
	"myblogx/middleware"
	"myblogx/models"
	dbservice "myblogx/service/db_service"
	"myblogx/utils/jwts"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// 创建或者编辑分类（传入ID则视为创建，不传入则视为编辑）
func (CategoryApi) CategoryCreateUpdateView(c *gin.Context) {
	cr := middleware.GetBindJson[CategoryRequest](c)
	claims := jwts.MustGetClaimsByGin(c)

	// 创建
	if cr.ID == 0 {
		// 创建分类时只看本次恢复/新建是否真正命中，避免并发下双成功。
		createdOrRestored, err := dbservice.RestoreOrCreateUnique(global.DB, dbservice.UniqueWriteOptions{
			Model: &models.CategoryModel{},
			CreateValue: &models.CategoryModel{
				Title:  cr.Title,
				UserID: claims.UserID,
			},
			Match: map[string]any{
				"user_id": claims.UserID,
				"title":   cr.Title,
			},
			RestoreAssignments: map[string]any{
				"deleted_at": nil,
				"updated_at": time.Now(),
			},
		})
		if err != nil {
			res.FailWithMsg(fmt.Sprintf("创建分类失败 %v", err), c)
			return
		}
		if !createdOrRestored {
			res.FailWithMsg("分类名称重复", c)
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
