package models

import (
	"myblogx/models/enum/message_enum"
	"time"
)

type ArticleMessageModel struct {
	Model
	Type message_enum.Type `json:"type"`

	ReceiverID         uint    `gorm:"index:idx_msg_receiver_read_created,priority:1" json:"receiver_id"`
	ActionUserID       *uint   `gorm:"index" json:"action_user_id"`
	ActionUserNickname *string `json:"action_user_nickname"`
	ActionUserAvatar   *string `json:"action_user_avatar"`

	Content string `gorm:"type:text" json:"content"`

	// 记录触发的业务对象
	ArticleID    uint   `json:"article_id"`
	CommentID    uint   `json:"comment_id"`
	ArticleTitle string `json:"article_title"`

	// 额外提示的链接
	LinkTitle string `gorm:"size:128" json:"link_title"`
	LinkHerf  string `gorm:"size:256" json:"link_herf"`

	// 是否已读
	IsRead bool       `gorm:"default:false;index:idx_msg_receiver_read_created,priority:2" json:"is_read"`
	ReadAt *time.Time `json:"read_at"`
}
