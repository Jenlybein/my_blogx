package flags

import (
	"bufio"
	"encoding/json"
	"io"
	"myblogx/conf"
	"myblogx/global"
	"myblogx/models"
	"myblogx/test/testutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/elastic/go-elasticsearch/v7"
)

func setupMockESClient(t *testing.T, handler http.HandlerFunc) {
	t.Helper()

	server := httptest.NewServer(handler)
	client, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{server.URL},
	})
	if err != nil {
		t.Fatalf("创建 mock ES 客户端失败: %v", err)
	}

	oldClient := global.ESClient
	global.ESClient = client
	t.Cleanup(func() {
		global.ESClient = oldClient
		server.Close()
	})
}

func writeESJSON(w http.ResponseWriter, code int, body string) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Elastic-Product", "Elasticsearch")
	w.WriteHeader(code)
	_, _ = io.WriteString(w, body)
}

func TestBuildArticleESDocument(t *testing.T) {
	categoryID := uint(7)
	article := models.ArticleModel{
		Model: models.Model{
			ID:        1,
			CreatedAt: time.Unix(1710000000, 0),
			UpdatedAt: time.Unix(1710003600, 0),
		},
		Title:          "文章标题",
		Abstract:       "文章摘要",
		Content:        "不会同步到 ES",
		HtmlContent:    "<p>正文</p>",
		CategoryID:     &categoryID,
		Cover:          "/cover.png",
		AuthorID:       9,
		ViewCount:      11,
		DiggCount:      12,
		CommentCount:   13,
		FavorCount:     14,
		CommentsToggle: true,
		Status:         3,
		TagList:        []string{"Go", "Redis"},
	}

	doc := buildArticleESDocument(article)

	if _, ok := doc["content"]; ok {
		t.Fatal("content 不应被同步到 ES 文档")
	}
	if got, ok := doc["comments_toggle"].(int); !ok || got != 1 {
		t.Fatalf("comments_toggle 应按 integer mapping 转成 1, got=%#v", doc["comments_toggle"])
	}
	tagList, ok := doc["tag_list"].([]string)
	if !ok || len(tagList) != 2 || tagList[0] != "Go" || tagList[1] != "Redis" {
		t.Fatalf("tag_list 同步结果不正确: %#v", doc["tag_list"])
	}
	if len(doc) != len(articleESSyncColumns) {
		t.Fatalf("ES 文档字段数不正确, got=%d want=%d", len(doc), len(articleESSyncColumns))
	}
}

func TestSyncArticleDocuments(t *testing.T) {
	db := testutil.SetupSQLite(t, &models.ArticleModel{})
	testutil.InitGlobals()
	global.Config = &conf.Config{
		ES:    conf.ES{Index: "article_index"},
		River: conf.River{BulkSize: 2},
	}

	articles := []models.ArticleModel{
		{
			Title:          "第一篇",
			Abstract:       "摘要1",
			HtmlContent:    "<p>a</p>",
			AuthorID:       1,
			ViewCount:      10,
			DiggCount:      2,
			CommentCount:   3,
			FavorCount:     4,
			CommentsToggle: true,
			Status:         3,
			TagList:        []string{"Go", "后端"},
		},
		{
			Title:          "第二篇",
			Abstract:       "摘要2",
			HtmlContent:    "<p>b</p>",
			AuthorID:       2,
			ViewCount:      20,
			DiggCount:      5,
			CommentCount:   6,
			FavorCount:     7,
			CommentsToggle: false,
			Status:         2,
			TagList:        []string{"Redis"},
		},
	}
	if err := db.Create(&articles).Error; err != nil {
		t.Fatalf("创建测试文章失败: %v", err)
	}

	var bulkDocs []map[string]any
	setupMockESClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/":
			writeESJSON(w, 200, `{"name":"mock","cluster_name":"mock","version":{"number":"7.17.10"},"tagline":"You Know, for Search"}`)
		case r.Method == http.MethodHead && r.URL.Path == "/article_index":
			w.Header().Set("X-Elastic-Product", "Elasticsearch")
			w.WriteHeader(200)
		case r.Method == http.MethodPost && r.URL.Path == "/article_index/_bulk":
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
			writeESJSON(w, 200, `{"took":1,"errors":false,"items":[]}`)
		default:
			writeESJSON(w, 404, `{"error":{"reason":"not found"}}`)
		}
	})

	total, err := syncArticleDocuments(db, "article_index", 2)
	if err != nil {
		t.Fatalf("同步文章到 ES 失败: %v", err)
	}
	if total != 2 {
		t.Fatalf("同步文章数量不正确: got=%d want=2", total)
	}
	if len(bulkDocs) != 2 {
		t.Fatalf("写入 ES 的文档数不正确: got=%d want=2", len(bulkDocs))
	}

	first := bulkDocs[0]
	if _, ok := first["content"]; ok {
		t.Fatal("bulk 文档中不应包含 content 字段")
	}
	if got, ok := first["comments_toggle"].(float64); !ok || got != 1 {
		t.Fatalf("comments_toggle 应写成 1, got=%#v", first["comments_toggle"])
	}
	tagList, ok := first["tag_list"].([]any)
	if !ok || len(tagList) != 2 {
		t.Fatalf("tag_list 应以数组写入 ES, got=%#v", first["tag_list"])
	}
}
