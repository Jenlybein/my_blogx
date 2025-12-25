package log_api

import (
	"myblogx/common/res"
	"myblogx/global"
	"myblogx/models"
	"myblogx/models/enum"

	"github.com/gin-gonic/gin"
)

type LogApi struct {
}

type LogListRequest struct {
	Limit       int               `form:"limit"`
	Page        int               `form:"page"`
	Key         string            `form:"key"`
	LogType     enum.LogType      `form:"log_type"`     // 日志类型
	Level       enum.LogLevelType `form:"level"`        // 日志级别
	UserID      uint              `form:"user_id"`      // 用户ID
	IP          string            `form:"ip"`           // 操作IP
	LoginStatus bool              `form:"login_status"` // 登录状态
	ServiceName string            `form:"service_name"` // 服务名称
}

type LogListResponse struct {
	models.LogModel
	UserNickname string `json:"user_nickname"`
	UserAvatar   string `json:"user_avatar"`
}

func (l *LogApi) LogListView(c *gin.Context) {
	// 分页 查询(精确匹配，模糊查询)
	var cr LogListRequest
	err := c.ShouldBindQuery(&cr)
	if err != nil {
		res.FailWithError(err, c)
		return
	}

	var list []models.LogModel // 日志列表
	var count int64            // 日志总数

	if cr.Page > 20 {
		cr.Page = 20
	}
	if cr.Page <= 0 {
		cr.Page = 1
	}
	if cr.Limit <= 0 || cr.Limit > 100 {
		cr.Limit = 10
	}

	offset := (cr.Page - 1) * cr.Limit

	model := models.LogModel{
		LogType:     cr.LogType,
		Level:       cr.Level,
		UserID:      cr.UserID,
		IP:          cr.IP,
		LoginStatus: cr.LoginStatus,
		ServiceName: cr.ServiceName,
	}

	like := global.DB.Debug().Where("title like ?", "%"+cr.Key+"%")

	global.DB.Preload("UserModel").Where(like).Where(model).Offset(offset).Limit(cr.Limit).Find(&list)

	global.DB.Where(like).Where(model).Model(&models.LogModel{}).Count(&count)

	var _list = make([]LogListResponse, 0)
	for _, logModel := range list {
		_list = append(_list, LogListResponse{
			LogModel:     logModel,
			UserNickname: logModel.UserModel.Nickname,
			UserAvatar:   logModel.UserModel.Avatar,
		})
	}

	res.OkWithList(_list, int(count), c)
}
