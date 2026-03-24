package log_api

import (
	"fmt"
	"myblogx/common"
	"myblogx/common/res"
	"myblogx/global"
	"myblogx/middleware"
	"myblogx/models"
	"myblogx/models/ctype"
	"myblogx/models/enum"
	"myblogx/service/log_service"

	"github.com/gin-gonic/gin"
)

type LogApi struct {
}

type LogListRequest struct {
	common.PageInfo
	LogType     enum.LogType      `form:"log_type"`     // 日志类型
	Level       enum.LogLevelType `form:"level"`        // 日志级别
	UserID      ctype.ID          `form:"user_id"`      // 用户ID
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
	cr := middleware.GetBindQuery[LogListRequest](c)

	list, count, err := common.ListQuery(models.LogModel{
		LogType:     cr.LogType,
		Level:       cr.Level,
		UserID:      cr.UserID,
		IP:          cr.IP,
		LoginStatus: cr.LoginStatus,
		ServiceName: cr.ServiceName,
	}, common.Options{
		PageInfo: cr.PageInfo,
		Likes:    []string{"Title"},
		Preloads: []string{"UserModel"},
		Debug:    true,
		// DefaultOrder: "created_at DESC",
	})
	if err != nil {
		res.FailWithError(err, c)
		return
	}

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

func (l *LogApi) LogReadView(c *gin.Context) {
	// 已读状态修改
	cr := middleware.GetBindUri[models.IDRequest](c)

	var log models.LogModel
	err := global.DB.Take(&log, cr.ID).Error
	if err != nil {
		res.FailWithError(err, c)
		return
	}

	// 如果日志已读，则返回错误
	if !log.IsRead {
		global.DB.Model(&log).Update("is_read", true)
	}

	res.OkWithData("日志读取成功", c)
}

func (l *LogApi) LogRemoveView(c *gin.Context) {
	cr := middleware.GetBindJson[models.IDListRequest](c)

	log := log_service.GetLog(c)
	log.SetShowRequest()
	log.SetShowResponse()

	var logList []models.LogModel
	global.DB.Find(&logList, "id IN ?", cr.IDList)

	if len(logList) > 0 {
		global.DB.Delete(&logList)
	}

	msg := fmt.Sprintf("删除了 %d 条日志", len(logList))
	res.OkWithData(msg, c)
}
