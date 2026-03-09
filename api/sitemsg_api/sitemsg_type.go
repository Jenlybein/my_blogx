package sitemsg_api

type UserMsgConfResponseAndRequest struct {
	DiggNoticeEnabled        bool `json:"digg_notice_enabled"`
	CommentNoticeEnabled     bool `json:"comment_notice_enabled"`
	FavorNoticeEnabled       bool `json:"favor_notice_enabled"`
	PrivateChatNoticeEnabled bool `json:"private_chat_notice_enabled"`
}
