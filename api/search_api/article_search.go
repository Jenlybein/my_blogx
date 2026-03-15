package search_api

import (
	"encoding/json"
	"myblogx/common/res"
	"myblogx/global"
	"myblogx/middleware"
	"myblogx/models"
	"myblogx/models/enum"
	"myblogx/service/es_service"
	"myblogx/utils/jwts"
	"strings"

	"github.com/gin-gonic/gin"
)

func (SearchApi) ArticleSearchView(c *gin.Context) {
	cr := middleware.GetBindQuery[ArticleSearchRequest](c)
	page := cr.Page
	if page <= 0 {
		page = 1
	}

	query := buildDefaultArticleSearchQuery(cr.Key)
	extraBody := buildArticleSearchExtraBody("created_at")

	switch cr.Type {
	case 0:
		claims, err := jwts.ParseTokenByGin(c)
		if err != nil || claims == nil {
			break
		}
		query = buildLikeTagsQuery(query, claims.UserID)
	case 1:
		extraBody = buildArticleSearchExtraBody("created_at")
	case 2:
		extraBody = buildArticleSearchExtraBody("comment_count")
	case 3:
		extraBody = buildArticleSearchExtraBody("digg_count")
	case 4:
		extraBody = buildArticleSearchExtraBody("favor_count")
	}

	resp := es_service.Search[map[string]any](
		models.ArticleModel{}.Index(),
		page,
		cr.GetLimit(),
		query,
		extraBody,
	)
	if !resp.Success {
		res.FailWithMsg(resp.Msg, c)
		return
	}

	data, ok := resp.Data.(map[string]any)
	if !ok {
		res.FailWithMsg("搜索结果格式错误", c)
		return
	}

	total, _ := data["total"].(float64)
	res.OkWithList(extractArticleSearchResults(data), int(total), c)
}

// buildDefaultArticleSearchQuery 构建默认文章搜索查询
func buildDefaultArticleSearchQuery(key string) map[string]any {
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
func extractArticleSearchResults(data map[string]any) (list []ArticleSearchResponse) {
	hits, _ := data["hits"].([]any)
	list = make([]ArticleSearchResponse, 0, len(hits))

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

		highlightMap, _ := item["highlight"].(map[string]any)
		var highlightResult map[string][]string
		if len(highlightMap) > 0 {
			highlightResult = make(map[string][]string, len(highlightMap))
			for field, rawList := range highlightMap {
				values, ok := rawList.([]any)
				if !ok {
					continue
				}
				for _, rawValue := range values {
					value, ok := rawValue.(string)
					if !ok {
						continue
					}
					highlightResult[field] = append(highlightResult[field], value)
				}
			}
			if len(highlightResult) == 0 {
				highlightResult = nil
			}
		}

		list = append(list, ArticleSearchResponse{
			ArticleModel: article,
			Highlight:    highlightResult,
		})
	}

	return
}
