package search_api

import (
	"myblogx/global"
	"myblogx/models"
	"myblogx/models/enum"
	"strings"
)

// buildDefaultArticleSearchQuery 构建默认文章搜索查询
func buildDefaultArticleSearchQuery(key string) map[string]any {
	return buildArticleSearchQuery(key, map[string]any{
		"filter": []any{
			map[string]any{
				"term": map[string]any{
					"status": enum.ArticleStatusPublished,
				},
			},
		},
	})
}

// buildSelfArticleSearchQuery 构建“我的文章”搜索查询。
// 默认查询当前用户除已删除外的全部文章；如果显式传入状态，则按指定状态精确筛选。
func buildSelfArticleSearchQuery(key string, userID uint, status enum.ArticleStatus) map[string]any {
	boolQuery := map[string]any{
		"filter": []any{
			map[string]any{
				"term": map[string]any{
					"author_id": userID,
				},
			},
		},
	}

	if status != 0 {
		filters, _ := boolQuery["filter"].([]any)
		boolQuery["filter"] = append(filters, map[string]any{
			"term": map[string]any{
				"status": status,
			},
		})
	} else {
		boolQuery["must_not"] = []any{
			map[string]any{
				"term": map[string]any{
					"status": enum.ArticleStatusDeleted,
				},
			},
		}
	}

	return buildArticleSearchQuery(key, boolQuery)
}

// buildAdminArticleSearchQuery 构建管理员文章搜索查询。
// 管理员默认可以搜索任意状态的文章；如果显式传入状态，则按指定状态精确筛选。
func buildAdminArticleSearchQuery(key string, status enum.ArticleStatus) map[string]any {
	boolQuery := map[string]any{}
	if status != 0 {
		boolQuery["filter"] = []any{
			map[string]any{
				"term": map[string]any{
					"status": status,
				},
			},
		}
	}
	return buildArticleSearchQuery(key, boolQuery)
}

// buildArticleSearchQuery 构建文章搜索查询。
// boolQuery 负责业务筛选条件，关键词匹配和综合评分在这里统一追加。
func buildArticleSearchQuery(key string, boolQuery map[string]any) map[string]any {
	key = strings.TrimSpace(key)

	if key != "" {
		boolQuery["must"] = []any{
			map[string]any{
				"multi_match": map[string]any{
					"query":  key,
					"fields": []string{"title", "abstract", "html_content"},
				},
			},
		}
	} else {
		boolQuery["must"] = []any{
			map[string]any{
				"match_all": map[string]any{},
			},
		}
	}

	return map[string]any{
		"function_score": map[string]any{
			"query": map[string]any{
				"bool": boolQuery,
			},
			"functions": []any{
				map[string]any{
					"gauss": map[string]any{
						"created_at": map[string]any{
							"origin": "now",
							"scale":  "30d",
							"offset": "7d",
							"decay":  0.5,
						},
					},
					"weight": 0.22,
				},
				map[string]any{
					"field_value_factor": map[string]any{
						"field":    "digg_count",
						"modifier": "log1p",
						"missing":  0,
					},
					"weight": 0.21,
				},
				map[string]any{
					"field_value_factor": map[string]any{
						"field":    "comment_count",
						"modifier": "log1p",
						"missing":  0,
					},
					"weight": 0.20,
				},
				map[string]any{
					"field_value_factor": map[string]any{
						"field":    "favor_count",
						"modifier": "log1p",
						"missing":  0,
					},
					"weight": 0.18,
				},
				map[string]any{
					"field_value_factor": map[string]any{
						"field":    "view_count",
						"modifier": "log1p",
						"missing":  0,
					},
					"weight": 0.12,
				},
			},
			"score_mode": "sum",
			"boost_mode": "sum",
		},
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

	boolQuery, ok := extractSearchBoolQuery(query)
	if !ok {
		return query
	}

	should, _ := boolQuery["should"].([]any)
	should = append(should, map[string]any{
		"terms": map[string]any{
			"tag_list": likeTagTitles,
			"boost":    2,
		},
	})
	boolQuery["should"] = should

	return query
}

// buildUserIDQuery 构建用户 ID 查询
func buildUserIDQuery(query map[string]any, userID uint) map[string]any {
	if userID == 0 {
		return query
	}

	boolQuery, ok := extractSearchBoolQuery(query)
	if !ok {
		return query
	}

	filters, _ := boolQuery["filter"].([]any)
	boolQuery["filter"] = append(filters, map[string]any{
		"term": map[string]any{
			"author_id": userID,
		},
	})
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

	boolQuery, ok := extractSearchBoolQuery(query)
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

// buildCategoryIDQuery 构建分类查询
func buildCategoryIDQuery(query map[string]any, categoryID uint) map[string]any {
	if categoryID == 0 {
		return query
	}

	boolQuery, ok := extractSearchBoolQuery(query)
	if !ok {
		return query
	}

	filters, _ := boolQuery["filter"].([]any)
	boolQuery["filter"] = append(filters, map[string]any{
		"term": map[string]any{
			"category_id": categoryID,
		},
	})
	return query
}

// buildArticleSearchExtraBody 构建文章搜索额外参数
func buildArticleSearchExtraBody(sortField string) map[string]any {
	sortList := []any{
		map[string]any{
			"_score": map[string]any{
				"order": "desc",
			},
		},
	}
	if sortField != "" {
		sortList = append(sortList, map[string]any{
			sortField: map[string]any{
				"order": "desc",
			},
		})
	}

	return map[string]any{
		"_source": []string{
			"id",
			"created_at",
			"updated_at",
			"title",
			"abstract",
			"content_head",
			"cover",
			"view_count",
			"digg_count",
			"comment_count",
			"favor_count",
			"comments_toggle",
			"status",
			"tag_list",
			"author_id",
			"category_id",
		},
		"sort": sortList,
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
