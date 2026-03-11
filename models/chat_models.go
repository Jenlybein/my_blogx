package models

import (
	"myblogx/models/enum/chat_msg_enum"
	"time"
)

// 存储时，为双方各存一份会话记录
// 用 UserID 存，做到 “单表快速查询”，避免复杂的联表 / 条件判断。
// 设计成只存一份的话，查的字段会多一个，性能差

// id 是数据库内部标识
// session_id 是语义化标识
// 跨表关联时（比如chat_message关联chat_session），如果只用id，一旦会话表数据迁移（如分库分表），id可能重复，导致关联错误；

// 聊天会话
type ChatSessionModel struct {
	Model
	SessionID      string    `json:"session_id"`
	UserID         uint      `json:"user_id"`
	ReceiverID     uint      `json:"receiver_id"`
	LastMsgID      uint      `json:"last_msg_id"`
	LastMsgContent string    `json:"last_msg_content"`
	LastMsgTime    time.Time `json:"last_msg_time"`
	UnreadCount    int       `json:"unread_count"`
	UserModel      UserModel `gorm:"foreignKey:UserID;references:ID" json:"-"`
	ReceiverModel  UserModel `gorm:"foreignKey:ReceiverID;references:ID" json:"-"`
	IsTop          bool      `json:"is_top"`  // 是否置顶
	IsMute         bool      `json:"is_mute"` // 是否静音
}

// 聊天消息
type ChatMsgModel struct {
	Model
	SenderID     uint                    `json:"sender_id"`
	ReceiverID   uint                    `json:"receiver_id"`
	SessionID    string                  `json:"session_id"`
	Content      string                  `json:"content"`
	SendTime     time.Time               `json:"send_time"`
	ReadAt       *time.Time              `json:"read_at"`
	MsgStatus    chat_msg_enum.MsgStatus `json:"msg_status"`
	MsgType      chat_msg_enum.MsgType   `json:"msg_type"`
	SessionModel ChatSessionModel        `gorm:"foreignKey:SessionID;references:ID" json:"-"`
}
