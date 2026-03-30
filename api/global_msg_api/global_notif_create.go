package global_notif_api

import (
	"fmt"
	"myblogx/common/res"
	"myblogx/global"
	"myblogx/middleware"
	"myblogx/models"
	"myblogx/models/enum/global_notif_enum"
	"myblogx/service/log_service"
	"myblogx/utils/jwts"
	"time"

	"github.com/gin-gonic/gin"
)

func (GlobalNotifApi) GlobalNotifCreateView(c *gin.Context) {
	cr := middleware.GetBindJson[GlobalNotifCreateRequest](c)

	claims := jwts.MustGetClaimsByGin(c)
	if claims.IsAdmin() == false {
		res.FailWithMsg("权限错误", c)
		return
	}

	// 默认过期时间
	now := time.Now()
	if cr.ExpireTime != nil {
		if cr.ExpireTime.Before(now.Add(23 * time.Hour)) {
			res.FailWithMsg("过期时间不能小于一天", c)
			return
		}
	} else {
		// 默认过期时间为一周
		nextWeek := now.Add(7 * 24 * time.Hour)
		cr.ExpireTime = &nextWeek
	}

	// 默认可视规则
	if cr.UserVisibleRule == 0 {
		cr.UserVisibleRule = global_notif_enum.UserVisibleRegisteredUsers
	}

	// 检测是否有重复
	var model models.GlobalNotifModel
	if err := global.DB.Take(&model, "title = ?", cr.Title).Error; err == nil {
		res.FailWithMsg("全局通知标题重复", c)
		return
	}

	// 执行创建
	if err := global.DB.Create(&models.GlobalNotifModel{
		ActionUser:      claims.UserID,
		ExpireTime:      *cr.ExpireTime,
		UserVisibleRule: cr.UserVisibleRule,
		Title:           cr.Title,
		Content:         cr.Content,
		Href:            cr.Href,
		Icon:            cr.Icon,
	}).Error; err != nil {
		res.FailWithMsg(fmt.Sprintf("全局通知创建失败 %v", err), c)
		return
	}
	res.OkWithMsg("创建成功", c)
	log_service.EmitActionAuditFromGin(c, log_service.GinAuditInput{
		ActionName:        "global_notif_create",
		TargetType:        "global_notif",
		Success:           true,
		Message:           "创建全局通知成功",
		RequestBody:       cr,
		UseRawRequestBody: true,
		UseRawRequestHead: true,
	})
}
