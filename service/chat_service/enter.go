package chat_service

import (
	"time"

	"myblogx/global"
	"myblogx/models"
	"myblogx/models/ctype"
	"myblogx/models/enum/chat_msg_enum"

	"gorm.io/gorm"
)

type ChatService struct {
}

// ToChat 创建一条聊天消息
type ToChatRequest struct {
	SenderID   ctype.ID
	ReceiverID ctype.ID
	MsgType    chat_msg_enum.MsgType
	Content    string
	SendTime   time.Time
	MsgStatus  chat_msg_enum.MsgStatus
}

func ToChat(req ToChatRequest) (*models.ChatMsgModel, error) {
	// 基础校验
	if err := validateChatBase(&req); err != nil {
		return nil, err
	}

	// 构建会话标识
	sessionID := buildSessionID(req.SenderID, req.ReceiverID)

	var msg *models.ChatMsgModel
	err := global.DB.Transaction(func(tx *gorm.DB) error {
		// 会话查找或创建
		if err := ensureChatSessions(tx, req, sessionID); err != nil {
			return err
		}

		// 消息落库
		msg = &models.ChatMsgModel{
			SenderID:   req.SenderID,
			ReceiverID: req.ReceiverID,
			SessionID:  sessionID,
			Content:    req.Content,
			SendTime:   req.SendTime,
			MsgStatus:  req.MsgStatus,
			MsgType:    req.MsgType,
		}

		if err := tx.Create(msg).Error; err != nil {
			global.Logger.Errorf("创建聊天消息失败: %v", err)
			return err
		}

		// 生成最后一条消息的摘要
		lastMsgContent := buildSessionLastMsg(req.MsgType, req.Content)

		// 同步更新双方会话的最后一条消息
		if err := updateLastMsgSession(tx, sessionID, msg.ID, lastMsgContent, req.SendTime, req.SenderID, req.ReceiverID); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}
	return msg, nil
}
