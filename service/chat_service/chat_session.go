package chat_service

import (
	"fmt"
	"time"

	"myblogx/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// 为一对聊天用户生成稳定的逻辑会话标识。
func buildSessionID(a, b uint) string {
	if a < b {
		return fmt.Sprintf("chat:%d:%d", a, b)
	}
	return fmt.Sprintf("chat:%d:%d", b, a)
}

// 检查聊天双方的会话记录，不存在则分别创建。
func ensureChatSessions(tx *gorm.DB, req ToChatRequest, sessionID string) error {
	if err := findOrCreateSession(tx, sessionID, req.SenderID, req.ReceiverID); err != nil {
		return err
	}
	return findOrCreateSession(tx, sessionID, req.ReceiverID, req.SenderID)
}

// 使用唯一索引 + upsert，避免“先查后插”在并发下产生重复会话。
func findOrCreateSession(tx *gorm.DB, sessionID string, userID, receiverID uint) error {
	session := models.ChatSessionModel{
		SessionID:  sessionID,
		UserID:     userID,
		ReceiverID: receiverID,
	}

	return tx.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "user_id"},
			{Name: "receiver_id"},
		},
		DoUpdates: clause.Assignments(map[string]any{
			"session_id": sessionID,
		}),
	}).Create(&session).Error
}

// updateLastMsgSession 更新双方会话的最后一条消息。
// 发送方只更新摘要，接收方同时累加未读数。
func updateLastMsgSession(tx *gorm.DB, sessionID string, lastMsgID uint, lastMsgContent string, sendTime time.Time, senderID, receiverID uint) error {
	senderUpdates := map[string]any{
		"last_msg_id":      lastMsgID,
		"last_msg_content": lastMsgContent,
		"last_msg_time":    sendTime,
	}
	senderResult := tx.Model(&models.ChatSessionModel{}).
		Where("session_id = ? and user_id = ? and receiver_id = ?", sessionID, senderID, receiverID).
		Updates(senderUpdates)
	if senderResult.Error != nil {
		return senderResult.Error
	}
	if senderResult.RowsAffected == 0 {
		return fmt.Errorf("会话不存在: session_id=%s user_id=%d receiver_id=%d", sessionID, senderID, receiverID)
	}

	receiverUpdates := map[string]any{
		"last_msg_id":      lastMsgID,
		"last_msg_content": lastMsgContent,
		"last_msg_time":    sendTime,
		"unread_count":     gorm.Expr("unread_count + ?", 1),
	}
	receiverResult := tx.Model(&models.ChatSessionModel{}).
		Where("session_id = ? and user_id = ? and receiver_id = ?", sessionID, receiverID, senderID).
		Updates(receiverUpdates)
	if receiverResult.Error != nil {
		return receiverResult.Error
	}
	if receiverResult.RowsAffected == 0 {
		return fmt.Errorf("会话不存在: session_id=%s user_id=%d receiver_id=%d", sessionID, receiverID, senderID)
	}

	return nil
}
