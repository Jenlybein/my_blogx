package chat_api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"myblogx/common"
	"myblogx/global"
	"myblogx/models"
	"myblogx/models/enum"
	"myblogx/models/enum/chat_msg_enum"
	"myblogx/test/testutil"
	"myblogx/utils/jwts"

	"github.com/gin-gonic/gin"
)

type chatListTestResponse struct {
	Code int             `json:"code"`
	Data json.RawMessage `json:"data"`
	Msg  string          `json:"msg"`
}

type chatListPayload struct {
	List  []ChatSessionListResponse `json:"list"`
	Count int                       `json:"count"`
}

type chatMsgListPayload struct {
	List  []ChatMsgListResponse `json:"list"`
	Count int                   `json:"count"`
}

func TestChatSessionListView(t *testing.T) {
	api := ChatApi{}
	users := setupChatListEnv(t)

	now := time.Date(2026, 3, 12, 9, 0, 0, 0, time.Local)
	rows := []models.ChatSessionModel{
		{
			SessionID:      "chat:1:2",
			UserID:         users.owner.ID,
			ReceiverID:     users.friendA.ID,
			LastMsgID:      101,
			LastMsgContent: "first",
			LastMsgTime:    now.Add(-2 * time.Hour),
			UnreadCount:    3,
		},
		{
			SessionID:      "chat:1:3",
			UserID:         users.owner.ID,
			ReceiverID:     users.friendB.ID,
			LastMsgID:      102,
			LastMsgContent: "top",
			LastMsgTime:    now.Add(-3 * time.Hour),
			IsTop:          true,
		},
		{
			SessionID:      "chat:4:5",
			UserID:         users.other.ID,
			ReceiverID:     users.friendA.ID,
			LastMsgID:      103,
			LastMsgContent: "other",
			LastMsgTime:    now,
		},
	}
	if err := global.DB.Create(&rows).Error; err != nil {
		t.Fatalf("创建会话数据失败: %v", err)
	}

	c, w := newChatListCtx(t, users.owner, ChatSessionListRequest{
		PageInfo: common.PageInfo{Page: 1, Limit: 10},
	})
	api.ChatSessionListView(c)

	resp := readChatListResponse(t, w)
	if resp.Code != 0 {
		t.Fatalf("chat_list 应成功, body=%s", w.Body.String())
	}
	if resp.Data.Count != 2 {
		t.Fatalf("会话数量错误: %d", resp.Data.Count)
	}
	if len(resp.Data.List) != 2 {
		t.Fatalf("返回列表长度错误: %d", len(resp.Data.List))
	}

	if resp.Data.List[0].ReceiverID != users.friendB.ID {
		t.Fatalf("置顶会话应排第一: %+v", resp.Data.List[0])
	}
	if resp.Data.List[0].ReceiverNickname != users.friendB.Nickname {
		t.Fatalf("应带出对端昵称: %+v", resp.Data.List[0])
	}

	if resp.Data.List[1].ReceiverID != users.friendA.ID {
		t.Fatalf("非置顶会话顺序错误: %+v", resp.Data.List[1])
	}
	if resp.Data.List[1].UnreadCount != 3 {
		t.Fatalf("未读数错误: %+v", resp.Data.List[1])
	}
	expectTime := time.Date(2026, 3, 12, 7, 0, 0, 0, time.Local)
	if !resp.Data.List[1].LastMsgTime.Equal(expectTime) {
		t.Fatalf("时间错误: %v", resp.Data.List[1].LastMsgTime)
	}
}

