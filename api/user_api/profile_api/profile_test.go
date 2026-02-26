package profile_api_test

import (
	"encoding/json"
	"myblogx/api/user_api/profile_api"
	"myblogx/models"
	"myblogx/models/enum"
	"myblogx/test/testutil"
	"myblogx/utils/jwts"
	"net/http/httptest"
	"testing"
	"time"

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

func TestProfileHandlers(t *testing.T) {
	db := testutil.SetupSQLite(t, &models.UserModel{}, &models.UserConfModel{})
	user := models.UserModel{
		Username: "u1",
		Password: "x",
		Nickname: "nick",
		Role:     enum.RoleUser,
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("创建用户失败: %v", err)
	}

	api := profile_api.ProfileApi{}

	{
		c, w := newCtx()
		c.Set("claims", &jwts.MyClaims{Claims: jwts.Claims{UserID: user.ID, Role: user.Role}})
		api.UserDetailView(c)
		if code := readCode(t, w); code != 0 {
			t.Fatalf("用户详情失败, code=%d body=%s", code, w.Body.String())
		}
	}

	{
		c, w := newCtx()
		c.Set("requestQuery", models.IDRequest{ID: user.ID})
		api.UserBaseInfoView(c)
		if code := readCode(t, w); code != 0 {
			t.Fatalf("用户基础信息失败, code=%d body=%s", code, w.Body.String())
		}
	}

	{
		c, w := newCtx()
		newNick := "new-nick"
		c.Set("claims", &jwts.MyClaims{Claims: jwts.Claims{UserID: user.ID, Role: user.Role}})
		c.Set("requestJson", profile_api.UserInfoUpdateRequest{
			Nickname: &newNick,
		})
		api.UserInfoUpdateView(c)
		if code := readCode(t, w); code != 0 {
			t.Fatalf("用户信息更新失败, code=%d body=%s", code, w.Body.String())
		}
	}

	{
		c, w := newCtx()
		role := enum.RoleAdmin
		c.Set("requestJson", profile_api.AdminUserInfoUpdateRequest{
			UserID: user.ID,
			Role:   &role,
		})
		api.AdminUserInfoUpdateView(c)
		if code := readCode(t, w); code != 0 {
			t.Fatalf("管理员更新用户失败, code=%d body=%s", code, w.Body.String())
		}
	}

	{
		// 覆盖用户名更新频率限制分支
		now := time.Now()
		_ = db.Model(&models.UserConfModel{}).Where("user_id = ?", user.ID).
			Update("updated_username_date", &now).Error
		c, w := newCtx()
		name := "u2_newname"
		c.Set("claims", &jwts.MyClaims{Claims: jwts.Claims{UserID: user.ID, Role: user.Role}})
		c.Set("requestJson", profile_api.UserInfoUpdateRequest{
			Username: &name,
		})
		api.UserInfoUpdateView(c)
		if code := readCode(t, w); code == 0 {
			t.Fatalf("用户名频率限制分支应失败, body=%s", w.Body.String())
		}
	}
}
