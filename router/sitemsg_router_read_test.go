package router_test

import (
	"myblogx/conf"
	"myblogx/global"
	"myblogx/models"
	"myblogx/models/enum"
	"myblogx/models/enum/message_enum"
	"myblogx/router"
	"myblogx/test/testutil"
	"myblogx/utils/jwts"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func setupSitemsgRouterEnv(t *testing.T) (*models.UserModel, string) {
	t.Helper()
	testutil.SetupMiniRedis(t)
	db := testutil.SetupSQLite(t, &models.UserModel{}, &models.UserConfModel{}, &models.ArticleMessageModel{})
	global.Config = &conf.Config{
		Jwt: conf.Jwt{
			Expire: 1,
			Secret: "router-test-secret",
			Issuer: "blogx-test",
		},
	}

	user := &models.UserModel{
		Username: "msg_user",
		Password: "x",
		Nickname: "msg",
		Role:     enum.RoleUser,
	}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("创建用户失败: %v", err)
	}

	token, err := jwts.GetToken(jwts.Claims{
		UserID:   user.ID,
		Role:     user.Role,
		Username: user.Username,
	})
	if err != nil {
		t.Fatalf("生成 token 失败: %v", err)
	}
	return user, token
}

func newSitemsgRouterEngine() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	apiGroup := r.Group("/api")
	router.SitemsgRouter(apiGroup)
	return r
}

func TestSitemsgRouterPostSupportsReadByID(t *testing.T) {
	user, token := setupSitemsgRouterEnv(t)
	db := global.DB

	msg := models.ArticleMessageModel{
		ReceiverID: user.ID,
		Type:       message_enum.SystemType,
		Content:    "system",
	}
	if err := db.Create(&msg).Error; err != nil {
		t.Fatalf("创建消息失败: %v", err)
	}

	req := testutil.NewJSONRequest(http.MethodPost, "/api/sitemsg", `{"id":1}`)
	req.Header.Set("token", token)

	w := httptest.NewRecorder()
	newSitemsgRouterEngine().ServeHTTP(w, req)

	if w.Code != http.StatusOK || readBizCode(t, w) != 0 {
		t.Fatalf("按 id 标记已读失败, body=%s", w.Body.String())
	}

	var check models.ArticleMessageModel
	if err := db.Take(&check, msg.ID).Error; err != nil {
		t.Fatalf("查询消息失败: %v", err)
	}
	if !check.IsRead {
		t.Fatalf("消息未标记已读: %+v", check)
	}
}

func TestSitemsgRouterPostSupportsBatchReadByType(t *testing.T) {
	user, token := setupSitemsgRouterEnv(t)
	db := global.DB

	msgs := []models.ArticleMessageModel{
		{ReceiverID: user.ID, Type: message_enum.DiggArticleType, Content: "d1"},
		{ReceiverID: user.ID, Type: message_enum.FavorArticleType, Content: "f1"},
		{ReceiverID: user.ID, Type: message_enum.SystemType, Content: "s1"},
	}
	if err := db.Create(&msgs).Error; err != nil {
		t.Fatalf("创建消息失败: %v", err)
	}

	req := testutil.NewJSONRequest(http.MethodPost, "/api/sitemsg", `{"t":2}`)
	req.Header.Set("token", token)

	w := httptest.NewRecorder()
	newSitemsgRouterEngine().ServeHTTP(w, req)

	if w.Code != http.StatusOK || readBizCode(t, w) != 0 {
		t.Fatalf("按类型批量已读失败, body=%s", w.Body.String())
	}

	var checks []models.ArticleMessageModel
	if err := db.Order("id asc").Find(&checks).Error; err != nil {
		t.Fatalf("查询消息失败: %v", err)
	}
	if !checks[0].IsRead || !checks[1].IsRead {
		t.Fatalf("点赞/收藏消息未全部标记已读: %+v", checks)
	}
	if checks[2].IsRead {
		t.Fatalf("系统消息不应被批量已读: %+v", checks[2])
	}
}
