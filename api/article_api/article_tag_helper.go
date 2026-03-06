package article_api

import (
	"errors"

	"myblogx/global"
	"myblogx/models"
	"myblogx/service/redis_service/redis_tag"

	"gorm.io/gorm"
)

func validateArticleCategory(db *gorm.DB, userID uint, categoryID *uint) error {
	if categoryID == nil {
		return nil
	}

	var category models.CategoryModel
	return db.Take(&category, "id = ? AND user_id = ?", *categoryID, userID).Error
}

func loadEnabledTagsByIDs(db *gorm.DB, tagIDs []uint) ([]models.TagModel, error) {
	uniqueIDs := normalizeTagIDs(tagIDs)
	if len(uniqueIDs) == 0 {
		return []models.TagModel{}, nil
	}

	var tagList []models.TagModel
	if err := db.Where("id IN ? AND is_enabled = ?", uniqueIDs, true).Find(&tagList).Error; err != nil {
		return nil, err
	}

	if len(tagList) != len(uniqueIDs) {
		return nil, errors.New("标签不存在或已停用")
	}
	return tagList, nil
}

func normalizeTagIDs(tagIDs []uint) []uint {
	uniqueIDs := make([]uint, 0, len(tagIDs))
	seen := make(map[uint]struct{}, len(tagIDs))
	for _, tagID := range tagIDs {
		if tagID == 0 {
			continue
		}
		if _, ok := seen[tagID]; ok {
			continue
		}
		seen[tagID] = struct{}{}
		uniqueIDs = append(uniqueIDs, tagID)
	}
	return uniqueIDs
}

func extractTagIDs(tags []models.TagModel) []uint {
	ids := make([]uint, 0, len(tags))
	for _, tag := range tags {
		ids = append(ids, tag.ID)
	}
	return ids
}

func loadArticleTagIDs(db *gorm.DB, articleID uint) ([]uint, error) {
	var relationList []models.ArticleTagModel
	if err := db.Select("tag_id").Where("article_id = ?", articleID).Find(&relationList).Error; err != nil {
		return nil, err
	}

	tagIDs := make([]uint, 0, len(relationList))
	for _, item := range relationList {
		tagIDs = append(tagIDs, item.TagID)
	}
	return tagIDs, nil
}

func buildTagArticleCountDelta(oldTagIDs, newTagIDs []uint) map[uint]int {
	deltaMap := make(map[uint]int)
	oldSet := make(map[uint]struct{}, len(oldTagIDs))
	newSet := make(map[uint]struct{}, len(newTagIDs))

	for _, tagID := range normalizeTagIDs(oldTagIDs) {
		oldSet[tagID] = struct{}{}
	}
	for _, tagID := range normalizeTagIDs(newTagIDs) {
		newSet[tagID] = struct{}{}
	}

	for tagID := range newSet {
		if _, ok := oldSet[tagID]; !ok {
			deltaMap[tagID]++
		}
	}
	for tagID := range oldSet {
		if _, ok := newSet[tagID]; !ok {
			deltaMap[tagID]--
		}
	}

	return deltaMap
}

func applyTagArticleCountDelta(deltaMap map[uint]int) {
	for tagID, delta := range deltaMap {
		if delta == 0 {
			continue
		}
		if err := redis_tag.SetCacheArticleCount(tagID, delta); err != nil {
			global.Logger.Errorf("标签文章数缓存更新失败 tag_id=%d delta=%d err=%v", tagID, delta, err)
		}
	}
}
