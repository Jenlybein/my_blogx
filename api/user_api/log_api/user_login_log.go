package log_api

import (
	"myblogx/common"
	"myblogx/common/res"
	"myblogx/middleware"
	"myblogx/models/ctype"
	"myblogx/models/enum"
	"myblogx/service/log_service"
	"myblogx/utils/jwts"

	"github.com/gin-gonic/gin"
)

type UserLoginListRequest struct {
	common.PageInfo
	UserID  ctype.ID `form:"user_id"`
	IP      string   `form:"ip"`
	StartAt string   `form:"start_at"`
	EndAt   string   `form:"end_at"`
	Type    int8     `form:"type" binding:"required,oneof=1 2"` // 1：用户查自己 2：管理员查任意用户
}

func (LogApi) UserLoginLogList(c *gin.Context) {
	cr := middleware.GetBindQuery[UserLoginListRequest](c)
	claims := jwts.MustGetClaimsByGin(c)

	if cr.Type == 1 {
		if cr.UserID == 0 {
			cr.UserID = claims.UserID
		}
		if claims.UserID != cr.UserID {
			res.FailWithMsg("非管理员用户不能查询其他用户登录日志", c)
			return
		}
	}
	if cr.Type == 2 && claims.Role != enum.RoleAdmin {
		res.FailWithMsg("非管理员用户不能查询其他用户登录日志", c)
		return
	}

	eventName := "login_success"
	success := true
	list, count, err := log_service.ListLoginEvents(log_service.LoginEventQuery{
		PageInfo: common.PageInfo{
			Limit: cr.Limit,
			Page:  cr.Page,
		},
		LogTimeRange: log_service.LogTimeRange{
			StartAt: cr.StartAt,
			EndAt:   cr.EndAt,
		},
		UserID:    cr.UserID,
		IP:        cr.IP,
		EventName: eventName,
		Success:   &success,
	})
	if err != nil {
		res.FailWithError(err, c)
		return
	}

	res.OkWithList(list, int(count), c)
}
