package router_test

import (
	"myblogx/global"
	"myblogx/models"
	"myblogx/models/enum/message_enum"
	"myblogx/test/testutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSitemsgRouterDeleteSupportsRemoveByID(t *testing.T) {
	user, token := setupSitemsgRouterEnv(t)
	db := global.DB

	msg := models.ArticleMessageModel{
		ReceiverID: user.ID,
		Type:       message_enum.SystemType,
		Content:    "system",
	}
	if err := db.Create(&msg).Error; err != nil {
		t.Fatalf("创建消息失败: %v", err)
	}

	req := testutil.NewJSONRequest(http.MethodDelete, "/api/sitemsg", `{"id":1}`)
	req.Header.Set("token", token)

	w := httptest.NewRecorder()
	newSitemsgRouterEngine().ServeHTTP(w, req)

	if w.Code != http.StatusOK || readBizCode(t, w) != 0 {
		t.Fatalf("按 id 删除消息失败, body=%s", w.Body.String())
	}

	var count int64
	if err := db.Model(&models.ArticleMessageModel{}).Where("id = ?", msg.ID).Count(&count).Error; err != nil {
		t.Fatalf("查询消息失败: %v", err)
	}
	if count != 0 {
		t.Fatalf("消息未删除, count=%d", count)
	}
}

func TestSitemsgRouterDeleteSupportsBatchRemoveByType(t *testing.T) {
	user, token := setupSitemsgRouterEnv(t)
	db := global.DB

	msgs := []models.ArticleMessageModel{
		{ReceiverID: user.ID, Type: message_enum.DiggArticleType, Content: "d1"},
		{ReceiverID: user.ID, Type: message_enum.FavorArticleType, Content: "f1"},
		{ReceiverID: user.ID, Type: message_enum.SystemType, Content: "s1"},
	}
	if err := db.Create(&msgs).Error; err != nil {
		t.Fatalf("创建消息失败: %v", err)
	}

	req := testutil.NewJSONRequest(http.MethodDelete, "/api/sitemsg", `{"t":2}`)
	req.Header.Set("token", token)

	w := httptest.NewRecorder()
	newSitemsgRouterEngine().ServeHTTP(w, req)

	if w.Code != http.StatusOK || readBizCode(t, w) != 0 {
		t.Fatalf("按类型批量删除消息失败, body=%s", w.Body.String())
	}

	var remain []models.ArticleMessageModel
	if err := db.Order("id asc").Find(&remain).Error; err != nil {
		t.Fatalf("查询消息失败: %v", err)
	}
	if len(remain) != 1 || remain[0].Type != message_enum.SystemType {
		t.Fatalf("批量删除范围异常: %+v", remain)
	}
}
