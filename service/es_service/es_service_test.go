package es_service

import (
	"bufio"
	"encoding/json"
	"io"
	"myblogx/conf"
	confsite "myblogx/conf/site"
	"myblogx/global"
	"myblogx/models"
	"myblogx/models/ctype"
	"myblogx/models/enum"
	"myblogx/test/testutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
)

func setupMockES(t *testing.T, handler http.HandlerFunc) {
	t.Helper()
	srv := httptest.NewServer(handler)
	client, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{srv.URL},
	})
	if err != nil {
		t.Fatalf("创建 mock ES 客户端失败: %v", err)
	}

	old := global.ESClient
	global.ESClient = client
	t.Cleanup(func() {
		global.ESClient = old
		srv.Close()
	})
}

func writeJSON(w http.ResponseWriter, code int, body string) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Elastic-Product", "Elasticsearch")
	w.WriteHeader(code)
	_, _ = io.WriteString(w, body)
}

func TestIndexAndPipelineOps(t *testing.T) {
	indexExists := map[string]bool{"idx1": true}
	pipelineExists := map[string]bool{"p1": true}

	setupMockES(t, func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if path == "/" {
			writeJSON(w, 200, `{"name":"mock","cluster_name":"mock","version":{"number":"7.17.10"},"tagline":"You Know, for Search"}`)
			return
		}
		switch {
		case strings.HasPrefix(path, "/_ingest/pipeline/"):
			id := strings.TrimPrefix(path, "/_ingest/pipeline/")
			switch r.Method {
			case http.MethodGet:
				if pipelineExists[id] {
					writeJSON(w, 200, `{}`)
				} else {
					writeJSON(w, 404, `{"error":{"reason":"not found"}}`)
				}
			case http.MethodDelete:
				delete(pipelineExists, id)
				writeJSON(w, 200, `{}`)
			case http.MethodPut:
				pipelineExists[id] = true
				writeJSON(w, 200, `{}`)
			default:
				writeJSON(w, 500, `{"error":{"reason":"bad method"}}`)
			}
			return
		default:
			index := strings.TrimPrefix(path, "/")
			switch r.Method {
			case http.MethodHead:
				if indexExists[index] {
					w.WriteHeader(200)
				} else {
					w.WriteHeader(404)
				}
			case http.MethodDelete:
				delete(indexExists, index)
				writeJSON(w, 200, `{}`)
			case http.MethodPut:
				if strings.HasSuffix(path, "/_mapping") {
					writeJSON(w, 200, `{}`)
					return
				}
				indexExists[index] = true
				writeJSON(w, 200, `{}`)
			case http.MethodGet:
				if strings.HasSuffix(path, "/_mapping") {
					writeJSON(w, 200, `{"idx1":{"mappings":{"properties":{}}}}`)
					return
				}
				writeJSON(w, 500, `{"error":{"reason":"unexpected"}}`)
			default:
				writeJSON(w, 500, `{"error":{"reason":"bad method"}}`)
			}
		}
	})

	if err := CreateIndexForce("idx1", `{}`); err != nil {
		t.Fatalf("CreateIndexForce 失败: %v", err)
	}
	if exists, err := ExistsIndex("idx1"); err != nil || !exists {
		t.Fatalf("ExistsIndex 结果异常: exists=%v err=%v", exists, err)
	}
	if err := DeleteIndex("idx1"); err != nil {
		t.Fatalf("DeleteIndex 失败: %v", err)
	}
	if exists, err := ExistsIndex("idx1"); err != nil || exists {
		t.Fatalf("Delete 后 ExistsIndex 异常: exists=%v err=%v", exists, err)
	}

	if err := CreatePipelineForce("p1", `{}`); err != nil {
		t.Fatalf("CreatePipelineForce 失败: %v", err)
	}
	if exists, err := ExistsPipeline("p1"); err != nil || !exists {
		t.Fatalf("ExistsPipeline 结果异常: exists=%v err=%v", exists, err)
	}
	if err := DeletePipeline("p1"); err != nil {
		t.Fatalf("DeletePipeline 失败: %v", err)
	}
	if exists, err := ExistsPipeline("p1"); err != nil || exists {
		t.Fatalf("Delete 后 ExistsPipeline 异常: exists=%v err=%v", exists, err)
	}
}

