package es_service

import (
	"fmt"
	"myblogx/global"
	"myblogx/models"
	"myblogx/models/enum"
	"slices"
	"strconv"

	"gorm.io/gorm"
)

type articleESTop struct {
	AdminTop  bool
	AuthorTop bool
}

// BuildArticleESDocument 将文章及其聚合字段转换为 ES 文档。
func BuildArticleESDocument(article models.ArticleModel, adminTop, authorTop bool) map[string]any {
	commentsToggle := 0
	if article.CommentsToggle {
		commentsToggle = 1
	}

	tags := make([]models.ESTag, 0, len(article.Tags))
	for _, tag := range article.Tags {
		tags = append(tags, models.ESTag{
			ID:    tag.ID,
			Title: tag.Title,
		})
	}

	return map[string]any{
		"id":              article.ID,
		"created_at":      article.CreatedAt,
		"updated_at":      article.UpdatedAt,
		"title":           article.Title,
		"abstract":        article.Abstract,
		"html_content":    article.HtmlContent,
		"category_id":     article.CategoryID,
		"cover":           article.Cover,
		"author_id":       article.AuthorID,
		"view_count":      article.ViewCount,
		"digg_count":      article.DiggCount,
		"comment_count":   article.CommentCount,
		"favor_count":     article.FavorCount,
		"status":          article.Status,
		"comments_toggle": commentsToggle,
		"tags":            tags,
		"admin_top":       adminTop,
		"author_top":      authorTop,
	}
}

// SyncESDocs 按文章 ID 批量重建 ES 文档。
// 这里会从数据库重新读取文章、标签和置顶信息，再统一索引到 ES。
func SyncESDocs(articleIDs []uint) error {
	if global.DB == nil || global.ESClient == nil {
		return nil
	}

	articleIDs = normalizeArticleIDs(articleIDs)
	if len(articleIDs) == 0 {
		return nil
	}

	articleList, err := loadArticlesForES(global.DB, articleIDs)
	if err != nil {
		return err
	}
	if len(articleList) == 0 {
		return nil
	}

	topMap, err := loadArticleESTopMap(global.DB, articleIDs)
	if err != nil {
		return err
	}

	reqs := make([]*BulkRequest, 0, len(articleList))
	for _, article := range articleList {
		top := topMap[article.ID]
		reqs = append(reqs, &BulkRequest{
			Action: ActionIndex,
			ID:     strconv.FormatUint(uint64(article.ID), 10),
			Data:   BuildArticleESDocument(article, top.AdminTop, top.AuthorTop),
		})
	}

	if len(reqs) == 0 {
		return nil
	}

	resp := IndexBulk(models.ArticleModel{}.Index(), reqs)
	if !resp.Success {
		return fmt.Errorf("同步文章 ES 文档失败: %s", resp.Msg)
	}
	if data, ok := resp.Data.(map[string]any); ok {
		if hasErrors, ok := data["errors"].(bool); ok && hasErrors {
			return fmt.Errorf("同步文章 ES 文档失败: bulk errors")
		}
	}
	return nil
}

// UpdateESDocsTags 在文章标签关系变化后刷新对应文章的 ES 文档。
func UpdateESDocsTags(articleIDs []uint) error {
	return SyncESDocs(articleIDs)
}

// UpdateESDocsTop 在文章置顶状态变化后刷新对应文章的 ES 文档。
func UpdateESDocsTop(articleIDs []uint) error {
	return SyncESDocs(articleIDs)
}

func normalizeArticleIDs(articleIDs []uint) []uint {
	if len(articleIDs) == 0 {
		return nil
	}

	result := make([]uint, 0, len(articleIDs))
	seen := make(map[uint]struct{}, len(articleIDs))
	for _, articleID := range articleIDs {
		if articleID == 0 {
			continue
		}
		if _, ok := seen[articleID]; ok {
			continue
		}
		seen[articleID] = struct{}{}
		result = append(result, articleID)
	}
	slices.Sort(result)
	return result
}

func loadArticlesForES(db *gorm.DB, articleIDs []uint) ([]models.ArticleModel, error) {
	var articleList []models.ArticleModel
	err := db.Select(
		"id",
		"created_at",
		"updated_at",
		"title",
		"abstract",
		"html_content",
		"category_id",
		"cover",
		"author_id",
		"view_count",
		"digg_count",
		"comment_count",
		"favor_count",
		"status",
		"comments_toggle",
	).
		Where("id IN ?", articleIDs).
		Preload("Tags", func(db *gorm.DB) *gorm.DB {
			return db.Select("tag_models.id", "tag_models.title").Order("sort desc, id asc")
		}).
		Order("id asc").
		Find(&articleList).Error
	return articleList, err
}

func loadArticleESTopMap(db *gorm.DB, articleIDs []uint) (map[uint]articleESTop, error) {
	topMap := make(map[uint]articleESTop, len(articleIDs))
	if len(articleIDs) == 0 {
		return topMap, nil
	}

	type topRow struct {
		ArticleID uint
		TopUserID uint
		AuthorID  uint
		Role      enum.RoleType
	}

	var rows []topRow
	err := db.Model(&models.UserTopArticleModel{}).
		Select("user_top_article_models.article_id, user_top_article_models.user_id AS top_user_id, article_models.author_id, user_models.role").
		Joins("JOIN article_models ON article_models.id = user_top_article_models.article_id").
		Joins("JOIN user_models ON user_models.id = user_top_article_models.user_id").
		Where("user_top_article_models.article_id IN ?", articleIDs).
		Find(&rows).Error
	if err != nil {
		return topMap, err
	}

	for _, row := range rows {
		state := topMap[row.ArticleID]
		if row.Role == enum.RoleAdmin {
			state.AdminTop = true
		}
		if row.TopUserID == row.AuthorID {
			state.AuthorTop = true
		}
		topMap[row.ArticleID] = state
	}
	return topMap, nil
}
