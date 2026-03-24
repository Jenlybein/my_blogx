package chat_service

import (
	"fmt"
	"time"

	"myblogx/models"
	"myblogx/models/ctype"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// 为一对聊天用户生成稳定的逻辑会话标识。
func buildSessionID(a, b ctype.ID) string {
	if a < b {
		return fmt.Sprintf("chat:%d:%d", a, b)
	}
	return fmt.Sprintf("chat:%d:%d", b, a)
}

func isSelfChat(senderID, receiverID ctype.ID) bool {
	return senderID == receiverID
}

// 检查聊天双方的会话记录，不存在则分别创建。
func ensureChatSessions(tx *gorm.DB, req ToChatRequest, sessionID string) error {
	sessions := []models.ChatSessionModel{
		{
			SessionID:  sessionID,
			UserID:     req.SenderID,
			ReceiverID: req.ReceiverID,
		},
	}
	if !isSelfChat(req.SenderID, req.ReceiverID) {
		sessions = append(sessions, models.ChatSessionModel{
			SessionID:  sessionID,
			UserID:     req.ReceiverID,
			ReceiverID: req.SenderID,
		})
	}

	// 数据库中必须给 user_id + receiver_id 组合创建唯一索引（否则这个冲突判断不生效）
	return tx.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "user_id"},
			{Name: "receiver_id"},
		},
		DoUpdates: clause.Assignments(map[string]any{
			"session_id":   sessionID,
			"deleted_at":   nil,
			"unread_count": gorm.Expr("CASE WHEN deleted_at IS NOT NULL THEN 0 ELSE unread_count END"),
		}),
	}).Create(&sessions).Error
}

// updateLastMsgSession 更新双方会话的最后一条消息。
// 发送方只更新摘要，接收方同时累加未读数。
func updateLastMsgSession(tx *gorm.DB, sessionID string, lastMsgID ctype.ID, lastMsgContent string, sendTime time.Time, senderID, receiverID ctype.ID) error {
	expectedRows := int64(2)
	updates := map[string]any{
		"last_msg_id":      lastMsgID,
		"last_msg_content": lastMsgContent,
		"last_msg_time":    sendTime,
		"unread_count": gorm.Expr(
			"CASE WHEN user_id = ? AND receiver_id = ? THEN unread_count + 1 ELSE unread_count END",
			receiverID,
			senderID,
		),
	}
	if isSelfChat(senderID, receiverID) {
		expectedRows = 1
		updates["unread_count"] = gorm.Expr("unread_count")
	}
	result := tx.Model(&models.ChatSessionModel{}).
		Where(
			`session_id = ? AND (
				(user_id = ? AND receiver_id = ?) OR
				(user_id = ? AND receiver_id = ?)
			)`,
			sessionID,
			senderID, receiverID,
			receiverID, senderID,
		).
		Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != expectedRows {
		return fmt.Errorf("会话更新数量异常: session_id=%s affected=%d", sessionID, result.RowsAffected)
	}

	return nil
}
