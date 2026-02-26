package log_service_test

import (
	"errors"
	"myblogx/conf"
	"myblogx/global"
	"myblogx/models"
	"myblogx/models/enum"
	"myblogx/service/log_service"
	"myblogx/test/testutil"
	"myblogx/utils/jwts"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func setupLogServiceEnv(t *testing.T) {
	t.Helper()
	_ = testutil.SetupSQLite(t, &models.UserModel{}, &models.UserConfModel{}, &models.LogModel{})
	global.Config = &conf.Config{
		Jwt: conf.Jwt{
			Expire: 1,
			Secret: "log-test-secret",
			Issuer: "test",
		},
	}
}

func newLogCtx(method, path, body string) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.RemoteAddr = "127.0.0.1:12345"
	req.Header.Set("Content-Type", "application/json")
	c.Request = req
	return c, w
}

func TestLoginLogCreate(t *testing.T) {
	setupLogServiceEnv(t)

	token, err := jwts.GetToken(jwts.Claims{
		UserID:   7,
		Role:     enum.RoleUser,
		Username: "alice",
	})
	if err != nil {
		t.Fatalf("生成 token 失败: %v", err)
	}

	successCtx, _ := newLogCtx(http.MethodPost, "/login", "")
	successCtx.Request.Header.Set("token", token)
	log_service.NewLoginSuccess(successCtx, enum.PasswordLoginType)

	failCtx, _ := newLogCtx(http.MethodPost, "/login", "")
	log_service.NewLoginFail(failCtx, enum.EmailLoginType, "密码错误", "bob", "123")

	var logs []models.LogModel
	if err := global.DB.Order("id asc").Find(&logs).Error; err != nil {
		t.Fatalf("查询登录日志失败: %v", err)
	}
	if len(logs) != 2 {
		t.Fatalf("日志数量错误: got=%d", len(logs))
	}

	if logs[0].LogType != enum.LoginLogType || !logs[0].LoginStatus || logs[0].UserID != 7 || logs[0].Username != "alice" {
		t.Fatalf("登录成功日志字段异常: %+v", logs[0])
	}
	if logs[1].LogType != enum.LoginLogType || logs[1].LoginStatus || logs[1].Username != "bob" || logs[1].Content != "密码错误" {
		t.Fatalf("登录失败日志字段异常: %+v", logs[1])
	}
}

func TestActionLogSaveAndMiddlewareSave(t *testing.T) {
	setupLogServiceEnv(t)

	token, err := jwts.GetToken(jwts.Claims{
		UserID:   9,
		Role:     enum.RoleAdmin,
		Username: "root",
	})
	if err != nil {
		t.Fatalf("生成 token 失败: %v", err)
	}

	c, _ := newLogCtx(http.MethodPost, "/articles", `{"title":"hello"}`)
	c.Request.Header.Set("token", token)
	c.Request.Header.Set("X-Trace-ID", "trace-1")
	c.Writer.Header().Set("X-Resp-ID", "resp-1")

	ac := log_service.NewActionLogByGin(c)
	ac.SetTitle("测试操作日志")
	ac.SetLevel(enum.LogWarnLevel)
	ac.SetShowRequest()
	ac.ShowRequestHeader()
	ac.SetRequest(c)
	ac.SetItemInfo("name", "alice")
	ac.SetItemWarn("step", 1)
	ac.SetItemError("error", "mock")
	ac.SetLink("doc", "https://example.com")
	ac.SetImage("https://example.com/a.png")
	ac.SetError("err", errors.New("boom"))

	id := ac.Save()
	if id == 0 {
		t.Fatal("首次保存日志应返回有效 id")
	}

	var first models.LogModel
	if err := global.DB.First(&first, id).Error; err != nil {
		t.Fatalf("查询操作日志失败: %v", err)
	}
	if first.UserID != 9 {
		t.Fatalf("解析 token user_id 失败: %+v", first)
	}
	if !strings.Contains(first.Content, "name") || !strings.Contains(first.Content, "step") {
		t.Fatalf("日志内容缺少关键字段: %s", first.Content)
	}

	ac.SetItemInfo("append", "ok")
	id2 := ac.Save()
	if id2 != id {
		t.Fatalf("重复保存应更新同一条日志: first=%d second=%d", id, id2)
	}

	var updated models.LogModel
	if err := global.DB.First(&updated, id).Error; err != nil {
		t.Fatalf("查询更新后日志失败: %v", err)
	}
	if !strings.Contains(updated.Content, "append") {
		t.Fatalf("更新内容未写入: %s", updated.Content)
	}

	noSaveCtx, _ := newLogCtx(http.MethodGet, "/m", "")
	noSaveCtx.Set("isSaveLog", false)
	noSave := log_service.NewActionLogByGin(noSaveCtx)
	noSave.SetTitle("middleware-no-save")
	noSave.MiddlewareSave()

	var cnt int64
	_ = global.DB.Model(&models.LogModel{}).Count(&cnt).Error
	if cnt != 1 {
		t.Fatalf("isSaveLog=false 不应新增日志, got=%d", cnt)
	}

	saveCtx, _ := newLogCtx(http.MethodGet, "/m", "")
	saveCtx.Set("isSaveLog", true)
	saveCtx.Writer.Header().Set("X-Mid", "1")

	midLog := log_service.NewActionLogByGin(saveCtx)
	midLog.SetTitle("middleware-save")
	midLog.SetShowResponse()
	midLog.ShowResponseHeader()
	midLog.SetResponseHeader(saveCtx)
	midLog.SetResponse([]byte(`{"ok":true}`))
	midLog.MiddlewareSave()

	_ = global.DB.Model(&models.LogModel{}).Count(&cnt).Error
	if cnt != 2 {
		t.Fatalf("isSaveLog=true 应新增日志, got=%d", cnt)
	}

	midLog.SetItemInfo("next", "round")
	midLog.MiddlewareSave()

	var middlewareLog models.LogModel
	if err := global.DB.Order("id desc").First(&middlewareLog).Error; err != nil {
		t.Fatalf("查询中间件日志失败: %v", err)
	}
	if !strings.Contains(middlewareLog.Content, `{"ok":true}`) {
		t.Fatalf("中间件响应体未写入日志: %s", middlewareLog.Content)
	}
}

func TestGetLog(t *testing.T) {
	setupLogServiceEnv(t)

	c, _ := newLogCtx(http.MethodGet, "/x", "")

	got := log_service.GetLog(c)
	if got == nil {
		t.Fatal("GetLog 在未设置 context 时不应返回 nil")
	}

	c.Set("log", "invalid")
	got = log_service.GetLog(c)
	if got == nil {
		t.Fatal("GetLog 在类型错误时不应返回 nil")
	}

	expect := log_service.NewActionLogByGin(c)
	c.Set("log", expect)
	got = log_service.GetLog(c)
	if got != expect {
		t.Fatal("GetLog 未返回 context 中已有的日志对象")
	}

	v, ok := c.Get("isSaveLog")
	if !ok || v != true {
		t.Fatalf("GetLog 应设置 isSaveLog=true, got=%v ok=%v", v, ok)
	}
}
