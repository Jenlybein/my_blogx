package log_api_test

import (
	"encoding/json"
	"myblogx/api/log_api"
	"myblogx/common"
	"myblogx/models"
	"myblogx/models/enum"
	"myblogx/test/testutil"
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

func TestLogListReadRemove(t *testing.T) {
	db := testutil.SetupSQLite(t, &models.UserModel{}, &models.UserConfModel{}, &models.LogModel{})

	user := models.UserModel{Username: "u1", Password: "x", Nickname: "nick"}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("创建用户失败: %v", err)
	}
	log := models.LogModel{
		LogType: enum.ActionLogType,
		Title:   "t1",
		UserID:  user.ID,
		Level:   enum.LogInfoLevel,
	}
	if err := db.Create(&log).Error; err != nil {
		t.Fatalf("创建日志失败: %v", err)
	}

	api := &log_api.LogApi{}

	{
		c, w := newCtx()
		c.Set("requestQuery", log_api.LogListRequest{
			PageInfo: common.PageInfo{Page: 1, Limit: 10},
			UserID:   user.ID,
		})
		api.LogListView(c)
		if code := readCode(t, w); code != 0 {
			t.Fatalf("日志列表失败, code=%d body=%s", code, w.Body.String())
		}
	}

	{
		c, w := newCtx()
		c.Set("requestUri", models.IDRequest{ID: log.ID})
		api.LogReadView(c)
		if code := readCode(t, w); code != 0 {
			t.Fatalf("日志读取失败, code=%d body=%s", code, w.Body.String())
		}
		var got models.LogModel
		_ = db.First(&got, log.ID).Error
		if !got.IsRead {
			t.Fatal("日志读取后应置为已读")
		}
	}

	{
		c, w := newCtx()
		c.Set("requestJson", models.IDListRequest{IDList: []uint{log.ID}})
		api.LogRemoveView(c)
		if code := readCode(t, w); code != 0 {
			t.Fatalf("日志删除失败, code=%d body=%s", code, w.Body.String())
		}
	}
}
