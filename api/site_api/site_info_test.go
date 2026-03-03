package site_api_test

import (
	"bytes"
	"encoding/json"
	"myblogx/api/site_api"
	"myblogx/conf"
	"myblogx/global"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func newSiteCtx(req *http.Request) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	return c, w
}

func readSiteCode(t *testing.T, w *httptest.ResponseRecorder) int {
	t.Helper()
	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("解析响应失败: %v body=%s", err, w.Body.String())
	}
	return int(body["code"].(float64))
}

func setupSiteApiEnv(t *testing.T) {
	t.Helper()
	cfgFile := filepath.Join(t.TempDir(), "settings.yaml")
	global.Flags = &global.FlagRecord{File: cfgFile}
	global.Config = &conf.Config{
		Site: conf.Site{},
		QQ: conf.QQ{
			AppID:    "app-id",
			AppKey:   "app-key-origin",
			Redirect: "https://example.com/callback",
		},
		Email: conf.Email{
			Domain:   "smtp.example.com",
			Port:     465,
			AuthCode: "email-auth-origin",
		},
		QiNiu: conf.QiNiu{
			SecretKey: "qiniu-secret-origin",
		},
		AI: conf.AI{
			SecretKey: "ai-secret-origin",
		},
	}
}

func TestSiteInfoViews(t *testing.T) {
	setupSiteApiEnv(t)
	api := site_api.SiteApi{}

	t.Run("QQ登录地址", func(t *testing.T) {
		c, w := newSiteCtx(httptest.NewRequest(http.MethodGet, "/site/qq_url", nil))
		api.SiteInfoQQView(c)
		if code := readSiteCode(t, w); code != 0 {
			t.Fatalf("QQ 地址接口应成功, body=%s", w.Body.String())
		}
		if !strings.Contains(w.Body.String(), "graph.qq.com") {
			t.Fatalf("QQ 地址返回异常: %s", w.Body.String())
		}
	})

	t.Run("站点信息", func(t *testing.T) {
		c, w := newSiteCtx(httptest.NewRequest(http.MethodGet, "/site/site", nil))
		c.Set("requestUri", site_api.SiteInfoRequest{Name: "site"})
		api.SiteInfoView(c)
		if code := readSiteCode(t, w); code != 0 {
			t.Fatalf("站点信息接口应成功, body=%s", w.Body.String())
		}
		if !strings.Contains(w.Body.String(), global.Version) {
			t.Fatalf("站点信息应包含版本号, body=%s", w.Body.String())
		}
	})

	t.Run("管理员敏感信息脱敏", func(t *testing.T) {
		type pair struct {
			name string
			key  string
		}
		cases := []pair{
			{name: "email", key: "auth_code"},
			{name: "qq", key: "app_key"},
			{name: "qiniu", key: "secret_key"},
			{name: "ai", key: "secret"},
		}
		for _, tc := range cases {
			c, w := newSiteCtx(httptest.NewRequest(http.MethodGet, "/admin/"+tc.name, nil))
			c.Set("requestUri", site_api.SiteInfoRequest{Name: tc.name})
			api.SiteInfoAdminView(c)
			if code := readSiteCode(t, w); code != 0 {
				t.Fatalf("%s 管理接口应成功, body=%s", tc.name, w.Body.String())
			}
			if !strings.Contains(w.Body.String(), "******") {
				t.Fatalf("%s 未脱敏, body=%s", tc.name, w.Body.String())
			}
		}
	})

	t.Run("管理员未知配置", func(t *testing.T) {
		c, w := newSiteCtx(httptest.NewRequest(http.MethodGet, "/admin/unknown", nil))
		c.Set("requestUri", site_api.SiteInfoRequest{Name: "unknown"})
		api.SiteInfoAdminView(c)
		if code := readSiteCode(t, w); code == 0 {
			t.Fatalf("未知配置应失败, body=%s", w.Body.String())
		}
	})
}

func TestSiteUpdateView(t *testing.T) {
	setupSiteApiEnv(t)
	api := site_api.SiteApi{}

	t.Run("未知配置名", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/site/unknown", bytes.NewBufferString(`{}`))
		req.Header.Set("Content-Type", "application/json")
		c, w := newSiteCtx(req)
		c.Set("requestUri", site_api.SiteInfoRequest{Name: "unknown"})
		api.SiteUpdateView(c)
		if code := readSiteCode(t, w); code == 0 {
			t.Fatalf("未知配置名应失败, body=%s", w.Body.String())
		}
	})

	t.Run("JSON绑定失败", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/site/email", bytes.NewBufferString(`{"port":"bad"}`))
		req.Header.Set("Content-Type", "application/json")
		c, w := newSiteCtx(req)
		c.Set("requestUri", site_api.SiteInfoRequest{Name: "email"})
		api.SiteUpdateView(c)
		if code := readSiteCode(t, w); code == 0 {
			t.Fatalf("JSON 类型错误应失败, body=%s", w.Body.String())
		}
	})

	t.Run("敏感字段占位符保留原值", func(t *testing.T) {
		body := `{"domain":"smtp2.example.com","port":465,"send_email":"a@b.com","auth_code":"******","send_nickname":"n","ssl":true,"tls":false}`
		req := httptest.NewRequest(http.MethodPost, "/site/email", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		c, w := newSiteCtx(req)
		c.Set("requestUri", site_api.SiteInfoRequest{Name: "email"})
		api.SiteUpdateView(c)
		if code := readSiteCode(t, w); code != 0 {
			t.Fatalf("email 更新应成功, body=%s", w.Body.String())
		}
		if global.Config.Email.AuthCode != "email-auth-origin" {
			t.Fatalf("占位符应保留原 auth_code, got=%s", global.Config.Email.AuthCode)
		}
	})
}
