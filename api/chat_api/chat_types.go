package chat_api

import "myblogx/common"

type ChatListRequest struct {
	common.PageInfo
	UserID uint `form:"user_id"`
	Type   int8 `form:"type" binding:"required,oneof=1 2"`
}
