package ai_api_test

import (
	"encoding/json"
	"io"
	"myblogx/api/ai_api"
	"myblogx/conf"
	"myblogx/global"
	"myblogx/models"
	"myblogx/models/enum"
	"myblogx/service/ai_service"
	"myblogx/service/search_service"
	"myblogx/test/testutil"
	"myblogx/utils/jwts"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/elastic/go-elasticsearch/v7"
)

func TestAIArticleSearchView(t *testing.T) {
	db := testutil.SetupSQLite(
		t,
		&models.UserModel{},
		&models.UserConfModel{},
		&models.CategoryModel{},
		&models.ArticleModel{},
		&models.TagModel{},
	)

	user := models.UserModel{
		Username: "search_user",
		Password: "x",
		Nickname: "搜索用户",
		Avatar:   "/avatar.png",
		Role:     enum.RoleUser,
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("创建用户失败: %v", err)
	}

	category := models.CategoryModel{
		Title:  "Go 分类",
		UserID: user.ID,
	}
	if err := db.Create(&category).Error; err != nil {
		t.Fatalf("创建分类失败: %v", err)
	}

	if err := db.Create(&models.TagModel{Title: "Go", IsEnabled: true}).Error; err != nil {
		t.Fatalf("创建标签失败: %v", err)
	}
	if err := db.Create(&models.TagModel{Title: "Gin", IsEnabled: true}).Error; err != nil {
		t.Fatalf("创建标签失败: %v", err)
	}

	article1 := models.ArticleModel{
		Model:      models.Model{ID: 1},
		Title:      "Gin 中间件实践",
		Abstract:   "讲 Gin 中间件",
		Content:    "Gin 中间件正文",
		CategoryID: &category.ID,
		AuthorID:   user.ID,
		Status:     enum.ArticleStatusPublished,
	}
	if err := db.Create(&article1).Error; err != nil {
		t.Fatalf("创建文章1失败: %v", err)
	}

	article2 := models.ArticleModel{
		Model:      models.Model{ID: 2},
		Title:      "Go Web 基础",
		Abstract:   "讲 Go Web",
		Content:    "Go Web 正文",
		CategoryID: &category.ID,
		AuthorID:   user.ID,
		Status:     enum.ArticleStatusPublished,
	}
	if err := db.Create(&article2).Error; err != nil {
		t.Fatalf("创建文章2失败: %v", err)
	}

	aiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req ai_service.Request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("解析 AI 请求失败: %v", err)
		}
		if len(req.Messages) != 2 {
			t.Fatalf("AI 请求消息数量错误: %+v", req.Messages)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{
				{
					"index": 0,
					"message": map[string]any{
						"role": "assistant",
						"content": `{
							"intent":"search",
							"query":["Gin","中间件"],
							"tag_list":["Gin"],
							"sort":1
						}`,
					},
					"finish_reason": "stop",
				},
			},
		})
	}))
	defer aiServer.Close()

	esServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			writeMockESJSON(w, 200, `{"name":"mock","cluster_name":"mock","version":{"number":"7.17.10"},"tagline":"You Know, for Search"}`)
			return
		}

		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("读取 ES 请求失败: %v", err)
		}
		bodyText := string(bodyBytes)

		if strings.Contains(bodyText, `"tags.title"`) {
			writeMockESJSON(w, 200, `{
				"hits":{
					"total":{"value":1},
					"hits":[
						{"_source":{
							"id":1,
							"title":"Gin 中间件实践",
							"abstract":"讲 Gin 中间件",
							"cover":"",
							"view_count":1,
							"digg_count":2,
							"comment_count":3,
							"favor_count":4,
							"comments_toggle":true,
							"status":2,
							"tags":[{"title":"Gin"}],
							"author_top":false,
							"admin_top":false
						}}
					]
				}
			}`)
			return
		}

		writeMockESJSON(w, 200, `{
			"hits":{
				"total":{"value":2},
				"hits":[
					{"_source":{
						"id":1,
						"title":"Gin 中间件实践",
						"abstract":"讲 Gin 中间件",
						"cover":"",
						"view_count":1,
						"digg_count":2,
						"comment_count":3,
						"favor_count":4,
						"comments_toggle":true,
						"status":2,
						"tags":[{"title":"Gin"}],
						"author_top":false,
						"admin_top":false
					}},
					{"_source":{
						"id":2,
						"title":"Go Web 基础",
						"abstract":"讲 Go Web",
						"cover":"",
						"view_count":5,
						"digg_count":6,
						"comment_count":7,
						"favor_count":8,
						"comments_toggle":true,
						"status":2,
						"tags":[{"title":"Go"}],
						"author_top":false,
						"admin_top":false
					}}
				]
			}
		}`)
	}))
	defer esServer.Close()

	esClient, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{esServer.URL},
	})
	if err != nil {
		t.Fatalf("创建 ES 客户端失败: %v", err)
	}
	global.ESClient = esClient

	global.Config = &conf.Config{
		AI: conf.AI{
			Enable:        true,
			SecretKey:     "test-key",
			BaseURL:       aiServer.URL,
			ChatModel:     "test-model",
			MaxInputChars: 2048,
		},
		ES: conf.ES{
			Index: "article_index",
		},
	}

	api := ai_api.AIApi{}
	c, w := newAICtx()
	c.Set("claims", &jwts.MyClaims{
		Claims: jwts.Claims{
			UserID:   user.ID,
			Role:     enum.RoleUser,
			Username: user.Username,
		},
	})
	c.Set("requestJson", ai_api.AIArticleMetaInfoRequest{
		Content: "帮我找讲 Gin 中间件的文章",
	})

	api.AIArticleSearchView(c)

	if code := readAICode(t, w); code != 0 {
		t.Fatalf("AI 文章搜索应成功, body=%s", w.Body.String())
	}

	var body struct {
		Code int                                 `json:"code"`
		Data []search_service.SearchListResponse `json:"data"`
		Msg  string                              `json:"msg"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("解析响应失败: %v body=%s", err, w.Body.String())
	}

	if len(body.Data) != 2 {
		t.Fatalf("结果数量错误: %+v", body.Data)
	}
	if body.Data[0].ID != 1 || body.Data[1].ID != 2 {
		t.Fatalf("结果去重合并顺序错误: %+v", body.Data)
	}
}

func writeMockESJSON(w http.ResponseWriter, code int, body string) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Elastic-Product", "Elasticsearch")
	w.WriteHeader(code)
	_, _ = io.WriteString(w, body)
}
