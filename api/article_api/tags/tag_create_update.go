package tags

import (
	"fmt"
	"myblogx/common/res"
	"myblogx/global"
	"myblogx/middleware"
	"myblogx/models"
	"myblogx/models/ctype"
	"myblogx/utils/jwts"
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

	if err := global.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&tag).Updates(map[string]any{
			"title":       title,
			"sort":        cr.Sort,
			"description": cr.Description,
			"is_enabled":  isEnabled,
		}).Error; err != nil {
			return err
		}
		if oldTitle != title {
			return syncArticleTagListByTagID(tx, tag.ID)
		}
		return nil
	}); err != nil {
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

func syncArticleTagListByTagID(tx *gorm.DB, tagID uint) error {
	var relationList []models.ArticleTagModel
	if err := tx.Select("article_id").Where("tag_id = ?", tagID).Find(&relationList).Error; err != nil {
		return err
	}
	if len(relationList) == 0 {
		return nil
	}

	articleIDs := make([]uint, 0, len(relationList))
	seen := make(map[uint]struct{}, len(relationList))
	for _, relation := range relationList {
		if _, ok := seen[relation.ArticleID]; ok {
			continue
		}
		seen[relation.ArticleID] = struct{}{}
		articleIDs = append(articleIDs, relation.ArticleID)
	}

	var articleList []models.ArticleModel
	if err := tx.Select("id").
		Where("id IN ?", articleIDs).
		Preload("Tags", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort desc, id asc")
		}).
		Find(&articleList).Error; err != nil {
		return err
	}

	for _, article := range articleList {
		tagList := make(ctype.List, 0, len(article.Tags))
		for _, item := range article.Tags {
			tagList = append(tagList, item.Title)
		}
		if err := tx.Model(&models.ArticleModel{}).
			Where("id = ?", article.ID).
			Update("tag_list", tagList).Error; err != nil {
			return err
		}
	}
	return nil
}
