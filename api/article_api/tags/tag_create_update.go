package tags

import (
	"errors"
	"fmt"
	"myblogx/common/res"
	"myblogx/global"
	"myblogx/middleware"
	"myblogx/models"
	"myblogx/models/ctype"
	"myblogx/service/es_service"
	"myblogx/service/log_service"
	"myblogx/utils/jwts"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
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
		// 标签创建改为直接创建新记录，不再恢复同名软删数据。
		if err := global.DB.Create(&models.TagModel{
			Title:       title,
			Sort:        cr.Sort,
			Description: cr.Description,
			IsEnabled:   isEnabled,
			CreatedBy:   claims.UserID,
		}).Error; err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				res.FailWithMsg("标签名称重复", c)
				return
			}
			res.FailWithMsg(fmt.Sprintf("创建标签失败: %v", err), c)
			return
		}
		res.OkWithMsg("创建标签成功", c)
		log_service.EmitActionAuditFromGin(c, log_service.GinAuditInput{
			ActionName:        "tag_create",
			TargetType:        "tag",
			Success:           true,
			Message:           "创建标签成功",
			RequestBody:       cr,
			UseRawRequestBody: true,
		})
		return
	}

	var tag models.TagModel
	if err := global.DB.Take(&tag, cr.ID).Error; err != nil {
		res.FailWithMsg("标签不存在", c)
		return
	}
	oldTitle := tag.Title
	if cr.IsEnabled != nil {
		isEnabled = *cr.IsEnabled
	} else {
		isEnabled = tag.IsEnabled
	}

	if err := ensureTagUnique(tag.ID, title); err != nil {
		res.FailWithMsg(err.Error(), c)
		return
	}

	var affectedArticleIDs []ctype.ID
	if err := global.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&tag).Updates(map[string]any{
			"title":       title,
			"sort":        cr.Sort,
			"description": cr.Description,
			"is_enabled":  isEnabled,
		}).Error; err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				return fmt.Errorf("标签名称重复")
			}
			return err
		}
		if oldTitle != title {
			var err error
			affectedArticleIDs, err = loadArticleIDsByTagID(tx, tag.ID)
			return err
		}
		return nil
	}); err != nil {
		res.FailWithMsg(fmt.Sprintf("更新标签失败: %v", err), c)
		return
	}
	if len(affectedArticleIDs) > 0 {
		if err := es_service.UpdateESDocsTags(affectedArticleIDs); err != nil {
			global.Logger.Errorf("标签改名后刷新 ES 标签失败: 标签ID=%d 错误=%v", tag.ID, err)
		}
	}
	res.OkWithMsg("更新标签成功", c)
	log_service.EmitActionAuditFromGin(c, log_service.GinAuditInput{
		ActionName:        "tag_update",
		TargetType:        "tag",
		TargetID:          strconv.FormatUint(uint64(tag.ID), 10),
		Success:           true,
		Message:           "更新标签成功",
		RequestBody:       cr,
		UseRawRequestBody: true,
	})
}

func ensureTagUnique(currentID ctype.ID, title string) error {
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

func loadArticleIDsByTagID(tx *gorm.DB, tagID ctype.ID) ([]ctype.ID, error) {
	var relationList []models.ArticleTagModel
	if err := tx.Select("article_id").Where("tag_id = ?", tagID).Find(&relationList).Error; err != nil {
		return nil, err
	}
	if len(relationList) == 0 {
		return nil, nil
	}

	articleIDs := make([]ctype.ID, 0, len(relationList))
	seen := make(map[ctype.ID]struct{}, len(relationList))
	for _, relation := range relationList {
		if _, ok := seen[relation.ArticleID]; ok {
			continue
		}
		seen[relation.ArticleID] = struct{}{}
		articleIDs = append(articleIDs, relation.ArticleID)
	}
	return articleIDs, nil
}
