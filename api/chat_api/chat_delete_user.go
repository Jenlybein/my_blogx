package chat_api

import (
	"fmt"
	"myblogx/common/res"
	"myblogx/global"
	"myblogx/middleware"
	"myblogx/models"
	"myblogx/utils/jwts"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// 低风险：普通用户消息列表的已删过滤还是 NOT IN (subquery)。
// chat_delete_user.go 在依赖状态表子查询排除消息。配合当前索引能跑，但如果以后状态表很大，NOT EXISTS 或显式 LEFT JOIN ... IS NULL 往往更容易拿到稳定执行计划。

func (ChatApi) ChatSessionDeleteUserView(c *gin.Context) {
	cr := middleware.GetBindJson[ChatSessionDeleteUserRequest](c)
	claims := jwts.MustGetClaimsByGin(c)

	if len(cr.SessionIDList) == 0 {
		res.FailWithMsg("请输入要删除的会话 session_id 列表", c)
		return
	}

	var list []models.ChatSessionModel
	if err := global.DB.Find(&list, "user_id = ? and session_id IN ?", claims.UserID, cr.SessionIDList).Error; err != nil {
		res.FailWithError(err, c)
		return
	}

	if len(list) > 0 {
		err := global.DB.Transaction(func(tx *gorm.DB) error {
			if err := clearChatSessions(tx, list); err != nil {
				return err
			}
			return tx.Delete(&list).Error
		})
		if err != nil {
			res.FailWithError(err, c)
			return
		}
	}

	res.OkWithMsg(fmt.Sprintf("请求删除会话%d个，成功%d条", len(cr.SessionIDList), len(list)), c)
}

func (ChatApi) ChatMsgDeleteUserView(c *gin.Context) {
	cr := middleware.GetBindJson[ChatMsgDeleteUserRequest](c)
	claims := jwts.MustGetClaimsByGin(c)

	if len(cr.MsgIDList) == 0 {
		res.FailWithMsg("请输入要删除的消息 id 列表", c)
		return
	}

	var msgList []models.ChatMsgModel
	if err := global.DB.Select("id", "session_id").
		Find(&msgList, "id IN ? AND (sender_id = ? OR receiver_id = ?)", cr.MsgIDList, claims.UserID, claims.UserID).Error; err != nil {
		res.FailWithError(err, c)
		return
	}

	if len(msgList) > 0 {
		if err := insertChatMsgUserStates(global.DB, claims.UserID, msgList); err != nil {
			res.FailWithError(err, c)
			return
		}
	}

	res.OkWithMsg(fmt.Sprintf("请求删除消息%d个，成功%d条", len(cr.MsgIDList), len(msgList)), c)
}

func buildChatMsgVisibleWhere(userID uint, sessionID string, clearBeforeMsgID uint, allowUnscoped bool) *gorm.DB {
	query := global.DB
	if clearBeforeMsgID > 0 {
		query = query.Where("id > ?", clearBeforeMsgID)
	}
	if allowUnscoped {
		return query
	}
	subQuery := global.DB.Unscoped().Model(&models.ChatMsgUserStateModel{}).
		Select("msg_id").
		Where("user_id = ? AND session_id = ? AND deleted_at IS NOT NULL", userID, sessionID)
	return query.Not("id IN (?)", subQuery)
}

func clearChatSessions(tx *gorm.DB, list []models.ChatSessionModel) error {
	if len(list) == 0 {
		return nil
	}

	sessionIDList := extractSessionIDs(list)
	maxMsgIDMap, err := loadSessionMaxMsgIDMap(tx, sessionIDList)
	if err != nil {
		return err
	}

	for _, session := range list {
		clearBeforeMsgID := session.ClearBeforeMsgID
		if maxMsgIDMap[session.SessionID] > clearBeforeMsgID {
			clearBeforeMsgID = maxMsgIDMap[session.SessionID]
		}

		if err := tx.Model(&models.ChatSessionModel{}).
			Where("id = ?", session.ID).
			Updates(map[string]any{
				"clear_before_msg_id": clearBeforeMsgID,
				"unread_count":        0,
			}).Error; err != nil {
			return err
		}
	}
	return nil
}

func insertChatMsgUserStates(tx *gorm.DB, userID uint, msgList []models.ChatMsgModel) error {
	if len(msgList) == 0 {
		return nil
	}

	now := time.Now()
	stateList := make([]models.ChatMsgUserStateModel, 0, len(msgList))
	for _, msg := range msgList {
		stateList = append(stateList, models.ChatMsgUserStateModel{
			Model: models.Model{
				DeletedAt: gorm.DeletedAt{Time: now, Valid: true},
			},
			MsgID:     msg.ID,
			UserID:    userID,
			SessionID: msg.SessionID,
		})
	}

	return tx.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "msg_id"},
			{Name: "user_id"},
		},
		DoUpdates: clause.AssignmentColumns([]string{"session_id", "deleted_at", "updated_at"}),
	}).Create(&stateList).Error
}

func extractSessionIDs(list []models.ChatSessionModel) []string {
	sessionIDList := make([]string, 0, len(list))
	for _, item := range list {
		sessionIDList = append(sessionIDList, item.SessionID)
	}
	return sessionIDList
}

type chatSessionMaxMsgRow struct {
	SessionID string
	MaxMsgID  uint
}

func loadSessionMaxMsgIDMap(tx *gorm.DB, sessionIDList []string) (map[string]uint, error) {
	if len(sessionIDList) == 0 {
		return map[string]uint{}, nil
	}

	var rows []chatSessionMaxMsgRow
	err := tx.Model(&models.ChatMsgModel{}).
		Select("session_id, MAX(id) AS max_msg_id").
		Where("session_id IN ?", sessionIDList).
		Group("session_id").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	result := make(map[string]uint, len(rows))
	for _, row := range rows {
		result[row.SessionID] = row.MaxMsgID
	}
	return result, nil
}
