package search_api

import "testing"

func TestExtractArticleSearchResults(t *testing.T) {
	data := map[string]any{
		"hits": []any{
			map[string]any{
				"_source": map[string]any{
					"id":       1,
					"title":    "go search",
					"abstract": "hello world",
				},
				"highlight": map[string]any{
					"title":        []any{"<em>go</em> search"},
					"html_content": []any{"prefix <em>go</em> suffix"},
				},
			},
		},
	}

	list := extractArticleSearchResults(data)
	if len(list) != 1 {
		t.Fatalf("结果数量错误: %d", len(list))
	}
	if list[0].ID != 1 || list[0].Title != "go search" {
		t.Fatalf("文章解析错误: %+v", list[0])
	}
	if len(list[0].Highlight["title"]) != 1 || list[0].Highlight["title"][0] != "<em>go</em> search" {
		t.Fatalf("标题高亮解析错误: %+v", list[0].Highlight)
	}
	if len(list[0].Highlight["html_content"]) != 1 {
		t.Fatalf("正文高亮解析错误: %+v", list[0].Highlight)
	}
}
