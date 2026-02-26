package log_api_test

import (
	"encoding/json"
	userlog "myblogx/api/user_api/log_api"
	"myblogx/common"
	"myblogx/models"
	"myblogx/models/enum"
	"myblogx/test/testutil"
	"myblogx/utils/jwts"
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

func TestUserLoginLogList(t *testing.T) {
	db := testutil.SetupSQLite(t, &models.UserModel{}, &models.UserConfModel{}, &models.UserLoginModel{})
	user := models.UserModel{Username: "u1", Password: "x", Nickname: "nick", Role: enum.RoleUser}
	admin := models.UserModel{Username: "a1", Password: "x", Nickname: "admin", Role: enum.RoleAdmin}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("创建用户失败: %v", err)
	}
	if err := db.Create(&admin).Error; err != nil {
		t.Fatalf("创建管理员失败: %v", err)
	}
	if err := db.Create(&models.UserLoginModel{UserID: user.ID, IP: "1.1.1.1", Addr: "x"}).Error; err != nil {
		t.Fatalf("创建登录日志失败: %v", err)
	}

	api := userlog.LogApi{}

	{
		c, w := newCtx()
		c.Set("claims", &jwts.MyClaims{Claims: jwts.Claims{UserID: admin.ID, Role: enum.RoleAdmin}})
		c.Set("requestQuery", userlog.UserLoginListRequest{
			PageInfo: common.PageInfo{Page: 1, Limit: 10},
			UserID:   user.ID,
			Type:     2,
		})
		api.UserLoginLogList(c)
		if code := readCode(t, w); code != 0 {
			t.Fatalf("管理员查询失败, code=%d body=%s", code, w.Body.String())
		}
	}

	{
		c, w := newCtx()
		c.Set("claims", &jwts.MyClaims{Claims: jwts.Claims{UserID: user.ID, Role: enum.RoleUser}})
		c.Set("requestQuery", userlog.UserLoginListRequest{
			PageInfo: common.PageInfo{Page: 1, Limit: 10},
			UserID:   admin.ID,
			Type:     1,
		})
		api.UserLoginLogList(c)
		if code := readCode(t, w); code == 0 {
			t.Fatalf("普通用户越权查询应失败, body=%s", w.Body.String())
		}
	}
}
