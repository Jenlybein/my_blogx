package search_api

import (
	"encoding/json"
	"myblogx/global"
	"myblogx/models"
	"myblogx/models/enum"
	"myblogx/utils/markdown"
	"strings"
)

// buildDefaultArticleSearchQuery 构建默认文章搜索查询
func buildDefaultArticleSearchQuery(key string) map[string]any {
	// 只能查已发布的文章

	key = strings.TrimSpace(key)

	boolQuery := map[string]any{
		"filter": []any{
			map[string]any{
				"term": map[string]any{
					"status": enum.ArticleStatusPublished,
				},
			},
		},
	}

	if key != "" {
		boolQuery["must"] = []any{
			map[string]any{
				"multi_match": map[string]any{
					"query":  key,
					"fields": []string{"title", "abstract", "html_content"},
				},
			},
		}
	}

	return map[string]any{
		"bool": boolQuery,
	}
}

// buildLikeTagsQuery 构建喜欢标签查询
func buildLikeTagsQuery(query map[string]any, userID uint) map[string]any {
	var userConf models.UserConfModel
	if err := global.DB.Select("user_id", "like_tags").Take(&userConf, userID).Error; err != nil {
		return query
	}
	if len(userConf.LikeTags) == 0 {
		return query
	}

	var likeTagTitles []string
	if err := global.DB.Model(&models.TagModel{}).
		Where("id IN ? AND is_enabled = ?", userConf.LikeTags, true).
		Pluck("title", &likeTagTitles).Error; err != nil {
		return query
	}
	if len(likeTagTitles) == 0 {
		return query
	}

	boolQuery := query["bool"].(map[string]any)
	boolQuery["should"] = []any{
		map[string]any{
			"terms": map[string]any{
				"tag_list": likeTagTitles,
				"boost":    2,
			},
		},
	}

	return query
}

// buildTagListQuery 构建标签列表查询
func buildTagListQuery(query map[string]any, tagList []string) map[string]any {
	normalized := make([]string, 0, len(tagList))
	seen := make(map[string]struct{}, len(tagList))
	for _, item := range tagList {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		normalized = append(normalized, item)
	}
	if len(normalized) == 0 {
		return query
	}

	boolQuery, ok := query["bool"].(map[string]any)
	if !ok {
		return query
	}

	filters, _ := boolQuery["filter"].([]any)
	boolQuery["filter"] = append(filters, map[string]any{
		"terms": map[string]any{
			"tag_list": normalized,
		},
	})

	return query
}

// buildArticleSearchExtraBody 构建文章搜索额外参数
func buildArticleSearchExtraBody(sortField string) map[string]any {
	return map[string]any{
		"sort": []any{
			map[string]any{
				"_score": map[string]any{
					"order": "desc",
				},
			},
			map[string]any{
				sortField: map[string]any{
					"order": "desc",
				},
			},
		},
		// 高亮，用<em>标签包裹
		"highlight": map[string]any{
			"pre_tags":  []string{"<em>"},
			"post_tags": []string{"</em>"},
			"fields": map[string]any{
				"title": map[string]any{},
				"abstract": map[string]any{
					"number_of_fragments": 1,
				},
				"html_content": map[string]any{
					"fragment_size":       120,
					"number_of_fragments": 1,
				},
			},
		},
	}
}

// extractArticleSearchResults 提取文章搜索结果
func extractArticleSearchResults(data map[string]any) (list []SearchListResponse) {
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
		// TODO: 数据库增加一个字段用于展示文章前200个字，减少返回字数
		abstract := markdown.ExtractText(article.Abstract, 120)
		htmlContent := article.HtmlContent
		if len(highlightMap) > 0 {
			if values := extractHighlightValues(highlightMap, "title"); len(values) > 0 {
				title = values[0]
			}
			if values := extractHighlightValues(highlightMap, "abstract"); len(values) > 0 {
				abstract = values[0]
			}
			if values := extractHighlightValues(highlightMap, "html_content"); len(values) > 0 {
				htmlContent = values[0]
			}
		}

		list = append(list, SearchListResponse{
			ID:             article.ID,
			CreatedAt:      article.CreatedAt,
			UpdatedAt:      article.UpdatedAt,
			Title:          title,
			Abstract:       abstract,
			HtmlContent:    htmlContent,
			Cover:          article.Cover,
			ViewCount:      article.ViewCount,
			DiggCount:      article.DiggCount,
			CommentCount:   article.CommentCount,
			FavorCount:     article.FavorCount,
			CommentsToggle: article.CommentsToggle,
			Status:         article.Status,
			Tags:           []string(article.TagList),
		})
	}

	return
}

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
