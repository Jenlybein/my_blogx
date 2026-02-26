package es_service_test

import (
	"myblogx/service/es_service"
	"testing"
)

func TestExtractArticles(t *testing.T) {
	input := map[string]any{
		"hits": []any{
			map[string]any{
				"_source": map[string]any{
					"id":    1,
					"title": "title-1",
				},
			},
		},
	}

	articles := es_service.ExtractArticles(input)
	if len(articles) != 1 {
		t.Fatalf("数量错误: %d", len(articles))
	}
	if articles[0].ID != uint(1) || articles[0].Title != "title-1" {
		t.Fatalf("解析结果异常: %+v", articles[0])
	}
}
