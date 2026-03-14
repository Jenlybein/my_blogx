package search_api

import (
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

	// TODO:增加 highline
	// TODO:检测是不是真的用得上 es_service.ExtractArticles ？

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
	res.OkWithList(es_service.ExtractArticles(data), int(total), c)
}

func buildArticleSearchQuery(key string) map[string]any {
	key = strings.TrimSpace(key)
	if key == "" {
		return map[string]any{
			"match_all": map[string]any{},
		}
	}

	return map[string]any{
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
	}
}
