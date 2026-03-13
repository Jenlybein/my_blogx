package chat_api

import (
	"testing"
	"time"

	"myblogx/global"
	"myblogx/models"
	"myblogx/test/testutil"
)

func TestValidateChatSendPermissionStrangerDependsOnReceiverConfig(t *testing.T) {
	users := setupChatWSPermissionEnv(t)

	if err := global.DB.Model(&models.UserConfModel{}).
		Where("user_id = ?", users.receiver.ID).
		Update("stranger_chat_enabled", false).Error; err != nil {
		t.Fatalf("更新陌生人私信配置失败: %v", err)
	}
	if err := global.DB.Preload("UserConfModel").Take(&users.receiver, users.receiver.ID).Error; err != nil {
		t.Fatalf("查询接收人失败: %v", err)
	}

	if err := validateChatSendPermission(users.sender.ID, &users.receiver); err == nil || err.Error() != "对方未开启陌生人私信" {
		t.Fatalf("陌生人应受接收配置限制, got=%v", err)
	}

	if err := global.DB.Model(&models.UserConfModel{}).
		Where("user_id = ?", users.receiver.ID).
		Update("stranger_chat_enabled", true).Error; err != nil {
		t.Fatalf("恢复陌生人私信配置失败: %v", err)
	}
	if err := global.DB.Preload("UserConfModel").Take(&users.receiver, users.receiver.ID).Error; err != nil {
		t.Fatalf("查询接收人失败: %v", err)
	}

	if err := validateChatSendPermission(users.sender.ID, &users.receiver); err != nil {
		t.Fatalf("开启陌生人私信后应允许发送: %v", err)
	}
}

func TestValidateChatSendPermissionLimitedByWeeklyCountUntilReply(t *testing.T) {
	users := setupChatWSPermissionEnv(t)
	createFollowRelation(t, users.sender.ID, users.receiver.ID)

	now := time.Now()
	createChatRecord(t, users.sender.ID, users.receiver.ID, now.Add(-6*24*time.Hour))
	createChatRecord(t, users.sender.ID, users.receiver.ID, now.Add(-5*24*time.Hour))
	createChatRecord(t, users.sender.ID, users.receiver.ID, now.Add(-4*24*time.Hour))
	createChatRecord(t, users.sender.ID, users.receiver.ID, now.Add(-3*24*time.Hour))

	if err := validateChatSendPermission(users.sender.ID, &users.receiver); err == nil || err.Error() != "本周可发送消息次数已达上限，请等待对方回复" {
		t.Fatalf("单向关系超过每周 4 条应受限, got=%v", err)
	}

	createChatRecord(t, users.receiver.ID, users.sender.ID, now.Add(-time.Hour))
	if err := validateChatSendPermission(users.sender.ID, &users.receiver); err != nil {
		t.Fatalf("对方回复后应重新允许发送: %v", err)
	}
}

func TestValidateChatSendPermissionFriendAlwaysAllowed(t *testing.T) {
	users := setupChatWSPermissionEnv(t)
	createFollowRelation(t, users.sender.ID, users.receiver.ID)
	createFollowRelation(t, users.receiver.ID, users.sender.ID)

	now := time.Now()
	for i := 0; i < 6; i++ {
		createChatRecord(t, users.sender.ID, users.receiver.ID, now.Add(time.Duration(-i)*time.Hour))
	}

	if err := validateChatSendPermission(users.sender.ID, &users.receiver); err != nil {
		t.Fatalf("好友之间应允许发送: %v", err)
	}
}

type chatWSPermissionUsers struct {
	sender   models.UserModel
	receiver models.UserModel
}

func setupChatWSPermissionEnv(t *testing.T) chatWSPermissionUsers {
	t.Helper()

	testutil.SetupSQLite(t,
		&models.UserModel{},
		&models.UserConfModel{},
		&models.UserFollowModel{},
		&models.ChatMsgModel{},
	)

	sender := createChatWSPermissionUser(t, "ws_sender")
	receiver := createChatWSPermissionUser(t, "ws_receiver")
	return chatWSPermissionUsers{
		sender:   sender,
		receiver: receiver,
	}
}

func createChatWSPermissionUser(t *testing.T, username string) models.UserModel {
	t.Helper()

	user := models.UserModel{
		Username: username,
		Nickname: username + "_nick",
	}
	if err := global.DB.Create(&user).Error; err != nil {
		t.Fatalf("创建用户失败: %v", err)
	}
	if err := global.DB.Preload("UserConfModel").Take(&user, user.ID).Error; err != nil {
		t.Fatalf("查询用户配置失败: %v", err)
	}
	return user
}

func createFollowRelation(t *testing.T, fansUserID, followedUserID uint) {
	t.Helper()

	row := models.UserFollowModel{
		FansUserID:     fansUserID,
		FollowedUserID: followedUserID,
	}
	if err := global.DB.Create(&row).Error; err != nil {
		t.Fatalf("创建关注关系失败: %v", err)
	}
}

func createChatRecord(t *testing.T, senderID, receiverID uint, sendTime time.Time) {
	t.Helper()

	row := models.ChatMsgModel{
		SenderID:   senderID,
		ReceiverID: receiverID,
		Content:    "hello",
		SessionID:  "chat:test",
		SendTime:   sendTime,
	}
	if err := global.DB.Create(&row).Error; err != nil {
		t.Fatalf("创建聊天记录失败: %v", err)
	}
}
