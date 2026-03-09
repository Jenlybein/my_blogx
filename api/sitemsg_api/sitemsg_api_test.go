package sitemsg_api_test

import (
	"encoding/json"
	"myblogx/api/sitemsg_api"
	"myblogx/global"
	"myblogx/models"
	"myblogx/models/enum"
	"myblogx/models/enum/message_enum"
	"myblogx/test/testutil"
	"myblogx/utils/jwts"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func newSitemsgCtx() (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	return c, w
}

func readSitemsgBody(t *testing.T, w *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("解析响应失败: %v, body=%s", err, w.Body.String())
	}
	return body
}

func readSitemsgCode(t *testing.T, w *httptest.ResponseRecorder) int {
	t.Helper()
	return int(readSitemsgBody(t, w)["code"].(float64))
}

func setupSitemsgEnv(t *testing.T) *models.UserModel {
	t.Helper()
	db := testutil.SetupSQLite(t,
		&models.UserModel{},
		&models.UserConfModel{},
		&models.ArticleMessageModel{},
	)

	user := &models.UserModel{
		Username: "msg_user",
		Password: "x",
		Role:     enum.RoleUser,
	}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("创建用户失败: %v", err)
	}
	return user
}

func setClaims(c *gin.Context, user *models.UserModel) {
	c.Set("claims", &jwts.MyClaims{
		Claims: jwts.Claims{
			UserID:   user.ID,
			Role:     user.Role,
			Username: user.Username,
		},
	})
}

func TestUserMsgConfViewAndUpdate(t *testing.T) {
	user := setupSitemsgEnv(t)
	api := sitemsg_api.SitemsgApi{}

	{
		c, w := newSitemsgCtx()
		setClaims(c, user)
		api.UserMsgConfView(c)
		if code := readSitemsgCode(t, w); code != 0 {
			t.Fatalf("查询消息配置失败, body=%s", w.Body.String())
		}

		data := readSitemsgBody(t, w)["data"].(map[string]any)
		if data["digg_notice_enabled"] != true || data["private_chat_notice_enabled"] != true {
			t.Fatalf("默认消息配置异常, body=%s", w.Body.String())
		}
	}

	{
		c, w := newSitemsgCtx()
		setClaims(c, user)
		c.Set("requestJson", sitemsg_api.UserMsgConfResponseAndRequest{
			DiggNoticeEnabled:        false,
			CommentNoticeEnabled:     false,
			FavorNoticeEnabled:       false,
			PrivateChatNoticeEnabled: false,
		})
		api.UserMsgConfUpdateView(c)
		if code := readSitemsgCode(t, w); code != 0 {
			t.Fatalf("更新消息配置失败, body=%s", w.Body.String())
		}
	}

	var conf models.UserConfModel
	if err := global.DB.Take(&conf, user.ID).Error; err != nil {
		t.Fatalf("查询配置失败: %v", err)
	}
	if conf.DiggNoticeEnabled || conf.CommentNoticeEnabled || conf.FavorNoticeEnabled || conf.PrivateChatNoticeEnabled {
		t.Fatalf("消息配置未更新: %+v", conf)
	}
}

func TestSitemsgListViewFiltersByType(t *testing.T) {
	user := setupSitemsgEnv(t)
	db := global.DB
	api := sitemsg_api.SitemsgApi{}

	other := &models.UserModel{
		Username: "other_user",
		Password: "x",
		Role:     enum.RoleUser,
	}
	if err := db.Create(other).Error; err != nil {
		t.Fatalf("创建其他用户失败: %v", err)
	}

	msgs := []models.ArticleMessageModel{
		{ReceiverID: user.ID, Type: message_enum.CommentArticleType, Content: "c1"},
		{ReceiverID: user.ID, Type: message_enum.CommentReplyType, Content: "c2"},
		{ReceiverID: user.ID, Type: message_enum.DiggArticleType, Content: "d1"},
		{ReceiverID: user.ID, Type: message_enum.SystemType, Content: "s1"},
		{ReceiverID: other.ID, Type: message_enum.CommentArticleType, Content: "other"},
	}
	if err := db.Create(&msgs).Error; err != nil {
		t.Fatalf("创建消息失败: %v", err)
	}

	{
		c, w := newSitemsgCtx()
		setClaims(c, user)
		c.Set("requestQuery", sitemsg_api.SitemsgListRequest{T: 1})
		api.SitemsgListView(c)
		if code := readSitemsgCode(t, w); code != 0 {
			t.Fatalf("查询评论消息失败, body=%s", w.Body.String())
		}
		data := readSitemsgBody(t, w)["data"].(map[string]any)
		if int(data["count"].(float64)) != 2 {
			t.Fatalf("评论消息数量异常, body=%s", w.Body.String())
		}
	}

	{
		c, w := newSitemsgCtx()
		setClaims(c, user)
		c.Set("requestQuery", sitemsg_api.SitemsgListRequest{T: 3})
		api.SitemsgListView(c)
		if code := readSitemsgCode(t, w); code != 0 {
			t.Fatalf("查询系统消息失败, body=%s", w.Body.String())
		}
		data := readSitemsgBody(t, w)["data"].(map[string]any)
		if int(data["count"].(float64)) != 1 {
			t.Fatalf("系统消息数量异常, body=%s", w.Body.String())
		}
	}
}

