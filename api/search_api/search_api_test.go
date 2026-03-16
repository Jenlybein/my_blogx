package search_api

import (
	"myblogx/models"
	"myblogx/models/enum"
	"myblogx/test/testutil"
	"testing"
)

func TestBuildDefaultArticleSearchQueryOnlyPublished(t *testing.T) {
	query := buildDefaultArticleSearchQuery("golang")
	functionScore, ok := query["function_score"].(map[string]any)
	if !ok {
		t.Fatalf("function_score 查询结构错误: %#v", query)
	}
	queryBody, ok := functionScore["query"].(map[string]any)
	if !ok {
		t.Fatalf("function_score.query 结构错误: %#v", functionScore)
	}
	boolQuery, ok := queryBody["bool"].(map[string]any)
	if !ok {
		t.Fatalf("bool 查询结构错误: %#v", query)
	}

	filters, ok := boolQuery["filter"].([]any)
	if !ok || len(filters) != 1 {
		t.Fatalf("过滤条件异常: %#v", boolQuery["filter"])
	}

	term, ok := filters[0].(map[string]any)
	if !ok {
		t.Fatalf("term 过滤结构错误: %#v", filters[0])
	}
	statusTerm, ok := term["term"].(map[string]any)
	if !ok {
		t.Fatalf("status term 结构错误: %#v", term)
	}
	if statusTerm["status"] != enum.ArticleStatusPublished {
		t.Fatalf("搜索应只查询已发布文章，当前状态条件=%#v", statusTerm["status"])
	}

	functions, ok := functionScore["functions"].([]any)
	if !ok || len(functions) != 5 {
		t.Fatalf("综合评分函数异常: %#v", functionScore["functions"])
	}
	weights := []float64{0.22, 0.21, 0.20, 0.18, 0.12}
	for index, raw := range functions {
		item, ok := raw.(map[string]any)
		if !ok {
			t.Fatalf("评分函数结构错误: %#v", raw)
		}
		if item["weight"] != weights[index] {
			t.Fatalf("评分权重错误 index=%d weight=%#v", index, item["weight"])
		}
	}
}

func TestBuildLikeTagsQueryWithoutUserConf(t *testing.T) {
	testutil.SetupSQLite(t, &models.UserModel{}, &models.UserConfModel{}, &models.TagModel{})

	query := buildDefaultArticleSearchQuery("")
	query = buildLikeTagsQuery(query, 0)

	boolQuery, ok := extractSearchBoolQuery(query)
	if !ok {
		t.Fatalf("bool 查询结构错误: %#v", query)
	}
	if _, ok = boolQuery["should"]; ok {
		t.Fatalf("无有效用户配置时不应追加喜欢标签加权: %#v", boolQuery)
	}
}

func TestBuildLikeTagsQueryWithLikeTags(t *testing.T) {
	db := testutil.SetupSQLite(t, &models.UserModel{}, &models.UserConfModel{}, &models.TagModel{})

	user := models.UserModel{
		Username: "u1",
		Password: "x",
		Nickname: "nick",
		Role:     enum.RoleUser,
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("创建用户失败: %v", err)
	}

	tag := models.TagModel{Title: "Go", IsEnabled: true}
	if err := db.Create(&tag).Error; err != nil {
		t.Fatalf("创建标签失败: %v", err)
	}

	var userConf models.UserConfModel
	if err := db.Take(&userConf, "user_id = ?", user.ID).Error; err != nil {
		t.Fatalf("查询用户配置失败: %v", err)
	}
	if err := db.Model(&userConf).Updates(models.UserConfModel{
		LikeTags: []uint{tag.ID},
	}).Error; err != nil {
		t.Fatalf("更新偏好标签失败: %v", err)
	}

	query := buildDefaultArticleSearchQuery("golang")
	query = buildLikeTagsQuery(query, user.ID)

	boolQuery, ok := extractSearchBoolQuery(query)
	if !ok {
		t.Fatalf("bool 查询结构错误: %#v", query)
	}
	if _, ok = boolQuery["must"]; !ok {
		t.Fatalf("有关键词时应带 must 查询: %#v", boolQuery)
	}

	should, ok := boolQuery["should"].([]any)
	if !ok || len(should) != 1 {
		t.Fatalf("有偏好标签时应追加 tag 加权: %#v", boolQuery["should"])
	}
}

func TestBuildTagListQueryWithTags(t *testing.T) {
	query := buildDefaultArticleSearchQuery("golang")
	query = buildTagListQuery(query, []string{" Go ", "ES", "Go", ""})

	boolQuery, ok := extractSearchBoolQuery(query)
	if !ok {
		t.Fatalf("bool 查询结构错误: %#v", query)
	}

	filters, ok := boolQuery["filter"].([]any)
	if !ok || len(filters) != 2 {
		t.Fatalf("标签匹配过滤条件异常: %#v", boolQuery["filter"])
	}

	terms, ok := filters[1].(map[string]any)
	if !ok {
		t.Fatalf("标签 terms 结构错误: %#v", filters[1])
	}
	tagTerms, ok := terms["terms"].(map[string]any)
	if !ok {
		t.Fatalf("标签匹配结构错误: %#v", terms)
	}
	values, ok := tagTerms["tag_list"].([]string)
	if !ok || len(values) != 2 || values[0] != "Go" || values[1] != "ES" {
		t.Fatalf("标签名归一化异常: %#v", tagTerms["tag_list"])
	}
}

func TestBuildArticleSearchExtraBody(t *testing.T) {
	extraBody := buildArticleSearchExtraBody("view_count")
	sortList, ok := extraBody["sort"].([]any)
	if !ok || len(sortList) != 2 {
		t.Fatalf("排序条件异常: %#v", extraBody["sort"])
	}
}

func TestExtractArticleSearchResults(t *testing.T) {
	data := map[string]any{
		"hits": []any{
			map[string]any{
				"_source": map[string]any{
					"id":           1,
					"title":        "go search article",
					"abstract":     "hello world",
					"html_content": "origin html content",
					"tag_list":     []any{"Go", "ES"},
				},
				"highlight": map[string]any{
					"title":        []any{"<em>go</em> search"},
					"abstract":     []any{"<em>hello</em> world", "another <em>piece</em>"},
					"html_content": []any{"prefix <em>go</em> suffix", "second <em>segment</em>"},
				},
			},
		},
	}

	list := extractArticleSearchResults(data)
	if len(list) != 1 {
		t.Fatalf("结果数量错误: %d", len(list))
	}
	if list[0].ID != 1 || list[0].Title != "<em>go</em> search" {
		t.Fatalf("文章解析错误: %+v", list[0])
	}
	if list[0].Abstract != "<em>hello</em> world" {
		t.Fatalf("摘要高亮回填错误: %+v", list[0])
	}
	if list[0].HtmlContent != "prefix <em>go</em> suffix" {
		t.Fatalf("正文高亮回填错误: %+v", list[0])
	}
	if len(list[0].Tags) != 2 || list[0].Tags[0] != "Go" {
		t.Fatalf("标签解析错误: %+v", list[0].Tags)
	}
}
