package service_test

import (
	"myblogx/conf"
	"myblogx/global"
	"myblogx/models/enum"
	redis_jwt "myblogx/service/redis_service/redis_jwt"
	"myblogx/test/testutil"
	"myblogx/utils/jwts"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestBlackTypeHelpers(t *testing.T) {
	if redis_jwt.UserBlackType.String() == "" {
		t.Fatal("BlackType.String 不应为空")
	}

	bt, err := redis_jwt.BlackTypeFromString("2")
	if err != nil || bt != redis_jwt.AdminBlackType {
		t.Fatalf("BlackTypeFromString 失败: bt=%v err=%v", bt, err)
	}
	if _, err = redis_jwt.BlackTypeFromString("x"); err == nil {
		t.Fatal("非法字符串应报错")
	}

	if redis_jwt.BlackTypeMsg(redis_jwt.DeviceBlackType) == "" {
		t.Fatal("BlackTypeMsg 不应为空")
	}
}

func TestTokenBlacklistFlow(t *testing.T) {
	_ = testutil.SetupMiniRedis(t)
	global.Config = &conf.Config{
		Jwt: conf.Jwt{
			Expire: 1,
			Secret: "test-secret",
			Issuer: "test",
		},
	}

	token, err := jwts.GetToken(jwts.Claims{
		UserID:   1,
		Role:     enum.RoleUser,
		Username: "u1",
	})
	if err != nil {
		t.Fatalf("生成 token 失败: %v", err)
	}

	redis_jwt.SetTokenBlack(token, redis_jwt.UserBlackType)
	msg, ok := redis_jwt.HasTokenBlack(token)
	if ok {
		t.Fatalf("黑名单 token 应返回 ok=false: %s", msg)
	}

	// 黑名单不存在时应允许通过
	if _, ok = redis_jwt.HasTokenBlack("not-exist-token"); !ok {
		t.Fatal("不存在黑名单的 token 应返回 ok=true")
	}
}

func TestHasTokenBlackByGin(t *testing.T) {
	_ = testutil.SetupMiniRedis(t)
	global.Config = &conf.Config{
		Jwt: conf.Jwt{
			Expire: 1,
			Secret: "test-secret",
			Issuer: "test",
		},
	}

	token, err := jwts.GetToken(jwts.Claims{
		UserID:   2,
		Role:     enum.RoleUser,
		Username: "u2",
	})
	if err != nil {
		t.Fatalf("生成 token 失败: %v", err)
	}
	redis_jwt.SetTokenBlack(token, redis_jwt.AdminBlackType)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodGet, "/?token="+token, nil)
	c.Request = req

	if _, ok := redis_jwt.HasTokenBlackByGin(c); ok {
		t.Fatal("黑名单 token 通过 query 检查时应返回 ok=false")
	}
}
