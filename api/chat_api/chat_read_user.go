package chat_api

import (
	"fmt"
	"myblogx/common/res"
	"myblogx/global"
	"myblogx/middleware"
	"myblogx/models"
	"myblogx/models/enum/chat_msg_enum"
	"myblogx/utils/jwts"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ChatMsgReadUserView 批量标记当前用户收到的消息为已读。
// 这里只处理“当前用户是接收方”的消息；自己发送的消息和不存在的消息会被自动忽略。
func (ChatApi) ChatMsgReadUserView(c *gin.Context) {
	cr := middleware.GetBindJson[ChatMsgReadUserRequest](c)
	claims := jwts.MustGetClaimsByGin(c)

	if len(cr.MsgIDList) == 0 {
		res.FailWithMsg("请输入要标记已读的消息 id 列表", c)
		return
	}

	var msgList []models.ChatMsgModel
	if err := global.DB.Select("id", "session_id").
		Find(&msgList, "id IN ? AND receiver_id = ? AND msg_status < ?", cr.MsgIDList, claims.UserID, chat_msg_enum.MsgStatusRead).Error; err != nil {
		res.FailWithError(err, c)
		return
	}
	if len(msgList) == 0 {
		res.FailWithMsg("没有可标记已读的消息", c)
		return
	}

	now := time.Now()
	msgIDList := extractChatMsgIDs(msgList)
	sessionUnreadDelta := buildChatSessionUnreadDelta(msgList)

	err := global.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.ChatMsgModel{}).
			Where("id IN ?", msgIDList).
			Updates(map[string]any{
				"msg_status": chat_msg_enum.MsgStatusRead,
				"read_at":    &now,
			}).Error; err != nil {
			return err
		}
		return decreaseChatSessionUnreadCount(tx, claims.UserID, sessionUnreadDelta)
	})
	if err != nil {
		res.FailWithError(err, c)
		return
	}

	res.OkWithMsg(fmt.Sprintf("批量标记已读%d条消息", len(msgList)), c)
}

// 从消息列表中提取主键 id，供批量更新消息状态使用。
func extractChatMsgIDs(msgList []models.ChatMsgModel) []uint {
	idList := make([]uint, 0, len(msgList))
	for _, item := range msgList {
		idList = append(idList, item.ID)
	}
	return idList
}

// 统计本次每个会话实际减少的未读数量。
func buildChatSessionUnreadDelta(msgList []models.ChatMsgModel) map[string]int {
	deltaMap := make(map[string]int, len(msgList))
	for _, item := range msgList {
		deltaMap[item.SessionID]++
	}
	return deltaMap
}

// 按本次批量已读命中的消息数量递减会话未读数，保证未读数不会减成负数。
func decreaseChatSessionUnreadCount(tx *gorm.DB, userID uint, sessionUnreadDelta map[string]int) error {
	if len(sessionUnreadDelta) == 0 {
		return nil
	}

	caseSQL := "CASE session_id"
	args := make([]any, 0, len(sessionUnreadDelta)*3)
	sessionIDList := make([]string, 0, len(sessionUnreadDelta))
	for sessionID, delta := range sessionUnreadDelta {
		caseSQL += " WHEN ? THEN CASE WHEN unread_count >= ? THEN unread_count - ? ELSE 0 END"
		args = append(args, sessionID, delta, delta)
		sessionIDList = append(sessionIDList, sessionID)
	}
	caseSQL += " ELSE unread_count END"

	return tx.Unscoped().Model(&models.ChatSessionModel{}).
		Where("user_id = ? AND session_id IN ?", userID, sessionIDList).
		Update("unread_count", gorm.Expr(caseSQL, args...)).Error
}
