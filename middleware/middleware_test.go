package middleware_test

import (
	"encoding/json"
	"myblogx/conf"
	confsite "myblogx/conf/site"
	"myblogx/global"
	"myblogx/middleware"
	"myblogx/models"
	"myblogx/models/enum"
	redisEmail "myblogx/service/redis_service/redis_email"
	redisJWT "myblogx/service/redis_service/redis_jwt"
	"myblogx/test/testutil"
	"myblogx/utils/jwts"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

type bindReq struct {
	Name string `json:"name" form:"name" uri:"name" binding:"required"`
}

func TestBindMiddlewares(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.POST("/json", middleware.BindJson[bindReq], func(c *gin.Context) {
		req := middleware.GetBindJson[bindReq](c)
		c.JSON(http.StatusOK, gin.H{"name": req.Name})
	})
	r.GET("/query", middleware.BindQuery[bindReq], func(c *gin.Context) {
		req := middleware.GetBindQuery[bindReq](c)
		c.JSON(http.StatusOK, gin.H{"name": req.Name})
	})
	r.GET("/uri/:name", middleware.BindUri[bindReq], func(c *gin.Context) {
		req := middleware.GetBindUri[bindReq](c)
		c.JSON(http.StatusOK, gin.H{"name": req.Name})
	})

	{
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/json", strings.NewReader(`{"name":"alice"}`))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("BindJson 状态码异常: %d", w.Code)
		}
	}
	{
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/json", nil)
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		if !hasBizCode(w, 1002) {
			t.Fatalf("空 JSON 请求体应返回 1002, body=%s", w.Body.String())
		}
		if !strings.Contains(w.Body.String(), "请求体不能为空，请在 Body 中传 JSON 参数") {
			t.Fatalf("空 JSON 请求体提示不明确, body=%s", w.Body.String())
		}
	}
	{
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/query?name=bob", nil)
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("BindQuery 状态码异常: %d", w.Code)
		}
	}
	{
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/uri/cindy", nil)
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("BindUri 状态码异常: %d", w.Code)
		}
	}
}

func setupAuthEnv(t *testing.T) {
	t.Helper()
	_ = testutil.SetupMiniRedis(t)
	_ = testutil.SetupSQLite(t, &models.UserModel{})
	global.Config = &conf.Config{
		Jwt: conf.Jwt{
			Expire: 1,
			Secret: "secret",
			Issuer: "issuer",
		},
		Site: conf.Site{
			Login: confsite.Login{Captcha: false},
		},
	}
}

func TestAuthAndAdminMiddleware(t *testing.T) {
	setupAuthEnv(t)
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.GET("/auth", middleware.AuthMiddleware, func(c *gin.Context) {
		claims := jwts.MustGetClaimsByGin(c)
		c.JSON(http.StatusOK, gin.H{"user": claims.Username})
	})
	r.GET("/admin", middleware.AuthMiddleware, middleware.AdminMiddleware, func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	user := &models.UserModel{Username: "u1", Password: "x", Role: enum.RoleUser}
	admin := &models.UserModel{Username: "admin", Password: "x", Role: enum.RoleAdmin}
	if err := global.DB.Create(user).Error; err != nil {
		t.Fatalf("创建普通用户失败: %v", err)
	}
	if err := global.DB.Create(admin).Error; err != nil {
		t.Fatalf("创建管理员失败: %v", err)
	}
	userToken := testutil.IssueAccessToken(t, user)
	adminToken := testutil.IssueAccessToken(t, admin)

	{
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/auth", nil)
		req.Header.Set("token", userToken)
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("AuthMiddleware 应通过, code=%d", w.Code)
		}
	}

	{
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/admin", nil)
		req.Header.Set("token", userToken)
		r.ServeHTTP(w, req)
		if !hasBizCode(w, 1001) {
			t.Fatalf("普通用户访问 admin 应返回 1001, body=%s", w.Body.String())
		}
	}

	{
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/admin", nil)
		req.Header.Set("token", adminToken)
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("管理员应通过, code=%d", w.Code)
		}
	}

	{
		redisJWT.SetTokenBlack(userToken, redisJWT.UserBlackType)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/auth", nil)
		req.Header.Set("token", userToken)
		r.ServeHTTP(w, req)
		if !hasBizCode(w, 1001) {
			t.Fatalf("黑名单 token 应被拦截, body=%s", w.Body.String())
		}
	}
}

func TestCaptchaAndEmailVerifyMiddleware(t *testing.T) {
	setupAuthEnv(t)
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.POST("/captcha", middleware.CaptchaMiddleware, func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	r.POST("/email", middleware.EmailVerifyMiddleware, func(c *gin.Context) {
		email, _ := c.Get("email")
		c.JSON(http.StatusOK, gin.H{"email": email})
	})

	global.Config.Site.Login.Captcha = false
	{
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/captcha", strings.NewReader(`{}`))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("验证码关闭时应通过, code=%d", w.Code)
		}
	}

	global.Config.Site.Login.Captcha = true
	_ = global.ImageCaptchaStore.Set("cid", "1234")
	{
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/captcha", strings.NewReader(`{"captcha_id":"cid","captcha_code":"1234"}`))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("正确验证码应通过, code=%d", w.Code)
		}
	}

	_ = global.ImageCaptchaStore.Set("cid2", "5678")
	{
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/captcha", strings.NewReader(`{"captcha_id":"cid2","captcha_code":"wrong"}`))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		if !hasBizCode(w, 1001) {
			t.Fatalf("错误验证码应失败, body=%s", w.Body.String())
		}
	}

	if err := redisEmail.Store("eid", "u@example.com", "8888", 1, 3); err != nil {
		t.Fatalf("存储邮箱验证码失败: %v", err)
	}
	{
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/email", strings.NewReader(`{"email_id":"eid","email_code":"8888"}`))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("邮箱验证码应通过, code=%d", w.Code)
		}
		if !strings.Contains(w.Body.String(), "u@example.com") {
			t.Fatalf("未写入 email 到 context, body=%s", w.Body.String())
		}
	}
}

func hasBizCode(w *httptest.ResponseRecorder, code int) bool {
	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		return false
	}
	v, ok := resp["code"]
	if !ok {
		return false
	}
	return int(v.(float64)) == code
}
