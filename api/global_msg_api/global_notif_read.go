package global_notif_api

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
)

func (GlobalNotifApi) GlobalNotifReadView(c *gin.Context) {
	cr := middleware.GetBindJson[models.IDListRequest](c)
	claims := jwts.MustGetClaimsByGin(c)

	if len(cr.IDList) == 0 {
		res.FailWithMsg("请输入要读取的消息 id 列表", c)
		return
	}

	var user models.UserModel
	if err := global.DB.Take(&user, claims.UserID).Error; err != nil {
		res.FailWithMsg("用户不存在", c)
		return
	}

	var notifList []models.GlobalNotifModel
	if err := global.DB.Where("id IN ?", cr.IDList).Where(buildUserVisibleGlobalNotifQuery(user)).Find(&notifList).Error; err != nil {
		res.FailWithError(err, c)
		return
	}
	if len(notifList) == 0 {
		res.FailWithMsg("消息不存在", c)
		return
	}

	msgIDList := make([]uint, 0, len(notifList))
	for _, item := range notifList {
		msgIDList = append(msgIDList, item.ID)
	}

	var userNotifList []models.UserGlobalNotifModel
	if err := global.DB.Unscoped().Find(&userNotifList, "user_id = ? and msg_id IN ?", claims.UserID, msgIDList).Error; err != nil && err != gorm.ErrRecordNotFound {
		res.FailWithError(err, c)
		return
	}

	userNotifMap := make(map[uint]models.UserGlobalNotifModel, len(userNotifList))
	for _, item := range userNotifList {
		userNotifMap[item.MsgID] = item
	}

	var successCount int
	err := global.DB.Transaction(func(tx *gorm.DB) error {
		now := time.Now()
		for _, notif := range notifList {
			userNotif, ok := userNotifMap[notif.ID]
			if ok {
				if userNotif.DeletedAt != nil || userNotif.IsRead {
					continue
				}
				if err := tx.Model(&userNotif).Updates(map[string]any{
					"is_read": true,
					"read_at": &now,
				}).Error; err != nil {
					return err
				}
				successCount++
				continue
			}

			userNotif = models.UserGlobalNotifModel{
				MsgID:  notif.ID,
				UserID: claims.UserID,
			}
			if err := tx.Create(&userNotif).Error; err != nil {
				return err
			}
			if err := tx.Model(&userNotif).Updates(map[string]any{
				"is_read": true,
				"read_at": &now,
			}).Error; err != nil {
				return err
			}
			successCount++
		}
		return nil
	})
	if err != nil {
		res.FailWithError(err, c)
		return
	}

	if successCount == 0 {
		res.FailWithMsg("没有可标记已读的消息", c)
		return
	}

	res.OkWithMsg(fmt.Sprintf("标记已读%d条消息", successCount), c)
}
