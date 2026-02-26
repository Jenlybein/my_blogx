package captcha_api_test

import (
	"encoding/json"
	"myblogx/api/captcha_api"
	"myblogx/conf"
	confsite "myblogx/conf/site"
	"myblogx/global"
	"myblogx/test/testutil"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type captchaResp struct {
	Code int            `json:"code"`
	Data map[string]any `json:"data"`
	Msg  string         `json:"msg"`
}

func newCaptchaCtx() (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	return c, w
}

func readCaptchaResp(t *testing.T, w *httptest.ResponseRecorder) captchaResp {
	t.Helper()
	var resp captchaResp
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("解析验证码接口响应失败: %v, body=%s", err, w.Body.String())
	}
	return resp
}

func setupCaptchaConfig(enable bool) {
	testutil.InitGlobals()
	global.Config = &conf.Config{
		Site: conf.Site{
			Login: confsite.Login{Captcha: enable},
		},
	}
}

func TestCaptchaViewDisabled(t *testing.T) {
	setupCaptchaConfig(false)
	api := captcha_api.ImageCaptchaApi{}

	c, w := newCaptchaCtx()
	api.CaptchaView(c)

	resp := readCaptchaResp(t, w)
	if resp.Code == 0 {
		t.Fatalf("验证码关闭时应返回失败, body=%s", w.Body.String())
	}
}

func TestCaptchaViewSuccess(t *testing.T) {
	setupCaptchaConfig(true)
	api := captcha_api.ImageCaptchaApi{}

	c, w := newCaptchaCtx()
	api.CaptchaView(c)

	resp := readCaptchaResp(t, w)
	if resp.Code != 0 {
		t.Fatalf("验证码生成应成功, body=%s", w.Body.String())
	}

	id, ok := resp.Data["captcha_id"].(string)
	if !ok || id == "" {
		t.Fatalf("captcha_id 为空或类型错误, data=%v", resp.Data)
	}

	base64, ok := resp.Data["base64"].(string)
	if !ok || len(base64) < 16 {
		t.Fatalf("base64 为空或类型错误, data=%v", resp.Data)
	}
}
