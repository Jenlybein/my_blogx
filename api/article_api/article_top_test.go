package article_api

import (
	"myblogx/common"
	"myblogx/global"
	"myblogx/models"
	"myblogx/models/enum"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestArticleListViewMarksAdminTopAndSortsFirst(t *testing.T) {
	user := setupArticleEnv(t)
	db := global.DB

	admin := models.UserModel{Username: "article_admin", Password: "x", Role: enum.RoleAdmin}
	if err := db.Create(&admin).Error; err != nil {
		t.Fatalf("创建管理员失败: %v", err)
	}

	adminTopArticle := models.ArticleModel{
		Title:    "admin-top",
		Content:  "content",
		AuthorID: user.ID,
		Status:   enum.ArticleStatusPublished,
	}
	if err := db.Create(&adminTopArticle).Error; err != nil {
		t.Fatalf("创建管理员置顶文章失败: %v", err)
	}

	normalArticle := models.ArticleModel{
		Title:    "normal",
		Content:  "content",
		AuthorID: user.ID,
		Status:   enum.ArticleStatusPublished,
	}
	if err := db.Create(&normalArticle).Error; err != nil {
		t.Fatalf("创建普通文章失败: %v", err)
	}

	if err := db.Create(&models.UserTopArticleModel{UserID: admin.ID, ArticleID: adminTopArticle.ID}).Error; err != nil {
		t.Fatalf("创建管理员置顶关系失败: %v", err)
	}

	c, w := newCtx()
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.Set("requestQuery", ArticleListRequest{
		PageInfo: common.PageInfo{Page: 1, Limit: 10},
		Type:     1,
		Status:   enum.ArticleStatusPublished,
	})
	ArticleApi{}.ArticleListView(c)

	body := readBody(t, w)
	if code := int(body["code"].(float64)); code != 0 {
		t.Fatalf("查询文章列表失败, code=%d body=%s", code, w.Body.String())
	}

	data := body["data"].(map[string]any)
	list := data["list"].([]any)
	if len(list) != 2 {
		t.Fatalf("文章列表长度错误, got=%d", len(list))
	}

	first := list[0].(map[string]any)
	if got := first["id"].(string); got != adminTopArticle.ID.String() {
		t.Fatalf("管理员置顶文章应排在第一位, got=%s want=%s", got, adminTopArticle.ID.String())
	}
	if top, ok := first["admin_top"].(bool); !ok || !top {
		t.Fatalf("管理员置顶标记错误, got=%#v", first["admin_top"])
	}
}