func TestDocumentAndBulkOps(t *testing.T) {
	setupMockES(t, func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if path == "/" {
			writeJSON(w, 200, `{"name":"mock","cluster_name":"mock","version":{"number":"7.17.10"},"tagline":"You Know, for Search"}`)
			return
		}
		switch {
		case strings.Contains(path, "/_search"):
			writeJSON(w, 200, `{"hits":{"total":{"value":2},"hits":[{"_source":{"id":1,"title":"a"}},{"_source":{"id":2,"title":"b"}}]}}`)
		case strings.Contains(path, "/_update/"):
			writeJSON(w, 200, `{"result":"updated"}`)
		case r.Method == http.MethodDelete && strings.Contains(path, "/_doc/"):
			writeJSON(w, 200, `{"result":"deleted"}`)
		case r.Method == http.MethodPost && strings.Contains(path, "/_doc"):
			writeJSON(w, 200, `{"result":"created"}`)
		case r.Method == http.MethodGet && strings.Contains(path, "/_doc/"):
			writeJSON(w, 200, `{"_id":"1","found":true}`)
		case r.Method == http.MethodHead && strings.Contains(path, "/_doc/"):
			w.WriteHeader(200)
		case path == "/_bulk" || strings.HasSuffix(path, "/_bulk"):
			writeJSON(w, 200, `{"errors":false,"items":[]}`)
		case strings.HasSuffix(path, "/_mapping") && r.Method == http.MethodGet:
			writeJSON(w, 200, `{"idx":{"mappings":{"properties":{}}}}`)
		case strings.HasSuffix(path, "/_mapping") && r.Method == http.MethodPut:
			writeJSON(w, 200, `{}`)
		case r.Method == http.MethodHead:
			w.WriteHeader(404)
		case r.Method == http.MethodPut:
			writeJSON(w, 200, `{}`)
		case r.Method == http.MethodDelete:
			writeJSON(w, 200, `{"acknowledged":true}`)
		default:
			writeJSON(w, 500, `{"error":{"reason":"unexpected"}}`)
		}
	})

	if resp := CreateDocument("idx", map[string]any{"title": "x"}); !resp.Success {
		t.Fatalf("CreateDocument 失败: %+v", resp)
	}
	if resp := Search[map[string]any]("idx", 1, 10, map[string]any{"match_all": map[string]any{}}); !resp.Success {
		t.Fatalf("Search 失败: %+v", resp)
	}
	if resp := UpdateDocument("idx", "1", map[string]any{"title": "y"}); !resp.Success {
		t.Fatalf("UpdateDocument 失败: %+v", resp)
	}
	if resp := DeleteDocument("idx", "1"); !resp.Success {
		t.Fatalf("DeleteDocument 失败: %+v", resp)
	}
	if resp := Get("idx", "_doc", "1"); !resp.Success {
		t.Fatalf("Get 失败: %+v", resp)
	}
	if resp := Exists("idx", "_doc", "1"); !resp.Success || resp.Data != true {
		t.Fatalf("Exists 结果异常: %+v", resp)
	}

	items := []*BulkRequest{
		{Action: ActionIndex, Index: "idx", ID: "1", Data: map[string]interface{}{"k": "v"}},
	}
	if resp := Bulk(items); !resp.Success {
		t.Fatalf("Bulk 失败: %+v", resp)
	}
	if resp := IndexBulk("idx", items); !resp.Success {
		t.Fatalf("IndexBulk 失败: %+v", resp)
	}
	if resp := IndexTypeBulk("idx", "_doc", items); !resp.Success {
		t.Fatalf("IndexTypeBulk 失败: %+v", resp)
	}

	if resp := CreateMapping("idx", "_doc", map[string]interface{}{"title": map[string]any{"type": "text"}}); !resp.Success {
		t.Fatalf("CreateMapping 失败: %+v", resp)
	}
	if resp := GetMapping("idx", "_doc"); !resp.Success {
		t.Fatalf("GetMapping 失败: %+v", resp)
	}
	if resp := DeleteIndexWithResponse("idx"); !resp.Success {
		t.Fatalf("DeleteIndexWithResponse 失败: %+v", resp)
	}
}

