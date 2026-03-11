package chat_service

import (
	"encoding/json"
	"testing"
	"time"

	"myblogx/global"
	"myblogx/models"
	"myblogx/models/enum/chat_msg_enum"
	"myblogx/test/testutil"
)

func setupChatServiceTestDB(t *testing.T) (*models.UserModel, *models.UserModel) {
	t.Helper()

	db := testutil.SetupSQLite(
		t,
		&models.UserModel{},
		&models.UserConfModel{},
		&models.ChatSessionModel{},
		&models.ChatMsgModel{},
	)

	userA := &models.UserModel{Username: "u1", Nickname: "u1"}
	userB := &models.UserModel{Username: "u2", Nickname: "u2"}
	if err := db.Create(userA).Error; err != nil {
		t.Fatalf("创建用户 A 失败: %v", err)
	}
	if err := db.Create(userB).Error; err != nil {
		t.Fatalf("创建用户 B 失败: %v", err)
	}
	return userA, userB
}

func mustGetSession(t *testing.T, userID, receiverID uint) models.ChatSessionModel {
	t.Helper()

	var session models.ChatSessionModel
	if err := global.DB.Take(&session, "user_id = ? and receiver_id = ?", userID, receiverID).Error; err != nil {
		t.Fatalf("查询会话失败 user=%d receiver=%d: %v", userID, receiverID, err)
	}
	return session
}

func TestValidateChatBase(t *testing.T) {
	t.Run("默认值补齐", func(t *testing.T) {
		req := ToChatRequest{
			SenderID:   1,
			ReceiverID: 2,
			MsgType:    chat_msg_enum.MsgTypeText,
			Content:    "hello",
		}

		if err := validateChatBase(&req); err != nil {
			t.Fatalf("validateChatBase 返回错误: %v", err)
		}
		if req.SendTime.IsZero() {
			t.Fatal("应补齐默认发送时间")
		}
		if req.MsgStatus != chat_msg_enum.MsgStatusSend {
			t.Fatalf("默认消息状态错误: %v", req.MsgStatus)
		}
	})

	t.Run("禁止给自己发私信", func(t *testing.T) {
		req := ToChatRequest{
			SenderID:   1,
			ReceiverID: 1,
			MsgType:    chat_msg_enum.MsgTypeText,
			Content:    "hello",
			SendTime:   time.Now(),
		}

		err := validateChatBase(&req)
		if err == nil || err.Error() != "不支持给自己发私信" {
			t.Fatalf("应返回自发消息错误, got=%v", err)
		}
	})
}

func TestToTextChatCreatesMessageAndSessions(t *testing.T) {
	userA, userB := setupChatServiceTestDB(t)
	sendTime := time.Date(2026, 3, 12, 10, 0, 0, 0, time.Local)

	msg, err := ToTextChat(ToTextChatRequest{
		SenderID:   userA.ID,
		ReceiverID: userB.ID,
		Text:       "hello",
		SendTime:   sendTime,
	})
	if err != nil {
		t.Fatalf("发送文本消息失败: %v", err)
	}

	if msg.MsgType != chat_msg_enum.MsgTypeText {
		t.Fatalf("消息类型错误: %v", msg.MsgType)
	}
	if msg.MsgStatus != chat_msg_enum.MsgStatusSend {
		t.Fatalf("消息状态错误: %v", msg.MsgStatus)
	}
	if msg.SessionID != buildSessionID(userA.ID, userB.ID) {
		t.Fatalf("session_id 错误: %s", msg.SessionID)
	}
	if !msg.SendTime.Equal(sendTime) {
		t.Fatalf("发送时间错误: %v", msg.SendTime)
	}

	var msgCount int64
	if err = global.DB.Model(&models.ChatMsgModel{}).Count(&msgCount).Error; err != nil {
		t.Fatalf("统计消息失败: %v", err)
	}
	if msgCount != 1 {
		t.Fatalf("消息数量错误: %d", msgCount)
	}

	var sessionCount int64
	if err = global.DB.Model(&models.ChatSessionModel{}).Count(&sessionCount).Error; err != nil {
		t.Fatalf("统计会话失败: %v", err)
	}
	if sessionCount != 2 {
		t.Fatalf("会话数量错误: %d", sessionCount)
	}

	senderSession := mustGetSession(t, userA.ID, userB.ID)
	receiverSession := mustGetSession(t, userB.ID, userA.ID)

	if senderSession.SessionID != msg.SessionID || receiverSession.SessionID != msg.SessionID {
		t.Fatal("双方会话应共享相同 session_id")
	}
	if senderSession.UnreadCount != 0 {
		t.Fatalf("发送方未读数错误: %d", senderSession.UnreadCount)
	}
	if receiverSession.UnreadCount != 1 {
		t.Fatalf("接收方未读数错误: %d", receiverSession.UnreadCount)
	}
	if senderSession.LastMsgContent != "hello..." {
		t.Fatalf("发送方最后消息摘要错误: %s", senderSession.LastMsgContent)
	}
	if receiverSession.LastMsgContent != "hello..." {
		t.Fatalf("接收方最后消息摘要错误: %s", receiverSession.LastMsgContent)
	}
}

