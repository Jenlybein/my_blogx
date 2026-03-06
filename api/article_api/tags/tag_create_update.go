package tags

import (
	"fmt"
	"myblogx/common/res"
	"myblogx/global"
	"myblogx/middleware"
	"myblogx/models"
	"myblogx/utils/jwts"
	"strings"

	"github.com/gin-gonic/gin"
)

func (TagsApi) TagCreateUpdateView(c *gin.Context) {
	cr := middleware.GetBindJson[TagRequest](c)
	claims := jwts.MustGetClaimsByGin(c)

	title := strings.TrimSpace(cr.Title)
	if title == "" {
		res.FailWithMsg("标签名称不能为空", c)
		return
	}

	isEnabled := true
	if cr.ID == 0 && cr.IsEnabled != nil {
		isEnabled = *cr.IsEnabled
	}

	if cr.ID == 0 {
		if err := ensureTagUnique(0, title); err != nil {
			res.FailWithMsg(err.Error(), c)
			return
		}

		if err := global.DB.Create(&models.TagModel{
			Title:       title,
			Sort:        cr.Sort,
			Description: cr.Description,
			IsEnabled:   isEnabled,
			CreatedBy:   claims.UserID,
		}).Error; err != nil {
			res.FailWithMsg(fmt.Sprintf("创建标签失败: %v", err), c)
			return
		}
		res.OkWithMsg("创建标签成功", c)
		return
	}

	var tag models.TagModel
	if err := global.DB.Take(&tag, cr.ID).Error; err != nil {
		res.FailWithMsg("标签不存在", c)
		return
	}
	if cr.IsEnabled != nil {
		isEnabled = *cr.IsEnabled
	} else {
		isEnabled = tag.IsEnabled
	}

	if err := ensureTagUnique(tag.ID, title); err != nil {
		res.FailWithMsg(err.Error(), c)
		return
	}

	if err := global.DB.Model(&tag).Updates(map[string]any{
		"title":       title,
		"sort":        cr.Sort,
		"description": cr.Description,
		"is_enabled":  isEnabled,
	}).Error; err != nil {
		res.FailWithMsg(fmt.Sprintf("更新标签失败: %v", err), c)
		return
	}
	res.OkWithMsg("更新标签成功", c)
}

func ensureTagUnique(currentID uint, title string) error {
	var count int64
	if err := global.DB.Model(&models.TagModel{}).
		Where("title = ? AND id <> ?", title, currentID).
		Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("标签名称重复")
	}
	return nil
}
