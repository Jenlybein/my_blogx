package category

import (
	"encoding/json"
	"myblogx/common"
	"myblogx/conf"
	"myblogx/global"
	"myblogx/models"
	"myblogx/models/ctype"
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

func setupCategoryEnv(t *testing.T) *models.UserModel {
	t.Helper()
	db := testutil.SetupSQLite(
		t,
		&models.UserModel{},
		&models.UserConfModel{},
		&models.CategoryModel{},
		&models.ArticleModel{},
	)
	global.Config = &conf.Config{
		Jwt: conf.Jwt{
			Expire: 1,
			Secret: "category-secret",
			Issuer: "category-test",
		},
	}

	user := &models.UserModel{
		Username: "category_user",
		Password: "x",
		Role:     enum.RoleUser,
	}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("创建用户失败: %v", err)
	}
	return user
}

func tokenForUser(t *testing.T, user *models.UserModel) string {
	t.Helper()
	return testutil.IssueAccessToken(t, user)
}

func TestCategoryCRUD(t *testing.T) {
	user := setupCategoryEnv(t)
	api := CategoryApi{}
	claims := &jwts.MyClaims{Claims: jwts.Claims{UserID: user.ID, Role: enum.RoleUser, Username: user.Username}}

	{
		c, w := newCtx()
		c.Set("claims", claims)
		c.Set("requestJson", CategoryRequest{Title: "后端分类"})
		api.CategoryCreateUpdateView(c)
		if code := readCode(t, w); code != 0 {
			t.Fatalf("创建分类应成功, body=%s", w.Body.String())
		}
	}

	{
		c, w := newCtx()
		c.Set("claims", claims)
		c.Set("requestJson", CategoryRequest{Title: "后端分类"})
		api.CategoryCreateUpdateView(c)
		if code := readCode(t, w); code == 0 {
			t.Fatalf("重复分类应失败, body=%s", w.Body.String())
		}
	}

	var cat models.CategoryModel
	if err := global.DB.Where("user_id = ? and title = ?", user.ID, "后端分类").First(&cat).Error; err != nil {
		t.Fatalf("查询分类失败: %v", err)
	}

	{
		c, w := newCtx()
		c.Set("claims", claims)
		c.Set("requestJson", CategoryRequest{ID: cat.ID, Title: "后端分类-更新"})
		api.CategoryCreateUpdateView(c)
		if code := readCode(t, w); code != 0 {
			t.Fatalf("更新分类应成功, body=%s", w.Body.String())
		}
	}

	token := tokenForUser(t, user)
	{
		c, w := newCtx()
		c.Set("requestQuery", CategoryListRequest{
			PageInfo: common.PageInfo{Page: 1, Limit: 10},
			Type:     1,
		})
		req := httptest.NewRequest(http.MethodGet, "/articles/category", nil)
		req.Header.Set("token", token)
		c.Request = req
		api.CategoryListView(c)
		if code := readCode(t, w); code != 0 {
			t.Fatalf("分类列表应成功, body=%s", w.Body.String())
		}
	}

	{
		c, w := newCtx()
		c.Set("requestJson", models.IDListRequest{IDList: []ctype.ID{}})
		req := httptest.NewRequest(http.MethodDelete, "/articles/category", nil)
		req.Header.Set("token", token)
		c.Request = req
		api.CategoryDeleteView(c)
		if code := readCode(t, w); code == 0 {
			t.Fatalf("空删除列表应失败, body=%s", w.Body.String())
		}
	}

	{
		c, w := newCtx()
		c.Set("requestJson", models.IDListRequest{IDList: []ctype.ID{cat.ID}})
		req := httptest.NewRequest(http.MethodDelete, "/articles/category", nil)
		req.Header.Set("token", token)
		c.Request = req
		api.CategoryDeleteView(c)
		if code := readCode(t, w); code != 0 {
			t.Fatalf("删除分类应成功, body=%s", w.Body.String())
		}
	}
}
