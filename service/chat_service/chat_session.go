package chat_service

import (
	"errors"
	"fmt"
	"time"

	"myblogx/models"

	"gorm.io/gorm"
)

// 为一对聊天用户生成稳定的逻辑会话标识。
func buildSessionID(a, b uint) string {
	if a < b {
		return fmt.Sprintf("chat:%d:%d", a, b)
	}
	return fmt.Sprintf("chat:%d:%d", b, a)
}

// 检查聊天会话记录，如果不存在则创建
func ensureChatSessions(tx *gorm.DB, req ToChatRequest, sessionID string) (err error) {
	var sender, receiver models.ChatSessionModel

	// 搜索会话
	err = tx.Take(&sender, "session_id = ? and user_id = ? and receiver_id = ?", sessionID, req.SenderID, req.ReceiverID).Error
	if err != nil {
		// 会话不存在，创建会话
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 创建发送者会话
			sender = models.ChatSessionModel{
				SessionID:  sessionID,
				UserID:     req.SenderID,
				ReceiverID: req.ReceiverID,
			}
			if err = tx.Create(&sender).Error; err != nil {
				return
			}

			// 创建接收者会话
			receiver = models.ChatSessionModel{
				SessionID:  sessionID,
				UserID:     req.ReceiverID,
				ReceiverID: req.SenderID,
			}
			if err = tx.Create(&receiver).Error; err != nil {
				return
			}
		} else {
			return
		}
	}

	return
}

// 更新会话的最后一条消息记录
func updateLastMsgSession(tx *gorm.DB, sessionID string, lastMsgID uint, lastMsgContent string, sendTime time.Time, receiverID uint) error {
	var sessions []models.ChatSessionModel
	if err := tx.Find(&sessions, "session_id = ?", sessionID).Error; err != nil {
		return err
	}

	var msg = map[string]any{
		"last_msg_id":      lastMsgID,
		"last_msg_content": lastMsgContent,
		"last_msg_time":    sendTime,
	}

	if sessions[0].UserID == receiverID {
		if err := tx.Model(&sessions[1]).Updates(msg).Error; err != nil {
			return err
		}
		msg["unread_count"] = gorm.Expr("unread_count + ?", 1)
		if err := tx.Model(&sessions[0]).Updates(msg).Error; err != nil {
			return err
		}
	} else if sessions[1].UserID == receiverID {
		if err := tx.Model(&sessions[0]).Updates(msg).Error; err != nil {
			return err
		}
		msg["unread_count"] = gorm.Expr("unread_count + ?", 1)
		if err := tx.Model(&sessions[1]).Updates(msg).Error; err != nil {
			return err
		}

	}

	return nil
}
