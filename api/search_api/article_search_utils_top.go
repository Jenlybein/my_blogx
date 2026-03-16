package search_api

import (
	"myblogx/global"
	"myblogx/models"
	"myblogx/models/enum"
)

const (
	// searchTopFlagUser 表示文章被作者本人置顶。
	// 这里使用位标记，方便一篇文章同时具备“作者置顶”和“管理员置顶”两种状态。
	searchTopFlagUser = 1
	// searchTopFlagAdmin 表示文章被管理员置顶。
	searchTopFlagAdmin = 2
)

// buildAdminTopQuery 为“空关键词的全站搜索”追加管理员置顶加权，
// 同时返回当前命中的管理员置顶文章标记。
//
// 返回值说明：
//  1. 第一个返回值是追加了 should/boost 之后的 ES 查询体。
//  2. 第二个返回值是文章置顶标记表，key 是文章 ID，value 是位标记：
//     1 表示作者置顶，2 表示管理员置顶，3 表示两者同时成立。
func buildAdminTopQuery(query map[string]any) (map[string]any, map[uint]int) {
	topMap := make(map[uint]int)
	type topRow struct {
		ArticleID uint
	}
	var rows []topRow
	if err := global.DB.Model(&models.UserTopArticleModel{}).
		Select("DISTINCT user_top_article_models.article_id").
		Joins("JOIN user_models ON user_models.id = user_top_article_models.user_id").
		Where("user_models.role = ?", enum.RoleAdmin).
		Order("user_top_article_models.created_at desc").
		Find(&rows).Error; err != nil || len(rows) == 0 {
		return query, topMap
	}

	articleIDs := make([]uint, 0, len(rows))
	for _, row := range rows {
		articleIDs = append(articleIDs, row.ArticleID)
		topMap[row.ArticleID] |= searchTopFlagAdmin
	}

	boolQuery, ok := extractSearchBoolQuery(query)
	if !ok {
		return query, topMap
	}

	should, _ := boolQuery["should"].([]any)
	should = append(should, map[string]any{
		"terms": map[string]any{
			"id":    articleIDs,
			"boost": 100,
		},
	})
	boolQuery["should"] = should

	return query, topMap
}

// buildAuthorAdminTopQuery 为“按指定作者搜索且关键词为空”的场景追加置顶加权。
//
// 这里的置顶范围只包含两类文章：
// 1. 被管理员置顶，且作者正好是传入 userID 的文章。
// 2. 被作者自己置顶，且作者正好是传入 userID 的文章。
//
// 这样可以保证 Type=7 且 Key 为空时，排在最前面的只会是该作者自己的置顶文章，
// 不会把其他作者的置顶文章混进来。
//
// 返回的 topMap 同样使用位标记：
// 1 表示作者置顶，2 表示管理员置顶，3 表示二者同时存在。
func buildAuthorAdminTopQuery(query map[string]any, userID uint) (map[string]any, map[uint]int) {
	topMap := make(map[uint]int)
	if userID == 0 {
		return query, topMap
	}

	type topRow struct {
		ArticleID uint
		TopUserID uint
		AuthorID  uint
		Role      enum.RoleType
	}

	var rows []topRow
	if err := global.DB.Model(&models.UserTopArticleModel{}).
		Select("DISTINCT user_top_article_models.article_id, user_top_article_models.user_id AS top_user_id, article_models.author_id, user_models.role").
		Joins("JOIN user_models ON user_models.id = user_top_article_models.user_id").
		Joins("JOIN article_models ON article_models.id = user_top_article_models.article_id").
		Where("article_models.author_id = ?", userID).
		Where("user_models.role = ? OR user_top_article_models.user_id = ?", enum.RoleAdmin, userID).
		Order("user_top_article_models.created_at desc").
		Find(&rows).Error; err != nil || len(rows) == 0 {
		return query, topMap
	}

	articleIDMap := make(map[uint]struct{}, len(rows))
	for _, row := range rows {
		articleIDMap[row.ArticleID] = struct{}{}
		if row.TopUserID == row.AuthorID {
			topMap[row.ArticleID] |= searchTopFlagUser
		}
		if row.Role == enum.RoleAdmin {
			topMap[row.ArticleID] |= searchTopFlagAdmin
		}
	}

	articleIDs := make([]uint, 0, len(articleIDMap))
	for articleID := range articleIDMap {
		articleIDs = append(articleIDs, articleID)
	}

	boolQuery, ok := extractSearchBoolQuery(query)
	if !ok {
		return query, topMap
	}

	should, _ := boolQuery["should"].([]any)
	should = append(should, map[string]any{
		"terms": map[string]any{
			"id":    articleIDs,
			"boost": 100,
		},
	})
	boolQuery["should"] = should
	return query, topMap
}

// isUserTop 判断某篇文章是否带有“作者置顶”标记。
func isUserTop(topMap map[uint]int, articleID uint) bool {
	return topMap[articleID]&searchTopFlagUser != 0
}

// isAdminTop 判断某篇文章是否带有“管理员置顶”标记。
func isAdminTop(topMap map[uint]int, articleID uint) bool {
	return topMap[articleID]&searchTopFlagAdmin != 0
}
