package profile_api_test

import (
	"encoding/json"
	"myblogx/api/user_api/profile_api"
	"myblogx/models"
	"myblogx/models/ctype"
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
	db := testutil.SetupSQLite(t, &models.UserModel{}, &models.UserConfModel{}, &models.TagModel{})
	email := "u1@example.com"
	user := models.UserModel{
		Username: "u1",
		Password: "x",
		Email:    &email,
		Nickname: "nick",
		Role:     enum.RoleUser,
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("创建用户失败: %v", err)
	}

	tag := models.TagModel{Title: "Go", IsEnabled: true}
	disabledTag := models.TagModel{Title: "Hidden", IsEnabled: false}
	if err := db.Create(&tag).Error; err != nil {
		t.Fatalf("创建启用标签失败: %v", err)
	}
	if err := db.Create(&disabledTag).Error; err != nil {
		t.Fatalf("创建停用标签失败: %v", err)
	}
	if err := db.Model(&disabledTag).Update("is_enabled", false).Error; err != nil {
		t.Fatalf("更新停用标签状态失败: %v", err)
	}

	api := profile_api.ProfileApi{}

	{
		c, w := newCtx()
		c.Set("claims", &jwts.MyClaims{Claims: jwts.Claims{UserID: user.ID, Role: user.Role}})
		api.UserDetailView(c)
		if code := readCode(t, w); code != 0 {
			t.Fatalf("用户详情失败, code=%d body=%s", code, w.Body.String())
		}

		var body struct {
			Code int `json:"code"`
			Data struct {
				Email       *string `json:"email"`
				HasPassword bool    `json:"has_password"`
			} `json:"data"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
			t.Fatalf("解析用户详情响应失败: %v", err)
		}
		if body.Data.Email == nil || *body.Data.Email != email {
			t.Fatalf("用户详情返回邮箱错误: %+v", body.Data.Email)
		}
		if !body.Data.HasPassword {
			t.Fatalf("用户详情应返回已绑定密码")
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
		likeTags := []ctype.ID{tag.ID, tag.ID, 0}
		c.Set("claims", &jwts.MyClaims{Claims: jwts.Claims{UserID: user.ID, Role: user.Role}})
		c.Set("requestJson", profile_api.UserInfoUpdateRequest{
			LikeTags: &likeTags,
		})
		api.UserInfoUpdateView(c)
		if code := readCode(t, w); code != 0 {
			t.Fatalf("偏好标签更新失败, code=%d body=%s", code, w.Body.String())
		}

		var conf models.UserConfModel
		if err := db.Take(&conf, "user_id = ?", user.ID).Error; err != nil {
			t.Fatalf("查询用户配置失败: %v", err)
		}
		if len(conf.LikeTags) != 1 || conf.LikeTags[0] != tag.ID {
			t.Fatalf("偏好标签去重结果异常: %+v", conf.LikeTags)
		}
	}

	{
		c, w := newCtx()
		likeTags := []ctype.ID{disabledTag.ID}
		c.Set("claims", &jwts.MyClaims{Claims: jwts.Claims{UserID: user.ID, Role: user.Role}})
		c.Set("requestJson", profile_api.UserInfoUpdateRequest{
			LikeTags: &likeTags,
		})
		api.UserInfoUpdateView(c)
		if code := readCode(t, w); code == 0 {
			t.Fatalf("停用标签不应允许更新, body=%s", w.Body.String())
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
