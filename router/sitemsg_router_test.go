package router_test

import (
	"encoding/json"
	"myblogx/conf"
	"myblogx/global"
	"myblogx/models"
	"myblogx/models/enum"
	"myblogx/router"
	"myblogx/test/testutil"
	"myblogx/utils/jwts"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func readBizCode(t *testing.T, w *httptest.ResponseRecorder) int {
	t.Helper()

	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("解析响应失败: %v, body=%s", err, w.Body.String())
	}

	return int(body["code"].(float64))
}

func TestSitemsgRouterPutConfBindsJSON(t *testing.T) {
	testutil.SetupMiniRedis(t)
	db := testutil.SetupSQLite(t, &models.UserModel{}, &models.UserConfModel{})
	global.Config = &conf.Config{
		Jwt: conf.Jwt{
			Expire: 1,
			Secret: "router-test-secret",
			Issuer: "blogx-test",
		},
	}

	user := models.UserModel{
		Username: "msg_user",
		Password: "x",
		Nickname: "msg",
		Role:     enum.RoleUser,
	}
	if err := db.Create(&user).Error; err != nil {
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

	gin.SetMode(gin.TestMode)
	r := gin.New()
	apiGroup := r.Group("/api")
	router.SitemsgRouter(apiGroup)

	req := testutil.NewJSONRequest(http.MethodPut, "/api/sitemsg/conf", `{
		"digg_notice_enabled": false,
		"comment_notice_enabled": false,
		"favor_notice_enabled": false,
		"private_chat_notice_enabled": false
	}`)
	req.Header.Set("token", token)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("状态码异常: %d, body=%s", w.Code, w.Body.String())
	}
	if code := readBizCode(t, w); code != 0 {
		t.Fatalf("业务码异常: %d, body=%s", code, w.Body.String())
	}

	var confModel models.UserConfModel
	if err := db.Take(&confModel, user.ID).Error; err != nil {
		t.Fatalf("查询用户消息配置失败: %v", err)
	}

	if confModel.DiggNoticeEnabled || confModel.CommentNoticeEnabled ||
		confModel.FavorNoticeEnabled || confModel.PrivateChatNoticeEnabled {
		t.Fatalf("消息配置未按请求更新: %+v", confModel)
	}
}