func TestUpdateESDocsContent(t *testing.T) {
	db := testutil.SetupSQLite(t,
		&models.UserModel{},
		&models.UserConfModel{},
		&models.ArticleModel{},
		&models.TagModel{},
		&models.ArticleTagModel{},
		&models.UserTopArticleModel{},
	)
	global.Config = &conf.Config{
		ES: conf.ES{Index: "article_index"},
		Site: conf.Site{
			SiteInfo: confsite.SiteInfo{Mode: enum.SiteModeCommunity},
		},
	}

	user := models.UserModel{Username: "author", Password: "x", Role: enum.RoleUser}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("创建用户失败: %v", err)
	}
	article := models.ArticleModel{
		Title:    "t1",
		Abstract: "摘要",
		Content:  "# 新标题\n新正文\n[错误链接](### 新标题)",
		AuthorID: user.ID,
		Status:   enum.ArticleStatusExamining,
	}
	if err := db.Create(&article).Error; err != nil {
		t.Fatalf("创建文章失败: %v", err)
	}

	var bulkDocs []map[string]any
	setupMockES(t, func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		switch {
		case path == "/":
			writeJSON(w, 200, `{"name":"mock","cluster_name":"mock","version":{"number":"7.17.10"},"tagline":"You Know, for Search"}`)
		case path == "/_bulk" || strings.HasSuffix(path, "/_bulk"):
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("读取 bulk body 失败: %v", err)
			}
			scanner := bufio.NewScanner(strings.NewReader(string(body)))
			lineNo := 0
			for scanner.Scan() {
				lineNo++
				line := scanner.Bytes()
				if len(strings.TrimSpace(string(line))) == 0 {
					continue
				}
				if lineNo%2 == 0 {
					var doc map[string]any
					if err = json.Unmarshal(line, &doc); err != nil {
						t.Fatalf("解析 bulk 文档失败: %v", err)
					}
					bulkDocs = append(bulkDocs, doc)
				}
			}
			writeJSON(w, 200, `{"took":1,"errors":false,"items":[]}`)
		default:
			writeJSON(w, 404, `{"error":{"reason":"not found"}}`)
		}
	})

	if err := UpdateESDocsContent([]ctype.ID{article.ID}); err != nil {
		t.Fatalf("UpdateESDocsContent 失败: %v", err)
	}
	if len(bulkDocs) != 1 {
		t.Fatalf("应重建一篇 ES 文档, got=%d", len(bulkDocs))
	}

	docMap, ok := bulkDocs[0]["doc"].(map[string]any)
	if !ok {
		t.Fatalf("应使用 partial update 文档结构, got=%#v", bulkDocs[0])
	}
	parts, ok := docMap["content_parts"].([]any)
	if !ok || len(parts) != 1 {
		t.Fatalf("content_parts 同步错误: %#v", docMap["content_parts"])
	}
	firstPart, ok := parts[0].(map[string]any)
	if !ok {
		t.Fatalf("content_parts 首段结构错误: %#v", parts[0])
	}
	content, _ := firstPart["content"].(string)
	if !strings.Contains(content, "新标题") || !strings.Contains(content, "新正文") {
		t.Fatalf("content_parts 未同步最新正文: %q", content)
	}
	if strings.Contains(content, "](### ") {
		t.Fatalf("content_parts 不应保留错误链接语法: %q", content)
	}
	if _, ok = docMap["tags"]; ok {
		t.Fatalf("UpdateESDocsContent 不应更新 tags: %#v", docMap)
	}
}

