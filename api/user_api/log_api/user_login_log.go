package log_api

import (
	"myblogx/common"
	"myblogx/common/res"
	"myblogx/global"
	"myblogx/models"
	"myblogx/models/enum"
	"myblogx/utils/jwts"
	"time"

	"github.com/gin-gonic/gin"
)

type UserLoginListRequest struct {
	common.PageInfo
	UserID  uint   `form:"user_id"`
	IP      string `form:"ip"`
	Addr    string `form:"addr"`
	StartAt string `form:"start_at"` // 起止时间，年月日分秒格式
	EndAt   string `form:"end_at"`
	Type    int8   `form:"type" binding:"required,oneof=1 2"` // 查询类型，1：用户查询，2：管理员查询
	// Device  string `form:"device"` // 设备类型，pc、mobile、tablet
}

type UserLoginListResponse struct {
	models.UserLoginModel
	UserNickname string `json:"user_nickname"`
	UserAvatar   string `json:"user_avatar"`
}

func (LogApi) UserLoginLogList(c *gin.Context) {
	var cr UserLoginListRequest
	if err := c.ShouldBindQuery(&cr); err != nil {
		res.FailWithError(err, c)
		return
	}

	claims, err := jwts.GetClaimsByGin(c)
	if err != nil {
		res.FailWithError(err, c)
		return
	}

	// 权限判断
	if (cr.Type == 1 && claims.UserID != cr.UserID) ||
		(cr.Type == 2 && claims.Role != enum.RoleAdmin) {
		res.FailWithMsg("非管理员用户不能查询其他用户登录日志", c)
		return
	}

	// 管理员操作
	var preloads []string
	if cr.Type == 2 {
		preloads = append(preloads, "UserModel")
	}

	// 条件附加：起止时间
	var query = global.DB.Where("")
	if cr.StartAt != "" {
		startTime, err := time.Parse("2006-01-02 15:04:05", cr.StartAt)
		if err != nil {
			res.FailWithMsg("开始时间格式错误", c)
			return
		}
		query = query.Where("created_at >= ?", startTime)
	}
	if cr.EndAt != "" {
		endTime, err := time.Parse("2006-01-02 15:04:05", cr.EndAt)
		if err != nil {
			res.FailWithMsg("结束时间格式错误", c)
			return
		}
		query = query.Where("created_at <= ?", endTime)
	}

	list, count, _ := common.ListQuery(models.UserLoginModel{
		UserID: cr.UserID,
		IP:     cr.IP,
		Addr:   cr.Addr,
	}, common.Options{
		PageInfo: cr.PageInfo,
		Where:    query,
		Preloads: preloads,
	})

	var respList []UserLoginListResponse
	for _, item := range list {
		respList = append(respList, UserLoginListResponse{
			UserLoginModel: item,
			UserNickname:   item.UserModel.Nickname,
			UserAvatar:     item.UserModel.Avatar,
		})
	}

	res.OkWithList(respList, count, c)
}
