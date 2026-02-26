package article_api

import (
	"encoding/json"
	"myblogx/common"
	"myblogx/conf"
	confsite "myblogx/conf/site"
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

func setupArticleEnv(t *testing.T) *models.UserModel {
	t.Helper()
	_ = testutil.SetupMiniRedis(t)
	db := testutil.SetupSQLite(
		t,
		&models.UserModel{},
		&models.UserConfModel{},
		&models.CategoryModel{},
		&models.ArticleModel{},
		&models.ArticleDiggModel{},
		&models.FavoriteModel{},
		&models.UserArticleFavorModel{},
		&models.UserTopArticleModel{},
		&models.UserArticleViewHistoryModel{},
		&models.CommentModel{},
	)
	global.Config = &conf.Config{
		Jwt: conf.Jwt{
			Expire: 1,
			Secret: "article-secret",
			Issuer: "article-test",
		},
		Site: conf.Site{
			SiteInfo: confsite.SiteInfo{Mode: enum.SiteModeCommunity},
			Article:  confsite.Article{SkipExamining: false},
		},
	}

	user := &models.UserModel{
		Username: "u1",
		Password: "x",
		Role:     enum.RoleUser,
	}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("创建用户失败: %v", err)
	}
	return user
}

func TestValidateRequestAndList(t *testing.T) {
	user := setupArticleEnv(t)

	{
		c, _ := newCtx()
		if err := validateRequest(ArticleListRequest{Type: 1}, nil, c); err == nil {
			t.Fatal("Type=1 且 user_id 为空应失败")
		}
	}

	article := models.ArticleModel{
		Title:    "a1",
		Content:  "hello",
		AuthorID: user.ID,
		Status:   enum.ArticleStatusPublished,
	}
	if err := global.DB.Create(&article).Error; err != nil {
		t.Fatalf("创建文章失败: %v", err)
	}

	{
		c, w := newCtx()
		c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
		c.Set("requestQuery", ArticleListRequest{
			PageInfo: common.PageInfo{Page: 1, Limit: 10},
			Type:     1,
			UserID:   user.ID,
			Status:   enum.ArticleStatusPublished,
		})
		ArticleApi{}.ArticleListView(c)
		if code := readCode(t, w); code != 0 {
			t.Fatalf("文章列表失败, code=%d body=%s", code, w.Body.String())
		}
	}
}

func TestArticleCreateUpdateExamineAndRemove(t *testing.T) {
	user := setupArticleEnv(t)
	db := global.DB

	cat := models.CategoryModel{Title: "go", UserID: user.ID}
	if err := db.Create(&cat).Error; err != nil {
		t.Fatalf("创建分类失败: %v", err)
	}

	api := ArticleApi{}
	claims := &jwts.MyClaims{Claims: jwts.Claims{UserID: user.ID, Role: enum.RoleUser, Username: user.Username}}

	{
		c, w := newCtx()
		c.Set("claims", claims)
		c.Set("requestJson", ArticleCreateRequest{
			Title:          "t1",
			Content:        "content",
			CategoryID:     &cat.ID,
			CommentsToggle: true,
			Status:         enum.ArticleStatusExamining,
		})
		api.ArticleCreateView(c)
		if code := readCode(t, w); code != 0 {
			t.Fatalf("创建文章失败, code=%d body=%s", code, w.Body.String())
		}
	}

	var created models.ArticleModel
	if err := db.Order("id desc").First(&created).Error; err != nil {
		t.Fatalf("查询创建文章失败: %v", err)
	}

	{
		c, w := newCtx()
		c.Set("claims", claims)
		c.Set("requestUri", models.IDRequest{ID: created.ID})
		c.Set("requestJson", ArticleUpdateRequest{
			Title:          "t1-updated",
			Content:        "new content",
			CategoryID:     &cat.ID,
			CommentsToggle: false,
		})
		api.ArticleUpdateView(c)
		if code := readCode(t, w); code != 0 {
			t.Fatalf("更新文章失败, code=%d body=%s", code, w.Body.String())
		}
	}

	{
		c, w := newCtx()
		c.Set("requestUri", models.IDRequest{ID: created.ID})
		c.Set("requestJson", ArticleExamineRequest{Status: enum.ArticleStatusPublished})
		api.ArticleExamineView(c)
		if code := readCode(t, w); code != 0 {
			t.Fatalf("审核文章失败, code=%d body=%s", code, w.Body.String())
		}
	}

	{
		c, w := newCtx()
		c.Set("requestJson", models.RemoveRequest{IDList: []uint{created.ID}})
		api.ArticleRemoveView(c)
		if code := readCode(t, w); code != 0 {
			t.Fatalf("删除文章失败, code=%d body=%s", code, w.Body.String())
		}
	}
}