func TestUpdateESDocsTags(t *testing.T) {
	db := testutil.SetupSQLite(t,
		&models.UserModel{},
		&models.UserConfModel{},
		&models.ArticleModel{},
		&models.TagModel{},
		&models.ArticleTagModel{},
	)
	global.Config = &conf.Config{ES: conf.ES{Index: "article_index"}}

	user := models.UserModel{Username: "author", Password: "x", Role: enum.RoleUser}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("创建用户失败: %v", err)
	}
	tagGo := models.TagModel{Title: "Go", IsEnabled: true}
	tagES := models.TagModel{Title: "ES", IsEnabled: true}
	if err := db.Create(&tagGo).Error; err != nil {
		t.Fatalf("创建标签失败: %v", err)
	}
	if err := db.Create(&tagES).Error; err != nil {
		t.Fatalf("创建标签失败: %v", err)
	}
	article := models.ArticleModel{Title: "t1", Content: "正文", AuthorID: user.ID, Status: enum.ArticleStatusExamining}
	if err := db.Create(&article).Error; err != nil {
		t.Fatalf("创建文章失败: %v", err)
	}
	if err := db.Model(&article).Association("Tags").Replace([]models.TagModel{tagGo, tagES}); err != nil {
		t.Fatalf("写入文章标签失败: %v", err)
	}

	var bulkDocs []map[string]any
	setupMockES(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/":
			writeJSON(w, 200, `{"name":"mock","cluster_name":"mock","version":{"number":"7.17.10"},"tagline":"You Know, for Search"}`)
		case r.URL.Path == "/_bulk" || strings.HasSuffix(r.URL.Path, "/_bulk"):
			body, _ := io.ReadAll(r.Body)
			scanner := bufio.NewScanner(strings.NewReader(string(body)))
			lineNo := 0
			for scanner.Scan() {
				lineNo++
				line := scanner.Bytes()
				if len(strings.TrimSpace(string(line))) == 0 {
					continue
				}
				if lineNo%2 == 0 {
					var doc map[string]any
					if err := json.Unmarshal(line, &doc); err != nil {
						t.Fatalf("解析 bulk 文档失败: %v", err)
					}
					bulkDocs = append(bulkDocs, doc)
				}
			}
			writeJSON(w, 200, `{"took":1,"errors":false,"items":[]}`)
		default:
			writeJSON(w, 404, `{"error":{"reason":"not found"}}`)
		}
	})

	if err := UpdateESDocsTags([]ctype.ID{article.ID}); err != nil {
		t.Fatalf("UpdateESDocsTags 失败: %v", err)
	}
	if len(bulkDocs) != 1 {
		t.Fatalf("应更新一篇 ES 文档, got=%d", len(bulkDocs))
	}
	docMap, ok := bulkDocs[0]["doc"].(map[string]any)
	if !ok {
		t.Fatalf("应使用 partial update 文档结构, got=%#v", bulkDocs[0])
	}
	tags, ok := docMap["tags"].([]any)
	if !ok || len(tags) != 2 {
		t.Fatalf("tags 更新错误: %#v", docMap["tags"])
	}
	if _, ok = docMap["content_parts"]; ok {
		t.Fatalf("UpdateESDocsTags 不应更新正文字段: %#v", docMap)
	}
}

func TestUpdateESDocsTop(t *testing.T) {
	db := testutil.SetupSQLite(t,
		&models.UserModel{},
		&models.UserConfModel{},
		&models.ArticleModel{},
		&models.UserTopArticleModel{},
	)
	global.Config = &conf.Config{ES: conf.ES{Index: "article_index"}}

	admin := models.UserModel{Username: "admin", Password: "x", Role: enum.RoleAdmin}
	author := models.UserModel{Username: "author", Password: "x", Role: enum.RoleUser}
	if err := db.Create(&admin).Error; err != nil {
		t.Fatalf("创建管理员失败: %v", err)
	}
	if err := db.Create(&author).Error; err != nil {
		t.Fatalf("创建作者失败: %v", err)
	}
	article := models.ArticleModel{Title: "t1", Content: "正文", AuthorID: author.ID, Status: enum.ArticleStatusExamining}
	if err := db.Create(&article).Error; err != nil {
		t.Fatalf("创建文章失败: %v", err)
	}
	if err := db.Create(&models.UserTopArticleModel{UserID: admin.ID, ArticleID: article.ID}).Error; err != nil {
		t.Fatalf("创建管理员置顶失败: %v", err)
	}
	if err := db.Create(&models.UserTopArticleModel{UserID: author.ID, ArticleID: article.ID}).Error; err != nil {
		t.Fatalf("创建作者置顶失败: %v", err)
	}

	var bulkDocs []map[string]any
	setupMockES(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/":
			writeJSON(w, 200, `{"name":"mock","cluster_name":"mock","version":{"number":"7.17.10"},"tagline":"You Know, for Search"}`)
		case r.URL.Path == "/_bulk" || strings.HasSuffix(r.URL.Path, "/_bulk"):
			body, _ := io.ReadAll(r.Body)
			scanner := bufio.NewScanner(strings.NewReader(string(body)))
			lineNo := 0
			for scanner.Scan() {
				lineNo++
				line := scanner.Bytes()
				if len(strings.TrimSpace(string(line))) == 0 {
					continue
				}
				if lineNo%2 == 0 {
					var doc map[string]any
					if err := json.Unmarshal(line, &doc); err != nil {
						t.Fatalf("解析 bulk 文档失败: %v", err)
					}
					bulkDocs = append(bulkDocs, doc)
				}
			}
			writeJSON(w, 200, `{"took":1,"errors":false,"items":[]}`)
		default:
			writeJSON(w, 404, `{"error":{"reason":"not found"}}`)
		}
	})

	if err := UpdateESDocsTop([]ctype.ID{article.ID}); err != nil {
		t.Fatalf("UpdateESDocsTop 失败: %v", err)
	}
	if len(bulkDocs) != 1 {
		t.Fatalf("应更新一篇 ES 文档, got=%d", len(bulkDocs))
	}
	docMap, ok := bulkDocs[0]["doc"].(map[string]any)
	if !ok {
		t.Fatalf("应使用 partial update 文档结构, got=%#v", bulkDocs[0])
	}
	if docMap["admin_top"] != true || docMap["author_top"] != true {
		t.Fatalf("置顶字段更新错误: %#v", docMap)
	}
	if _, ok = docMap["tags"]; ok {
		t.Fatalf("UpdateESDocsTop 不应更新其他字段: %#v", docMap)
	}
}

