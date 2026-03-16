package search_api

import (
	"encoding/json"
	"myblogx/global"
	"myblogx/models"
	"myblogx/service/redis_service/redis_article"
	"myblogx/utils/markdown"
)

// extractHighlightValues 提取高亮值
func extractHighlightValues(highlightMap map[string]any, field string) []string {
	rawList, ok := highlightMap[field].([]any)
	if !ok {
		return nil
	}

	result := make([]string, 0, len(rawList))
	for _, rawValue := range rawList {
		value, ok := rawValue.(string)
		if !ok {
			continue
		}
		result = append(result, value)
	}
	return result
}

// extractSearchBoolQuery 提取搜索 bool 查询
func extractSearchBoolQuery(query map[string]any) (map[string]any, bool) {
	functionScore, ok := query["function_score"].(map[string]any)
	if !ok {
		return nil, false
	}
	queryBody, ok := functionScore["query"].(map[string]any)
	if !ok {
		return nil, false
	}
	boolQuery, ok := queryBody["bool"].(map[string]any)
	return boolQuery, ok
}

// loadSearchArticleCounterMaps 批量读取 Redis 中的文章计数增量。
// 搜索结果里的计数字段以 ES 文档为基础值，再叠加 Redis 中尚未落库的实时增量。
func loadSearchArticleCounterMaps(articleIDs []uint) (favorMap, diggMap, viewMap, commentMap map[uint]int) {
	favorMap = make(map[uint]int)
	diggMap = make(map[uint]int)
	viewMap = make(map[uint]int)
	commentMap = make(map[uint]int)
	if global.Redis == nil || len(articleIDs) == 0 {
		return favorMap, diggMap, viewMap, commentMap
	}

	favorMap = redis_article.GetBatchCacheFavorite(articleIDs)
	diggMap = redis_article.GetBatchCacheDigg(articleIDs)
	viewMap = redis_article.GetBatchCacheView(articleIDs)
	commentMap = redis_article.GetBatchCacheComment(articleIDs)
	return favorMap, diggMap, viewMap, commentMap
}

// loadSearchArticleDisplayMetaMap 批量读取搜索列表需要的展示信息。
// 这里只补齐列表页展示字段，避免逐条查询分类和作者信息。
func loadSearchArticleDisplayMetaMap(articleIDs []uint) map[uint]SearchListResponse {
	metaMap := make(map[uint]SearchListResponse)
	if global.DB == nil || len(articleIDs) == 0 {
		return metaMap
	}

	type articleDisplayMeta struct {
		ID            uint
		CategoryTitle string
		UserNickname  string
		UserAvatar    string
	}

	var rows []articleDisplayMeta
	if err := global.DB.Model(&models.ArticleModel{}).
		Select(
			"article_models.id",
			"category_models.title AS category_title",
			"user_models.nickname AS user_nickname",
			"user_models.avatar AS user_avatar",
		).
		Joins("LEFT JOIN category_models ON category_models.id = article_models.category_id").
		Joins("LEFT JOIN user_models ON user_models.id = article_models.author_id").
		Where("article_models.id IN ?", articleIDs).
		Find(&rows).Error; err != nil {
		return metaMap
	}

	for _, row := range rows {
		metaMap[row.ID] = SearchListResponse{
			CategoryTitle: row.CategoryTitle,
			UserNickname:  row.UserNickname,
			UserAvatar:    row.UserAvatar,
		}
	}
	return metaMap
}

// extractArticleSearchResults 提取文章搜索结果
func extractArticleSearchResults(data map[string]any, topMap map[uint]int) (list []SearchListResponse) {
	hits, _ := data["hits"].([]any)
	list = make([]SearchListResponse, 0, len(hits))

	for _, hit := range hits {
		item, ok := hit.(map[string]any)
		if !ok {
			continue
		}

		var article models.ArticleModel
		if sourceMap, ok := item["_source"].(map[string]any); ok {
			jsonBytes, _ := json.Marshal(sourceMap)
			_ = json.Unmarshal(jsonBytes, &article)
		}

		// 处理高亮，获取第一个高亮字段
		highlightMap, _ := item["highlight"].(map[string]any)
		title := article.Title
		abstract := markdown.ExtractText(article.Abstract, 120)
		contentHead := article.ContentHead
		if contentHead == "" {
			contentHead = markdown.ExtractText(article.HtmlContent, 150)
		}
		if len(highlightMap) > 0 {
			if values := extractHighlightValues(highlightMap, "title"); len(values) > 0 {
				title = values[0]
			}
			if values := extractHighlightValues(highlightMap, "abstract"); len(values) > 0 {
				abstract = values[0]
			}
			if values := extractHighlightValues(highlightMap, "html_content"); len(values) > 0 {
				contentHead = values[0]
			}
		}

		list = append(list, SearchListResponse{
			ID:             article.ID,
			CreatedAt:      article.CreatedAt,
			UpdatedAt:      article.UpdatedAt,
			Title:          title,
			Abstract:       abstract,
			Content:        contentHead,
			Cover:          article.Cover,
			ViewCount:      article.ViewCount,
			DiggCount:      article.DiggCount,
			CommentCount:   article.CommentCount,
			FavorCount:     article.FavorCount,
			CommentsToggle: article.CommentsToggle,
			Status:         article.Status,
			Tags:           []string(article.TagList),
			UserTop:        isUserTop(topMap, article.ID),
			AdminTop:       isAdminTop(topMap, article.ID),
		})
	}

	articleIDs := make([]uint, 0, len(list))
	for _, item := range list {
		articleIDs = append(articleIDs, item.ID)
	}
	displayMetaMap := loadSearchArticleDisplayMetaMap(articleIDs)
	favorMap, diggMap, viewMap, commentMap := loadSearchArticleCounterMaps(articleIDs)
	for index := range list {
		list[index].FavorCount += favorMap[list[index].ID]
		list[index].DiggCount += diggMap[list[index].ID]
		list[index].ViewCount += viewMap[list[index].ID]
		list[index].CommentCount += commentMap[list[index].ID]
		list[index].CategoryTitle = displayMetaMap[list[index].ID].CategoryTitle
		list[index].UserNickname = displayMetaMap[list[index].ID].UserNickname
		list[index].UserAvatar = displayMetaMap[list[index].ID].UserAvatar
	}

	return
}
