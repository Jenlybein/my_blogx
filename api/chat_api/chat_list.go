package chat_api

import (
	"myblogx/common"
	"myblogx/common/res"
	"myblogx/global"
	"myblogx/middleware"
	"myblogx/models"
	"myblogx/models/enum/chat_msg_enum"
	"myblogx/utils/jwts"

	"github.com/gin-gonic/gin"
)

// ChatSessionListView 返回当前登录用户的会话列表。
func (a *ChatApi) ChatSessionListView(c *gin.Context) {
	cr := middleware.GetBindQuery[ChatSessionListRequest](c)
	claims := jwts.MustGetClaimsByGin(c)

	var opts = common.Options{
		PageInfo:      cr.PageInfo,
		ExactPreloads: map[string][]string{"ReceiverModel": {"id", "nickname", "avatar"}},
		DefaultOrder:  "is_top desc, last_msg_time desc, id desc",
	}

	switch cr.Type {
	case 1:
		cr.UserID = claims.UserID
	case 2:
		if !claims.IsAdmin() {
			res.FailWithMsg("权限不足", c)
			return
		}
		if cr.UserID == 0 {
			res.FailWithMsg("user_id 不能为 0", c)
			return
		}
		opts.Unscoped = true
	}

	list, count, err := common.ListQuery(models.ChatSessionModel{
		UserID: cr.UserID,
	}, opts)
	if err != nil {
		res.FailWithError(err, c)
		return
	}

	respList := make([]ChatSessionListResponse, 0, len(list))
	for _, item := range list {
		data := ChatSessionListResponse{
			SessionID:        item.SessionID,
			ReceiverID:       item.ReceiverID,
			ReceiverNickname: item.ReceiverModel.Nickname,
			ReceiverAvatar:   item.ReceiverModel.Avatar,
			LastMsgContent:   item.LastMsgContent,
			LastMsgTime:      item.LastMsgTime,
			UnreadCount:      item.UnreadCount,
			IsTop:            item.IsTop,
			IsMute:           item.IsMute,
		}

		if cr.Type == 2 && item.DeletedAt.Valid {
			data.DeletedAt = &item.DeletedAt.Time
		}

		respList = append(respList, data)
	}

	res.OkWithList(respList, count, c)
}

// ChatMsgListView 返回当前登录用户在某个会话下的消息列表。
func (a *ChatApi) ChatMsgListView(c *gin.Context) {
	cr := middleware.GetBindQuery[ChatMsgListRequest](c)
	claims := jwts.MustGetClaimsByGin(c)

	allowUnscoped := false
	switch cr.Type {
	case 1:
		cr.UserID = claims.UserID
	case 2:
		if !claims.IsAdmin() {
			res.FailWithMsg("权限不足", c)
			return
		}
		if cr.UserID == 0 {
			res.FailWithMsg("user_id 不能为 0", c)
			return
		}
		allowUnscoped = true
	}

	var session models.ChatSessionModel
	sessionQuery := global.DB.Select("session_id")
	if allowUnscoped {
		sessionQuery = sessionQuery.Unscoped()
	}
	if err := sessionQuery.
		Take(&session, "session_id = ? and user_id = ?", cr.SessionID, cr.UserID).Error; err != nil {
		res.FailWithMsg("会话不存在", c)
		return
	}

	list, count, err := common.ListQuery(models.ChatMsgModel{
		SessionID: cr.SessionID,
	}, common.Options{
		PageInfo:     cr.PageInfo,
		DefaultOrder: "send_time desc, id desc",
		Unscoped:     allowUnscoped,
	})
	if err != nil {
		res.FailWithError(err, c)
		return
	}

	respList := make([]ChatMsgListResponse, 0, len(list))
	for _, item := range list {
		data := ChatMsgListResponse{
			ID:         item.ID,
			SenderID:   item.SenderID,
			ReceiverID: item.ReceiverID,
			SessionID:  item.SessionID,
			Content:    item.Content,
			SendTime:   item.SendTime,
			MsgStatus:  item.MsgStatus,
			MsgType:    item.MsgType,
			IsSelf:     item.SenderID == cr.UserID,
			IsRead:     int8(item.MsgStatus) >= int8(chat_msg_enum.MsgStatusRead),
		}
		if allowUnscoped && item.DeletedAt.Valid {
			data.DeletedAt = &item.DeletedAt.Time
		}
		respList = append(respList, data)
	}

	res.OkWithList(respList, count, c)
}
