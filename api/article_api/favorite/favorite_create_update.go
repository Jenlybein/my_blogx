package favorite

import (
	"fmt"
	"myblogx/common/res"
	"myblogx/global"
	"myblogx/middleware"
	"myblogx/models"
	"myblogx/utils/jwts"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// 创建或者编辑收藏夹（传入ID则视为创建，不传入则视为编辑）
func (FavoriteApi) FavoriteCreateUpdateView(c *gin.Context) {
	cr := middleware.GetBindJson[FavoriteRequest](c)
	claims := jwts.MustGetClaimsByGin(c)

	// 创建
	if cr.ID == 0 {
		if err := global.DB.Take(&models.FavoriteModel{}, "user_id = ? and title = ?", claims.UserID, cr.Title).Error; err == nil {
			res.FailWithMsg("收藏夹名称重复", c)
			return
		} else if err != gorm.ErrRecordNotFound {
			res.FailWithMsg(fmt.Sprintf("创建收藏夹失败 %v", err), c)
			return
		}

		if err := global.DB.Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "user_id"},
				{Name: "title"},
			},
			DoUpdates: clause.Assignments(map[string]any{
				"cover":      cr.Cover,
				"abstract":   cr.Abstract,
				"deleted_at": nil,
				"updated_at": time.Now(),
			}),
		}).Create(&models.FavoriteModel{
			UserID:   claims.UserID,
			Title:    cr.Title,
			Cover:    cr.Cover,
			Abstract: cr.Abstract,
		}).Error; err != nil {
			res.FailWithMsg(fmt.Sprintf("创建收藏夹失败 %v", err), c)
			return
		}
		res.OkWithMsg("创建收藏夹成功", c)
		return
	}

	// 编辑
	var favorite models.FavoriteModel
	if err := global.DB.Take(&favorite, "user_id = ? and id = ?", claims.UserID, cr.ID).Error; err != nil {
		res.FailWithMsg("收藏夹不存在", c)
		return
	}

	if err := global.DB.Model(&favorite).Updates(map[string]any{
		"title":    cr.Title,
		"cover":    cr.Cover,
		"abstract": cr.Abstract,
	}).Error; err != nil {
		res.FailWithMsg(fmt.Sprintf("更新收藏夹失败 %v", err), c)
		return
	}
	res.OkWithMsg("更新收藏夹成功", c)
}
