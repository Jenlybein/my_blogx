package global_notif_api

import (
	"myblogx/global"
	"myblogx/models"
	"myblogx/models/enum/global_notif_enum"
	"time"

	"gorm.io/gorm"
)

type GlobalNotifApi struct {
}

type UserGlobalNotifState struct {
	User             models.UserModel
	UserNotifMap     map[uint]models.UserGlobalNotifModel
	DeletedMsgIDList []uint
}

func BuildUserVisibleGlobalNotifQuery(user models.UserModel) *gorm.DB {
	now := time.Now()
	return global.DB.
		Where("expire_time > ?", now).
		Where(global.DB.Where("user_visible_rule = ?", global_notif_enum.UserVisibleAllUsers).
			Or(global.DB.Where(
				"user_visible_rule = ? AND created_at >= ? AND expire_time >= ?",
				global_notif_enum.UserVisibleRegisteredUsers,
				user.CreatedAt,
				user.CreatedAt,
			)).
			Or(global.DB.Where(
				"user_visible_rule = ? AND created_at < ? AND expire_time >= ?",
				global_notif_enum.UserVisibleNewUsers,
				user.CreatedAt,
				user.CreatedAt,
			)))
}

func LoadUserGlobalNotifState(userID uint, msgIDList []uint) (state UserGlobalNotifState, err error) {
	if err = global.DB.Take(&state.User, userID).Error; err != nil {
		return state, err
	}

	query := global.DB.Unscoped().Where("user_id = ?", userID)
	if len(msgIDList) > 0 {
		query = query.Where("msg_id IN ?", msgIDList)
	}

	var userNotifList []models.UserGlobalNotifModel
	if err = query.Find(&userNotifList).Error; err != nil {
		return state, err
	}

	state.UserNotifMap = make(map[uint]models.UserGlobalNotifModel, len(userNotifList))
	state.DeletedMsgIDList = make([]uint, 0)
	for _, item := range userNotifList {
		state.UserNotifMap[item.MsgID] = item
		if item.DeletedAt != nil {
			state.DeletedMsgIDList = append(state.DeletedMsgIDList, item.MsgID)
		}
	}
	return state, nil
}

func BuildUserVisibleGlobalNotifListQuery(state UserGlobalNotifState) *gorm.DB {
	query := global.DB.Where(BuildUserVisibleGlobalNotifQuery(state.User))
	if len(state.DeletedMsgIDList) > 0 {
		query = query.Where("id NOT IN ?", state.DeletedMsgIDList)
	}
	return query
}
