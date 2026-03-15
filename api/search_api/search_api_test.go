package search_api

import (
	"myblogx/models"
	"myblogx/models/enum"
	"myblogx/test/testutil"
	"testing"
)

func TestBuildLikeTagsQueryWithoutUserConf(t *testing.T) {
	testutil.SetupSQLite(t, &models.UserModel{}, &models.UserConfModel{}, &models.TagModel{})

	query := buildDefaultArticleSearchQuery("")
	query = buildLikeTagsQuery(query, 0)

	boolQuery, ok := query["bool"].(map[string]any)
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

	boolQuery, ok := query["bool"].(map[string]any)
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

func TestBuildArticleSearchExtraBody(t *testing.T) {
	extraBody := buildArticleSearchExtraBody("favor_count")
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
