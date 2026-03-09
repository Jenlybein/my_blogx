package sitemsg_api

import "myblogx/common"

type UserMsgConfResponseAndRequest struct {
	DiggNoticeEnabled        bool `json:"digg_notice_enabled"`
	CommentNoticeEnabled     bool `json:"comment_notice_enabled"`
	FavorNoticeEnabled       bool `json:"favor_notice_enabled"`
	PrivateChatNoticeEnabled bool `json:"private_chat_notice_enabled"`
}

type SitemsgListRequest struct {
	common.PageInfo
	T int8 `form:"t" binding:"required,oneof=1 2 3"` // 1.评论和回复 2.点赞和收藏 3.系统通知
}

type SitemsgReadRequest struct {
	ID uint `json:"id"`
	T  int8 `json:"t" binding:"omitempty,oneof=1 2 3"` // 批量已读的类型
}
