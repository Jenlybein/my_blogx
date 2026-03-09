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

func buildUserVisibleGlobalNotifQuery(user models.UserModel) *gorm.DB {
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
