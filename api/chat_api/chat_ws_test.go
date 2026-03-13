package chat_api

import (
	"testing"

	"myblogx/global"
	"myblogx/models"
	"myblogx/service/chat_service"
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

	reservation, err := validateChatSendPermission(users.sender.ID, &users.receiver)
	if err == nil || err.Error() != "对方未开启陌生人私信" {
		t.Fatalf("陌生人应受接收配置限制, got=%v", err)
	}
	if reservation != nil {
		t.Fatal("受限时不应返回预占对象")
	}

	if err := global.DB.Model(&models.UserConfModel{}).
		Where("user_id = ?", users.receiver.ID).
		Update("stranger_chat_enabled", true).Error; err != nil {
		t.Fatalf("恢复陌生人私信配置失败: %v", err)
	}
	if err := global.DB.Preload("UserConfModel").Take(&users.receiver, users.receiver.ID).Error; err != nil {
		t.Fatalf("查询接收人失败: %v", err)
	}

	reservation, err = validateChatSendPermission(users.sender.ID, &users.receiver)
	if err != nil {
		t.Fatalf("开启陌生人私信后应允许发送: %v", err)
	}
	if reservation == nil {
		t.Fatal("允许发送时应返回预占对象")
	}
	if err := reservation.Rollback(); err != nil {
		t.Fatalf("回滚预占失败: %v", err)
	}
}

func TestValidateChatSendPermissionLimitedByWeeklyQuotaUntilReply(t *testing.T) {
	users := setupChatWSPermissionEnv(t)
	createFollowRelation(t, users.sender.ID, users.receiver.ID)

	for i := 0; i < 3; i++ {
		reservation, err := validateChatSendPermission(users.sender.ID, &users.receiver)
		if err != nil {
			t.Fatalf("第 %d 次发送前置校验失败: %v", i+1, err)
		}
		if err := reservation.Commit(); err != nil {
			t.Fatalf("提交预占失败: %v", err)
		}
	}

	reservation, err := validateChatSendPermission(users.sender.ID, &users.receiver)
	if err == nil || err.Error() != "本周可发送消息次数已达上限，请等待对方回复" {
		t.Fatalf("单向关系超过自然周 3 条应受限, got=%v", err)
	}
	if reservation != nil {
		t.Fatal("受限时不应返回预占对象")
	}

	replyReservation, err := validateChatSendPermission(users.receiver.ID, &users.sender)
	if err != nil {
		t.Fatalf("对方回复前置校验失败: %v", err)
	}
	if err := replyReservation.Commit(); err != nil {
		t.Fatalf("提交回复预占失败: %v", err)
	}

	reservation, err = validateChatSendPermission(users.sender.ID, &users.receiver)
	if err != nil {
		t.Fatalf("对方回复后应重新允许发送: %v", err)
	}
	if reservation == nil {
		t.Fatal("对方回复后应返回预占对象")
	}
	if err := reservation.Rollback(); err != nil {
		t.Fatalf("回滚预占失败: %v", err)
	}
}

func TestValidateChatSendPermissionFriendAlwaysAllowed(t *testing.T) {
	users := setupChatWSPermissionEnv(t)
	createFollowRelation(t, users.sender.ID, users.receiver.ID)
	createFollowRelation(t, users.receiver.ID, users.sender.ID)

	for i := 0; i < 5; i++ {
		reservation, err := validateChatSendPermission(users.sender.ID, &users.receiver)
		if err != nil {
			t.Fatalf("好友之间第 %d 次发送应允许: %v", i+1, err)
		}
		if err := reservation.Commit(); err != nil {
			t.Fatalf("提交预占失败: %v", err)
		}
	}
}

func TestValidateChatSendPermissionSessionMinuteLimit(t *testing.T) {
	users := setupChatWSPermissionEnv(t)

	for i := 0; i < 30; i++ {
		reservation, err := validateChatSendPermission(users.sender.ID, &users.receiver)
		if err != nil {
			t.Fatalf("第 %d 次会话发送应允许: %v", i+1, err)
		}
		if err := reservation.Commit(); err != nil {
			t.Fatalf("提交预占失败: %v", err)
		}
	}

	reservation, err := validateChatSendPermission(users.sender.ID, &users.receiver)
	if err == nil || err.Error() != "当前会话发送过于频繁，请稍后再试" {
		t.Fatalf("同一会话 60 秒内第 31 条应受限, got=%v", err)
	}
	if reservation != nil {
		t.Fatal("受限时不应返回预占对象")
	}
}

type chatWSPermissionUsers struct {
	sender   models.UserModel
	receiver models.UserModel
}

func setupChatWSPermissionEnv(t *testing.T) chatWSPermissionUsers {
	t.Helper()

	_ = testutil.SetupMiniRedis(t)
	testutil.SetupSQLite(t,
		&models.UserModel{},
		&models.UserConfModel{},
		&models.UserFollowModel{},
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

func validateChatSendPermission(senderID uint, receiver *models.UserModel) (*chat_service.ChatSendReservation, error) {
	return chat_service.CheckAndReserveChatSend(senderID, receiver)
}