func TestDecodeResponseAndHandleErrorFallback(t *testing.T) {
	data, err := decodeResponse(io.NopCloser(strings.NewReader(`{"a":1}`)))
	if err != nil {
		t.Fatalf("decodeResponse 失败: %v", err)
	}
	if v, ok := data["a"].(float64); !ok || v != 1 {
		t.Fatalf("decodeResponse 结果异常: %#v", data)
	}

	esRes := &esapi.Response{
		StatusCode: 403,
		Body:       io.NopCloser(strings.NewReader(`{"error":{}}`)),
	}
	err = handleError(esRes)
	if err == nil || !strings.Contains(err.Error(), "权限不足") {
		t.Fatalf("handleError 兜底信息异常: %v", err)
	}
}

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

	articles := ExtractArticles(input)
	if len(articles) != 1 {
		t.Fatalf("数量错误: %d", len(articles))
	}
	if articles[0].ID != ctype.ID(1) || articles[0].Title != "title-1" {
		t.Fatalf("解析结果异常: %+v", articles[0])
	}
}

func TestExtractArticlesMoreFields(t *testing.T) {
	src := map[string]any{
		"hits": []any{
			map[string]any{
				"_source": map[string]any{
					"id":              3,
					"title":           "t3",
					"comments_toggle": true,
				},
			},
		},
	}
	arts := ExtractArticles(src)
	if len(arts) != 1 || arts[0].ID != 3 || arts[0].Title != "t3" || !arts[0].CommentsToggle {
		b, _ := json.Marshal(arts)
		t.Fatalf("ExtractArticles 结果异常: %s", string(b))
	}
}

func TestBuildBulkBody(t *testing.T) {
	items := []*BulkRequest{
		{
			Action: ActionIndex,
			Index:  "idx",
			ID:     "1",
			Data: map[string]interface{}{
				"title": "hello",
			},
		},
		{
			Action: ActionUpdate,
			Index:  "idx",
			ID:     "2",
			Data: map[string]interface{}{
				"title": "world",
			},
		},
		{
			Action: ActionDelete,
			Index:  "idx",
			ID:     "3",
		},
	}

	body, err := buildBulkBody(items)
	if err != nil {
		t.Fatalf("buildBulkBody 失败: %v", err)
	}
	s := string(body)
	if !strings.Contains(s, "\"index\"") || !strings.Contains(s, "\"update\"") || !strings.Contains(s, "\"delete\"") {
		t.Fatalf("bulk body 缺少 action: %s", s)
	}
	if !strings.Contains(s, "\"doc\"") {
		t.Fatalf("update 文档结构缺失: %s", s)
	}
}

func TestHandleError(t *testing.T) {
	res := &esapi.Response{
		StatusCode: 400,
		Body: io.NopCloser(strings.NewReader(
			`{"error":{"reason":"bad request","caused_by":{"reason":"x"}}}`,
		)),
	}
	err := handleError(res)
	if err == nil {
		t.Fatal("handleError 应返回错误")
	}
	if !strings.Contains(err.Error(), "bad request") {
		t.Fatalf("错误信息异常: %v", err)
	}
}

func TestCloseResponse(t *testing.T) {
	res := &esapi.Response{
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader(`{}`)),
	}
	closeResponse(res)
}