func TestChatMsgListView(t *testing.T) {
	api := ChatApi{}
	users := setupChatListEnv(t)
	sessionID := "chat:1:2"
	sendTimeA := time.Date(2026, 3, 12, 8, 0, 0, 0, time.Local)
	sendTimeB := time.Date(2026, 3, 12, 9, 0, 0, 0, time.Local)

	sessions := []models.ChatSessionModel{
		{SessionID: sessionID, UserID: users.owner.ID, ReceiverID: users.friendA.ID},
		{SessionID: sessionID, UserID: users.friendA.ID, ReceiverID: users.owner.ID},
	}
	if err := global.DB.Create(&sessions).Error; err != nil {
		t.Fatalf("创建会话失败: %v", err)
	}

	readAt := sendTimeB.Add(10 * time.Minute)
	msgs := []models.ChatMsgModel{
		{
			SessionID:  sessionID,
			SenderID:   users.owner.ID,
			ReceiverID: users.friendA.ID,
			Content:    "old",
			SendTime:   sendTimeA,
			MsgType:    chat_msg_enum.MsgTypeText,
			MsgStatus:  chat_msg_enum.MsgStatusSend,
		},
		{
			SessionID:  sessionID,
			SenderID:   users.friendA.ID,
			ReceiverID: users.owner.ID,
			Content:    "new",
			SendTime:   sendTimeB,
			ReadAt:     &readAt,
			MsgType:    chat_msg_enum.MsgTypeText,
			MsgStatus:  chat_msg_enum.MsgStatusRead,
		},
		{
			SessionID:  "chat:3:4",
			SenderID:   users.other.ID,
			ReceiverID: users.friendB.ID,
			Content:    "other",
			SendTime:   sendTimeB,
			MsgType:    chat_msg_enum.MsgTypeText,
			MsgStatus:  chat_msg_enum.MsgStatusSend,
		},
	}
	if err := global.DB.Create(&msgs).Error; err != nil {
		t.Fatalf("创建消息失败: %v", err)
	}

	c, w := newChatMsgListCtx(t, users.owner, ChatMsgListRequest{
		PageInfo:  common.PageInfo{Page: 1, Limit: 10},
		SessionID: sessionID,
	})
	api.ChatMsgListView(c)

	resp := readChatMsgListResponse(t, w)
	if resp.Code != 0 {
		t.Fatalf("chat_msg_list 应成功, body=%s", w.Body.String())
	}
	if resp.Data.Count != 2 || len(resp.Data.List) != 2 {
		t.Fatalf("消息列表数量错误: %+v", resp.Data)
	}
	if resp.Data.List[0].Content != "new" {
		t.Fatalf("应按时间倒序返回最新消息: %+v", resp.Data.List[0])
	}
	if resp.Data.List[0].IsSelf {
		t.Fatalf("对方消息不应标记为自己发送: %+v", resp.Data.List[0])
	}
	if !resp.Data.List[0].IsRead {
		t.Fatalf("已读消息应标记 IsRead=true: %+v", resp.Data.List[0])
	}
	if !resp.Data.List[1].IsSelf {
		t.Fatalf("自己发送的消息应标记 IsSelf=true: %+v", resp.Data.List[1])
	}
	if resp.Data.List[1].IsRead {
		t.Fatalf("未读消息不应标记 IsRead=true: %+v", resp.Data.List[1])
	}
}

type chatUsers struct {
	owner   models.UserModel
	friendA models.UserModel
	friendB models.UserModel
	other   models.UserModel
}

func setupChatListEnv(t *testing.T) chatUsers {
	t.Helper()
	testutil.SetupSQLite(t, &models.UserModel{}, &models.UserConfModel{}, &models.ChatSessionModel{}, &models.ChatMsgModel{})

	return chatUsers{
		owner:   createChatUser(t, "chat_owner"),
		friendA: createChatUser(t, "chat_friend_a"),
		friendB: createChatUser(t, "chat_friend_b"),
		other:   createChatUser(t, "chat_other"),
	}
}

func createChatUser(t *testing.T, username string) models.UserModel {
	t.Helper()
	user := models.UserModel{
		Username: username,
		Nickname: username + "_nick",
		Avatar:   username + ".png",
		Abstract: username + "_abstract",
	}
	if err := global.DB.Create(&user).Error; err != nil {
		t.Fatalf("创建用户失败: %v", err)
	}
	return user
}

func newChatListCtx(t *testing.T, user models.UserModel, query ChatSessionListRequest) (*gin.Context, *httptest.ResponseRecorder) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/chat", nil)
	c.Set("requestQuery", query)
	c.Set("claims", &jwts.MyClaims{
		Claims: jwts.Claims{
			UserID:   user.ID,
			Role:     enum.RoleUser,
			Username: user.Username,
		},
	})
	return c, w
}

func newChatMsgListCtx(t *testing.T, user models.UserModel, query ChatMsgListRequest) (*gin.Context, *httptest.ResponseRecorder) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/chat/msg", nil)
	c.Set("requestQuery", query)
	c.Set("claims", &jwts.MyClaims{
		Claims: jwts.Claims{
			UserID:   user.ID,
			Role:     enum.RoleUser,
			Username: user.Username,
		},
	})
	return c, w
}

func readChatListResponse(t *testing.T, w *httptest.ResponseRecorder) struct {
	Code int
	Data chatListPayload
	Msg  string
} {
	t.Helper()

	var resp chatListTestResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("解析响应失败: %v body=%s", err, w.Body.String())
	}

	var payload chatListPayload
	if len(resp.Data) > 0 {
		if err := json.Unmarshal(resp.Data, &payload); err != nil {
			t.Fatalf("解析列表载荷失败: %v body=%s", err, w.Body.String())
		}
	}

	return struct {
		Code int
		Data chatListPayload
		Msg  string
	}{
		Code: resp.Code,
		Data: payload,
		Msg:  resp.Msg,
	}
}

func readChatMsgListResponse(t *testing.T, w *httptest.ResponseRecorder) struct {
	Code int
	Data chatMsgListPayload
	Msg  string
} {
	t.Helper()

	var resp chatListTestResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("解析响应失败: %v body=%s", err, w.Body.String())
	}

	var payload chatMsgListPayload
	if len(resp.Data) > 0 {
		if err := json.Unmarshal(resp.Data, &payload); err != nil {
			t.Fatalf("解析消息列表载荷失败: %v body=%s", err, w.Body.String())
		}
	}

	return struct {
		Code int
		Data chatMsgListPayload
		Msg  string
	}{
		Code: resp.Code,
		Data: payload,
		Msg:  resp.Msg,
	}
}