func TestSitemsgReadViewSingleAndBatch(t *testing.T) {
	user := setupSitemsgEnv(t)
	db := global.DB
	api := sitemsg_api.SitemsgApi{}

	other := &models.UserModel{
		Username: "msg_other",
		Password: "x",
		Role:     enum.RoleUser,
	}
	if err := db.Create(other).Error; err != nil {
		t.Fatalf("创建其他用户失败: %v", err)
	}

	single := models.ArticleMessageModel{ReceiverID: user.ID, Type: message_enum.SystemType, Content: "single"}
	batchA := models.ArticleMessageModel{ReceiverID: user.ID, Type: message_enum.DiggArticleType, Content: "batch-a"}
	batchB := models.ArticleMessageModel{ReceiverID: user.ID, Type: message_enum.FavorArticleType, Content: "batch-b"}
	otherUserMsg := models.ArticleMessageModel{ReceiverID: other.ID, Type: message_enum.DiggArticleType, Content: "other"}
	if err := db.Create(&[]models.ArticleMessageModel{single, batchA, batchB, otherUserMsg}).Error; err != nil {
		t.Fatalf("创建消息失败: %v", err)
	}

	var list []models.ArticleMessageModel
	if err := db.Order("id asc").Find(&list).Error; err != nil {
		t.Fatalf("查询消息失败: %v", err)
	}
	single = list[0]
	batchA = list[1]
	batchB = list[2]
	otherUserMsg = list[3]

	{
		c, w := newSitemsgCtx()
		setClaims(c, user)
		c.Set("requestJson", sitemsg_api.SitemsgReadRequest{})
		api.SitemsgReadView(c)
		if code := readSitemsgCode(t, w); code == 0 {
			t.Fatalf("id 和 t 同时为空时应失败, body=%s", w.Body.String())
		}
	}

	{
		c, w := newSitemsgCtx()
		setClaims(c, user)
		c.Set("requestJson", sitemsg_api.SitemsgReadRequest{ID: single.ID})
		api.SitemsgReadView(c)
		if code := readSitemsgCode(t, w); code != 0 {
			t.Fatalf("单条已读失败, body=%s", w.Body.String())
		}
	}

	var singleCheck models.ArticleMessageModel
	if err := db.Take(&singleCheck, single.ID).Error; err != nil {
		t.Fatalf("查询单条消息失败: %v", err)
	}
	if !singleCheck.IsRead {
		t.Fatalf("单条消息未标记已读: %+v", singleCheck)
	}
	if singleCheck.ReadAt == nil {
		t.Fatalf("单条消息未写入 read_at: %+v", singleCheck)
	}

	{
		c, w := newSitemsgCtx()
		setClaims(c, user)
		c.Set("requestJson", sitemsg_api.SitemsgReadRequest{ID: single.ID})
		api.SitemsgReadView(c)
		if code := readSitemsgCode(t, w); code == 0 {
			t.Fatalf("重复标记已读应失败, body=%s", w.Body.String())
		}
	}

	{
		c, w := newSitemsgCtx()
		setClaims(c, user)
		c.Set("requestJson", sitemsg_api.SitemsgReadRequest{T: 2})
		api.SitemsgReadView(c)
		if code := readSitemsgCode(t, w); code != 0 {
			t.Fatalf("批量已读失败, body=%s", w.Body.String())
		}
	}

	var batchCheck []models.ArticleMessageModel
	if err := db.Where("id in ?", []uint{batchA.ID, batchB.ID, otherUserMsg.ID}).Order("id asc").Find(&batchCheck).Error; err != nil {
		t.Fatalf("查询批量消息失败: %v", err)
	}
	if !batchCheck[0].IsRead || !batchCheck[1].IsRead {
		t.Fatalf("目标消息未全部标记已读: %+v", batchCheck)
	}
	if batchCheck[0].ReadAt == nil || batchCheck[1].ReadAt == nil {
		t.Fatalf("批量消息未写入 read_at: %+v", batchCheck)
	}
	if batchCheck[2].IsRead {
		t.Fatalf("其他用户消息不应被标记已读: %+v", batchCheck[2])
	}
	if batchCheck[2].ReadAt != nil {
		t.Fatalf("其他用户消息不应更新 read_at: %+v", batchCheck[2])
	}
}
