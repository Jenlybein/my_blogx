package image_api_test

import (
	"encoding/json"
	"myblogx/api/image_api"
	"myblogx/common"
	"myblogx/models"
	"myblogx/test/testutil"
	"net/http/httptest"
	"os"
	"path/filepath"
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

func TestImageListAndRemove(t *testing.T) {
	db := testutil.SetupSQLite(t, &models.ImageModel{})

	tmp := t.TempDir()
	filePath := filepath.Join(tmp, "img.png")
	if err := os.WriteFile(filePath, []byte("x"), 0644); err != nil {
		t.Fatalf("写测试文件失败: %v", err)
	}

	m := models.ImageModel{
		FileName: "img.png",
		Path:     filePath,
		Size:     1,
		Hash:     "abc",
	}
	if err := db.Create(&m).Error; err != nil {
		t.Fatalf("创建图片记录失败: %v", err)
	}

	api := image_api.ImageApi{}
	{
		c, w := newCtx()
		c.Set("requestQuery", common.PageInfo{Page: 1, Limit: 10})
		api.ImageListView(c)
		if code := readCode(t, w); code != 0 {
			t.Fatalf("图片列表失败, code=%d body=%s", code, w.Body.String())
		}
	}

	{
		c, w := newCtx()
		c.Set("requestJson", models.RemoveRequest{IDList: []uint{m.ID}})
		// 缺省 log 时 GetLog 会创建临时日志对象，满足调用路径
		api.ImageRemoveView(c)
		if code := readCode(t, w); code != 0 {
			t.Fatalf("图片删除失败, code=%d body=%s", code, w.Body.String())
		}
	}
}
