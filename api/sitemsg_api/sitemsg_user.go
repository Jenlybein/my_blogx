package sitemsg_api

import (
	global_notif_api "myblogx/api/global_msg_api"
	"myblogx/common/res"
	"myblogx/global"
	"myblogx/models"
	"myblogx/models/enum/message_enum"
	"myblogx/utils/jwts"

	"github.com/gin-gonic/gin"
)

func (SitemsgApi) SitemsgUserView(c *gin.Context) {
	claims := jwts.MustGetClaimsByGin(c)

	var msgList []models.ArticleMessageModel
	if err := global.DB.Find(&msgList, "receiver_id = ? and is_read = ?", claims.UserID, false).Error; err != nil {
		res.FailWithError(err, c)
		return
	}

	var data SitemsgUserResponse
	for _, v := range msgList {
		switch v.Type {
		case message_enum.CommentArticleType, message_enum.CommentReplyType:
			data.CommentMsgCount++
		case message_enum.DiggArticleType, message_enum.DiggCommentType, message_enum.FavorArticleType:
			data.DiggFavorMsgCount++
		case message_enum.SystemType:
			data.SystemMsgCount++
		}
	}

	// 计算未读的私信总数
	if err := global.DB.Model(&models.ChatSessionModel{}).
		Where("user_id = ? AND unread_count > 0", claims.UserID).
		Select("COALESCE(SUM(unread_count), 0)").
		Scan(&data.PrivateMsgCount).Error; err != nil {
		res.FailWithError(err, c)
		return
	}

	// 算未读的全局消息
	state, err := global_notif_api.LoadUserGlobalNotifState(claims.UserID, nil)
	if err != nil {
		res.FailWithMsg("用户不存在", c)
		return
	}

	var globalMsg []models.GlobalNotifModel
	if err := global_notif_api.BuildUserVisibleGlobalNotifListQuery(state).Find(&globalMsg).Error; err != nil {
		res.FailWithError(err, c)
		return
	}

	for _, item := range globalMsg {
		userNotif, ok := state.UserNotifMap[item.ID]
		if !ok || !userNotif.IsRead {
			data.SystemMsgCount++
		}
	}

	res.OkWithData(data, c)
}
