package favorite

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

func setupFavoriteEnv(t *testing.T) *models.UserModel {
	t.Helper()
	_ = testutil.SetupMiniRedis(t)
	db := testutil.SetupSQLite(
		t,
		&models.UserModel{},
		&models.UserConfModel{},
		&models.FavoriteModel{},
		&models.UserArticleFavorModel{},
		&models.ArticleModel{},
	)
	global.Config = &conf.Config{
		Jwt: conf.Jwt{
			Expire: 1,
			Secret: "favorite-secret",
			Issuer: "favorite-test",
		},
	}

	user := &models.UserModel{
		Username: "favorite_user",
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

func TestFavoriteCRUD(t *testing.T) {
	user := setupFavoriteEnv(t)
	api := FavoriteApi{}
	claims := &jwts.MyClaims{Claims: jwts.Claims{UserID: user.ID, Role: enum.RoleUser, Username: user.Username}}

	{
		c, w := newCtx()
		c.Set("claims", claims)
		c.Set("requestJson", FavoriteRequest{
			Title:    "默认收藏组",
			Abstract: "desc",
			Cover:    "cover.png",
		})
		api.FavoriteCreateUpdateView(c)
		if code := readCode(t, w); code != 0 {
			t.Fatalf("创建收藏夹应成功, body=%s", w.Body.String())
		}
	}

	{
		c, w := newCtx()
		c.Set("claims", claims)
		c.Set("requestJson", FavoriteRequest{
			Title:    "默认收藏组",
			Abstract: "desc2",
		})
		api.FavoriteCreateUpdateView(c)
		if code := readCode(t, w); code == 0 {
			t.Fatalf("重复收藏夹应失败, body=%s", w.Body.String())
		}
	}

	var fav models.FavoriteModel
	if err := global.DB.Where("user_id = ? and title = ?", user.ID, "默认收藏组").First(&fav).Error; err != nil {
		t.Fatalf("查询收藏夹失败: %v", err)
	}

	{
		c, w := newCtx()
		c.Set("claims", claims)
		c.Set("requestJson", FavoriteRequest{
			ID:       fav.ID,
			Title:    "默认收藏组-更新",
			Abstract: "desc3",
			Cover:    "new.png",
		})
		api.FavoriteCreateUpdateView(c)
		if code := readCode(t, w); code != 0 {
			t.Fatalf("更新收藏夹应成功, body=%s", w.Body.String())
		}
	}

	token := tokenForUser(t, user)
	{
		c, w := newCtx()
		c.Set("requestQuery", FavoriteListRequest{
			PageInfo: common.PageInfo{Page: 1, Limit: 10},
			Type:     1,
		})
		req := httptest.NewRequest(http.MethodGet, "/articles/favorite", nil)
		req.Header.Set("token", token)
		c.Request = req
		api.FavoriteListView(c)
		if code := readCode(t, w); code != 0 {
			t.Fatalf("收藏夹列表应成功, body=%s", w.Body.String())
		}
	}

	{
		c, w := newCtx()
		c.Set("requestJson", models.RemoveRequest{IDList: []uint{}})
		req := httptest.NewRequest(http.MethodDelete, "/articles/favorite", nil)
		req.Header.Set("token", token)
		c.Request = req
		api.FavoriteDeleteView(c)
		if code := readCode(t, w); code == 0 {
			t.Fatalf("空删除列表应失败, body=%s", w.Body.String())
		}
	}

	article := models.ArticleModel{
		Title:    "favorite-delete-article",
		Content:  "content",
		AuthorID: user.ID,
		Status:   enum.ArticleStatusPublished,
	}
	if err := global.DB.Create(&article).Error; err != nil {
		t.Fatalf("创建文章失败: %v", err)
	}
	if err := global.DB.Create(&models.UserArticleFavorModel{
		ArticleID: article.ID,
		UserID:    user.ID,
		FavorID:   fav.ID,
	}).Error; err != nil {
		t.Fatalf("创建收藏关联失败: %v", err)
	}

	{
		c, w := newCtx()
		c.Set("requestJson", models.RemoveRequest{IDList: []uint{fav.ID}})
		req := httptest.NewRequest(http.MethodDelete, "/articles/favorite", nil)
		req.Header.Set("token", token)
		c.Request = req
		api.FavoriteDeleteView(c)
		if code := readCode(t, w); code != 0 {
			t.Fatalf("删除收藏夹应成功, body=%s", w.Body.String())
		}
	}
}