func TestToTextChatReusesSessions(t *testing.T) {
	userA, userB := setupChatServiceTestDB(t)

	_, err := ToTextChat(ToTextChatRequest{
		SenderID:   userA.ID,
		ReceiverID: userB.ID,
		Text:       "first",
		SendTime:   time.Date(2026, 3, 12, 10, 0, 0, 0, time.Local),
	})
	if err != nil {
		t.Fatalf("第一次发送失败: %v", err)
	}

	secondMsg, err := ToTextChat(ToTextChatRequest{
		SenderID:   userA.ID,
		ReceiverID: userB.ID,
		Text:       "second",
		SendTime:   time.Date(2026, 3, 12, 11, 0, 0, 0, time.Local),
	})
	if err != nil {
		t.Fatalf("第二次发送失败: %v", err)
	}

	var sessionCount int64
	if err = global.DB.Model(&models.ChatSessionModel{}).Count(&sessionCount).Error; err != nil {
		t.Fatalf("统计会话失败: %v", err)
	}
	if sessionCount != 2 {
		t.Fatalf("重复发送后会话数量错误: %d", sessionCount)
	}

	senderSession := mustGetSession(t, userA.ID, userB.ID)
	receiverSession := mustGetSession(t, userB.ID, userA.ID)

	if senderSession.LastMsgID != secondMsg.ID || receiverSession.LastMsgID != secondMsg.ID {
		t.Fatalf("最后消息 ID 未更新: sender=%d receiver=%d msg=%d", senderSession.LastMsgID, receiverSession.LastMsgID, secondMsg.ID)
	}
	if receiverSession.UnreadCount != 2 {
		t.Fatalf("接收方未读数应累加到 2, got=%d", receiverSession.UnreadCount)
	}
	if senderSession.UnreadCount != 0 {
		t.Fatalf("发送方未读数应保持 0, got=%d", senderSession.UnreadCount)
	}
}

func TestToImageChatStoresJSONAndUpdatesSession(t *testing.T) {
	userA, userB := setupChatServiceTestDB(t)

	msg, err := ToImageChat(ToImageChatRequest{
		SenderID:    userA.ID,
		ReceiverID:  userB.ID,
		ImageURL:    "https://cdn.example.com/image.png",
		PreviewURL:  "https://cdn.example.com/image_small.png",
		FileName:    "image.png",
		MimeType:    "image/png",
		Width:       100,
		Height:      200,
		Size:        4096,
		Alt:         "demo",
		OriginalURL: "https://cdn.example.com/image_origin.png",
		SendTime:    time.Date(2026, 3, 12, 12, 0, 0, 0, time.Local),
	})
	if err != nil {
		t.Fatalf("发送图片消息失败: %v", err)
	}

	if msg.MsgType != chat_msg_enum.MsgTypeImage {
		t.Fatalf("消息类型错误: %v", msg.MsgType)
	}

	var payload imageChatContent
	if err = json.Unmarshal([]byte(msg.Content), &payload); err != nil {
		t.Fatalf("图片消息内容不是合法 JSON: %v", err)
	}
	if payload.ImageURL != "https://cdn.example.com/image.png" {
		t.Fatalf("图片地址错误: %s", payload.ImageURL)
	}
	if payload.PreviewURL != "https://cdn.example.com/image_small.png" {
		t.Fatalf("缩略图地址错误: %s", payload.PreviewURL)
	}

	senderSession := mustGetSession(t, userA.ID, userB.ID)
	receiverSession := mustGetSession(t, userB.ID, userA.ID)

	if senderSession.LastMsgContent != "[图片]" || receiverSession.LastMsgContent != "[图片]" {
		t.Fatalf("图片消息摘要错误: sender=%s receiver=%s", senderSession.LastMsgContent, receiverSession.LastMsgContent)
	}
	if receiverSession.UnreadCount != 1 {
		t.Fatalf("图片消息应给接收方增加 1 条未读, got=%d", receiverSession.UnreadCount)
	}
}
