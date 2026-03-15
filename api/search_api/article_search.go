package search_api

import (
	"encoding/json"
	"myblogx/common/res"
	"myblogx/middleware"
	"myblogx/models"
	"myblogx/service/es_service"
	"strings"

	"github.com/gin-gonic/gin"
)

func (SearchApi) ArticleSearchView(c *gin.Context) {
	cr := middleware.GetBindQuery[ArticleSearchRequest](c)
	page := cr.Page
	if page <= 0 {
		page = 1
	}

	resp := es_service.Search[map[string]any](
		models.ArticleModel{}.Index(),
		page,
		cr.GetLimit(),
		buildArticleSearchQuery(cr.Key),
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

// buildArticleSearchQuery 构建文章搜索查询
func buildArticleSearchQuery(key string) map[string]any {
	key = strings.TrimSpace(key)
	if key == "" {
		return map[string]any{
			"match_all": map[string]any{},
		}
	}

	return map[string]any{
		// 模糊搜索
		"bool": map[string]any{
			"must": []any{
				map[string]any{
					"multi_match": map[string]any{
						"query":  key,
						"fields": []string{"title", "abstract", "html_content"},
					},
				},
			},
		},
		// 高亮搜索结果
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

		// 获取文章内容结果
		var article models.ArticleModel
		if sourceMap, ok := item["_source"].(map[string]any); ok {
			jsonBytes, _ := json.Marshal(sourceMap)
			_ = json.Unmarshal(jsonBytes, &article)
		}

		// 获取文章高亮结果
		highlightMap, _ := item["highlight"].(map[string]any)
		if len(highlightMap) == 0 {
			return nil
		}

		highlightResult := make(map[string][]string, len(highlightMap))
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

		// 添加到返回列表
		list = append(list, ArticleSearchResponse{
			ArticleModel: article,
			Highlight:    highlightResult,
		})
	}

	return
}
