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
	"gorm.io/gorm/clause"
)

func (GlobalNotifApi) GlobalNotifAdminRemoveView(c *gin.Context) {
	cr := middleware.GetBindJson[models.IDListRequest](c)

	if len(cr.IDList) == 0 {
		res.FailWithMsg("请输入要删除的公告 id 列表", c)
		return
	}

	var list []models.GlobalNotifModel
	if err := global.DB.Find(&list, "id IN ?", cr.IDList).Error; err != nil {
		res.FailWithError(err, c)
		return
	}

	if len(list) > 0 {
		if err := global.DB.Delete(&list).Error; err != nil {
			res.FailWithError(err, c)
			return
		}
	} else {
		res.FailWithMsg("未找到需要删除的公告", c)
		return
	}

	res.OkWithMsg(fmt.Sprintf("请求删除公告%d个，成功%d条", len(cr.IDList), len(list)), c)
}

func (GlobalNotifApi) GlobalNotifUserRemoveView(c *gin.Context) {
	cr := middleware.GetBindJson[models.IDListRequest](c)
	claims := jwts.MustGetClaimsByGin(c)

	state, err := LoadUserGlobalNotifState(claims.UserID, nil)
	if err != nil {
		res.FailWithMsg("用户不存在", c)
		return
	}

	var notifList []models.GlobalNotifModel
	if err := BuildUserVisibleGlobalNotifListQuery(state).Where("id IN ?", cr.IDList).Find(&notifList).Error; err != nil {
		res.FailWithError(err, c)
		return
	}

	// 用户只能删除自己当前“本来就看得见”的通知。
	// 如果传入的 ID 不可见、已过期或不存在，这里会被自然过滤掉。
	if len(notifList) == 0 {
		res.OkWithMsg(fmt.Sprintf("请求删除公告%d个，成功0条", len(cr.IDList)), c)
		return
	}

	var successCount int
	err = global.DB.Transaction(func(tx *gorm.DB) error {
		now := time.Now()
		for _, notif := range notifList {
			userNotif, ok := state.UserNotifMap[notif.ID]
			if ok {
				// 执行软删除
				if userNotif.DeletedAt.Valid {
					continue
				}
				if err := tx.Model(&userNotif).Update("deleted_at", now).Error; err != nil {
					return err
				}
				successCount++
				continue
			}

			// 如果用户此前从未产生过这条通知的个人态记录，
			// 直接创建一条带 deleted_at 的墓碑记录即可。
			// 这样后续列表查询时，仍然能识别“这条通知用户已经删过”。
			userNotif = models.UserGlobalNotifModel{
				Model: models.Model{
					DeletedAt: gorm.DeletedAt{Time: now, Valid: true},
				},
				MsgID:  notif.ID,
				UserID: claims.UserID,
			}
			if err := tx.Clauses(clause.OnConflict{
				Columns: []clause.Column{
					{Name: "msg_id"},
					{Name: "user_id"},
				},
				DoUpdates: clause.Assignments(map[string]any{
					"deleted_at": now,
					"updated_at": now,
				}),
			}).Create(&userNotif).Error; err != nil {
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

	res.OkWithMsg(fmt.Sprintf("请求删除公告%d个，成功%d条", len(cr.IDList), successCount), c)
}