func TestArticleDiggFavoriteVisitDetailRemoveUser(t *testing.T) {
	user := setupArticleEnv(t)
	db := global.DB
	api := ArticleApi{}

	article := models.ArticleModel{
		Title:    "a1",
		Content:  "content",
		AuthorID: user.ID,
		Status:   enum.ArticleStatusPublished,
	}
	if err := db.Create(&article).Error; err != nil {
		t.Fatalf("创建文章失败: %v", err)
	}

	claims := &jwts.MyClaims{Claims: jwts.Claims{UserID: user.ID, Role: enum.RoleUser, Username: user.Username}}

	{
		c, w := newCtx()
		c.Set("claims", claims)
		c.Set("requestUri", models.IDRequest{ID: article.ID})
		api.ArticleDiggView(c)
		if code := readCode(t, w); code != 0 {
			t.Fatalf("点赞失败, code=%d body=%s", code, w.Body.String())
		}
	}
	{
		c, w := newCtx()
		c.Set("claims", claims)
		c.Set("requestUri", models.IDRequest{ID: article.ID})
		api.ArticleDiggView(c)
		if code := readCode(t, w); code != 0 {
			t.Fatalf("取消点赞失败, code=%d body=%s", code, w.Body.String())
		}
	}

	{
		c, w := newCtx()
		c.Set("claims", claims)
		c.Set("requestJson", ArticleFavoriteRequest{ArticleID: article.ID})
		api.ArticleFavoriteSaveView(c)
		if code := readCode(t, w); code != 0 {
			t.Fatalf("收藏失败, code=%d body=%s", code, w.Body.String())
		}
	}
	{
		c, w := newCtx()
		c.Set("claims", claims)
		c.Set("requestJson", ArticleFavoriteRequest{ArticleID: article.ID})
		api.ArticleFavoriteSaveView(c)
		if code := readCode(t, w); code != 0 {
			t.Fatalf("取消收藏失败, code=%d body=%s", code, w.Body.String())
		}
	}

	{
		// 未登录且缺失 UA 时，走跳过统计分支
		c, w := newCtx()
		c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
		c.Set("requestJson", ArticleViewCountRequest{ArticleID: article.ID})
		api.ArticleVisitView(c)
		if code := readCode(t, w); code != 0 {
			t.Fatalf("访问计数分支失败, code=%d body=%s", code, w.Body.String())
		}
	}

	token, _ := jwts.GetToken(jwts.Claims{
		UserID:   user.ID,
		Role:     user.Role,
		Username: user.Username,
	})
	{
		c, w := newCtx()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("token", token)
		c.Request = req
		c.Set("requestUri", models.IDRequest{ID: article.ID})
		api.ArticleDetailView(c)
		if code := readCode(t, w); code != 0 {
			t.Fatalf("文章详情失败, code=%d body=%s", code, w.Body.String())
		}
	}

	{
		c, w := newCtx()
		req := httptest.NewRequest(http.MethodDelete, "/", nil)
		req.Header.Set("token", token)
		c.Request = req
		c.Set("requestUri", models.IDRequest{ID: article.ID})
		api.ArticleRemoveUserView(c)
		if code := readCode(t, w); code != 0 {
			t.Fatalf("用户删除文章失败, code=%d body=%s", code, w.Body.String())
		}
	}
}
