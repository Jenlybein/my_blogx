package chat_api

import (
	"myblogx/common"
	"myblogx/models/enum/chat_msg_enum"
	"time"
)

type ChatMsgListRequest struct {
	common.PageInfo
	SessionID string `form:"session_id" binding:"required"`
}

type ChatMsgListResponse struct {
	ID         uint                    `json:"id"`
	SenderID   uint                    `json:"sender_id"`
	ReceiverID uint                    `json:"receiver_id"`
	SessionID  string                  `json:"session_id"`
	Content    string                  `json:"content"`
	SendTime   time.Time               `json:"send_time"`
	MsgStatus  chat_msg_enum.MsgStatus `json:"msg_status"`
	MsgType    chat_msg_enum.MsgType   `json:"msg_type"`
	IsSelf     bool                    `json:"is_self"`
	IsRead     bool                    `json:"is_read"`
}

type ChatSessionListRequest struct {
	common.PageInfo
}

type ChatSessionListResponse struct {
	SessionID        string    `json:"session_id"`
	ReceiverID       uint      `json:"receiver_id"`
	ReceiverNickname string    `json:"receiver_nickname"`
	ReceiverAvatar   string    `json:"receiver_avatar"`
	LastMsgContent   string    `json:"last_msg_content"`
	LastMsgTime      time.Time `json:"last_msg_time"`
	UnreadCount      int       `json:"unread_count"`
	IsTop            bool      `json:"is_top"`
	IsMute           bool      `json:"is_mute"`
}
