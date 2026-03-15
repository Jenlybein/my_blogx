package tags

import (
	"encoding/json"
	"myblogx/common"
	"myblogx/conf"
	"myblogx/global"
	"myblogx/models"
	"myblogx/models/enum"
	"myblogx/test/testutil"
	"myblogx/utils/jwts"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func newCtx() (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	return c, w
}

func readCode(t *testing.T, w *httptest.ResponseRecorder) int {
	t.Helper()
	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("解析响应失败: %v", err)
	}
	return int(body["code"].(float64))
}

func setupTagEnv(t *testing.T) *models.UserModel {
	t.Helper()
	db := testutil.SetupSQLite(
		t,
		&models.UserModel{},
		&models.UserConfModel{},
		&models.TagModel{},
		&models.ArticleModel{},
		&models.ArticleTagModel{},
	)
	global.Config = &conf.Config{
		Jwt: conf.Jwt{
			Expire: 1,
			Secret: "tag-secret",
			Issuer: "tag-test",
		},
	}

	admin := &models.UserModel{
		Username: "admin_user",
		Password: "x",
		Role:     enum.RoleAdmin,
	}
	if err := db.Create(admin).Error; err != nil {
		t.Fatalf("创建管理员失败: %v", err)
	}
	return admin
}

func tokenForUser(t *testing.T, user *models.UserModel) string {
	t.Helper()
	token, err := jwts.GetToken(jwts.Claims{
		UserID:   user.ID,
		Role:     user.Role,
		Username: user.Username,
	})
	if err != nil {
		t.Fatalf("生成 token 失败: %v", err)
	}
	return token
}

func TestTagCRUDAndOptions(t *testing.T) {
	admin := setupTagEnv(t)
	api := TagsApi{}
	claims := &jwts.MyClaims{Claims: jwts.Claims{UserID: admin.ID, Role: enum.RoleAdmin, Username: admin.Username}}

	{
		c, w := newCtx()
		enabled := true
		c.Set("claims", claims)
		c.Set("requestJson", TagRequest{
			Title:     "Golang",
			Sort:      10,
			IsEnabled: &enabled,
		})
		api.TagCreateUpdateView(c)
		if code := readCode(t, w); code != 0 {
			t.Fatalf("创建标签失败, body=%s", w.Body.String())
		}
	}

	var tag models.TagModel
	if err := global.DB.Where("title = ?", "Golang").First(&tag).Error; err != nil {
		t.Fatalf("查询标签失败: %v", err)
	}

	{
		c, w := newCtx()
		c.Set("claims", claims)
		c.Set("requestQuery", TagListRequest{
			PageInfo: common.PageInfo{Page: 1, Limit: 10},
		})
		api.TagListView(c)
		if code := readCode(t, w); code != 0 {
			t.Fatalf("标签列表失败, body=%s", w.Body.String())
		}
	}

	{
		c, w := newCtx()
		disabled := false
		c.Set("claims", claims)
		c.Set("requestJson", TagRequest{
			ID:        tag.ID,
			Title:     "Go",
			IsEnabled: &disabled,
		})
		api.TagCreateUpdateView(c)
		if code := readCode(t, w); code != 0 {
			t.Fatalf("更新标签失败, body=%s", w.Body.String())
		}
	}

	token := tokenForUser(t, admin)
	{
		c, w := newCtx()
		req := httptest.NewRequest(http.MethodGet, "/articles/tags/options", nil)
		req.Header.Set("token", token)
		c.Request = req
		api.ArticleTagOptionsView(c)
		if code := readCode(t, w); code != 0 {
			t.Fatalf("标签选项失败, body=%s", w.Body.String())
		}
	}

	{
		c, w := newCtx()
		c.Set("requestJson", models.IDListRequest{IDList: []uint{tag.ID}})
		api.TagDeleteView(c)
		if code := readCode(t, w); code != 0 {
			t.Fatalf("删除标签失败, body=%s", w.Body.String())
		}
	}
}

func TestTagUpdateSyncsArticleTagList(t *testing.T) {
	admin := setupTagEnv(t)
	api := TagsApi{}
	claims := &jwts.MyClaims{Claims: jwts.Claims{UserID: admin.ID, Role: enum.RoleAdmin, Username: admin.Username}}

	tag := models.TagModel{Title: "Golang", IsEnabled: true}
	if err := global.DB.Create(&tag).Error; err != nil {
		t.Fatalf("创建标签失败: %v", err)
	}

	article := models.ArticleModel{
		Title:    "article-with-tag",
		Content:  "content",
		AuthorID: admin.ID,
		Status:   enum.ArticleStatusPublished,
		TagList:  []string{"Golang"},
	}
	if err := global.DB.Create(&article).Error; err != nil {
		t.Fatalf("创建文章失败: %v", err)
	}
	if err := global.DB.Create(&models.ArticleTagModel{ArticleID: article.ID, TagID: tag.ID}).Error; err != nil {
		t.Fatalf("创建文章标签关系失败: %v", err)
	}

	{
		c, w := newCtx()
		c.Set("claims", claims)
		c.Set("requestJson", TagRequest{
			ID:    tag.ID,
			Title: "Go",
		})
		api.TagCreateUpdateView(c)
		if code := readCode(t, w); code != 0 {
			t.Fatalf("更新标签失败, body=%s", w.Body.String())
		}
	}

	var updated models.ArticleModel
	if err := global.DB.Take(&updated, article.ID).Error; err != nil {
		t.Fatalf("回查文章失败: %v", err)
	}
	if len(updated.TagList) != 1 || updated.TagList[0] != "Go" {
		t.Fatalf("标签改名后文章 tag_list 未同步: %+v", updated.TagList)
	}
}
