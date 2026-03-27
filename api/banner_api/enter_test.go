package banner_api_test

import (
	"encoding/json"
	"myblogx/api/banner_api"
	"myblogx/global"
	"myblogx/models"
	"myblogx/models/ctype"
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

func TestBannerCreateListUpdateRemove(t *testing.T) {
	db := testutil.SetupSQLite(t, &models.BannerModel{}, &models.ImageRefModel{})

	api := banner_api.BannerApi{}

	{
		c, w := newCtx()
		c.Set("requestJson", banner_api.BannerCreateRequest{
			Cover: "/a.png",
			Href:  "/a",
			Show:  true,
		})
		api.BannerCreateView(c)
		if code := readCode(t, w); code != 0 {
			t.Fatalf("创建失败, code=%d body=%s", code, w.Body.String())
		}
	}

	var created models.BannerModel
	if err := db.First(&created).Error; err != nil {
		t.Fatalf("查询创建数据失败: %v", err)
	}

	{
		c, w := newCtx()
		c.Set("requestQuery", banner_api.BannerListRequest{
			Show: true,
		})
		api.BannerListView(c)
		if code := readCode(t, w); code != 0 {
			t.Fatalf("列表失败, code=%d body=%s", code, w.Body.String())
		}
	}

	{
		c, w := newCtx()
		c.Set("requestUri", models.IDRequest{ID: created.ID})
		c.Set("requestJson", banner_api.BannerCreateRequest{
			Cover: "/b.png",
			Href:  "/b",
			Show:  false,
		})
		api.BannerUpdateView(c)
		if code := readCode(t, w); code != 0 {
			t.Fatalf("更新失败, code=%d body=%s", code, w.Body.String())
		}
	}

	{
		c, w := newCtx()
		c.Set("requestJson", models.IDListRequest{IDList: []ctype.ID{created.ID}})
		api.BannerRemoveView(c)
		if code := readCode(t, w); code != 0 {
			t.Fatalf("删除失败, code=%d body=%s", code, w.Body.String())
		}
	}

	var cnt int64
	_ = global.DB.Model(&models.BannerModel{}).Count(&cnt).Error
	if cnt != 0 {
		t.Fatalf("删除后数量异常: %d", cnt)
	}
}
