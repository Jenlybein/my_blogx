package log_api

import (
	"database/sql"

	"myblogx/common"
	"myblogx/common/res"
	"myblogx/middleware"
	"myblogx/models"
	"myblogx/models/ctype"
	"myblogx/service/log_service"

	"github.com/gin-gonic/gin"
)

type LogApi struct{}

type RuntimeLogListRequest struct {
	common.PageInfo
	StartAt string   `form:"start_at"`
	EndAt   string   `form:"end_at"`
	Service string   `form:"service"`
	Level   string   `form:"level"`
	Host    string   `form:"host"`
	Method  string   `form:"method"`
	Path    string   `form:"path"`
	UserID  ctype.ID `form:"user_id"`
}

type LoginLogListRequest struct {
	common.PageInfo
	StartAt   string   `form:"start_at"`
	EndAt     string   `form:"end_at"`
	UserID    ctype.ID `form:"user_id"`
	IP        string   `form:"ip"`
	Username  string   `form:"username"`
	LoginType string   `form:"login_type"`
	EventName string   `form:"event_name"`
	Success   *bool    `form:"success"`
}

type ActionAuditListRequest struct {
	common.PageInfo
	StartAt    string   `form:"start_at"`
	EndAt      string   `form:"end_at"`
	UserID     ctype.ID `form:"user_id"`
	IP         string   `form:"ip"`
	ActionName string   `form:"action_name"`
	TargetType string   `form:"target_type"`
	TargetID   string   `form:"target_id"`
	Success    *bool    `form:"success"`
}

func (l *LogApi) RuntimeLogListView(c *gin.Context) {
	cr := middleware.GetBindQuery[RuntimeLogListRequest](c)
	list, count, err := log_service.ListRuntimeLogs(log_service.RuntimeLogQuery{
		PageInfo: common.PageInfo{
			Limit: cr.Limit,
			Page:  cr.Page,
			Key:   cr.Key,
		},
		LogTimeRange: log_service.LogTimeRange{
			StartAt: cr.StartAt,
			EndAt:   cr.EndAt,
		},
		Service: cr.Service,
		Level:   cr.Level,
		Host:    cr.Host,
		Method:  cr.Method,
		Path:    cr.Path,
		UserID:  cr.UserID,
		Key:     cr.Key,
	})
	if err != nil {
		res.FailWithError(err, c)
		return
	}
	res.OkWithList(list, int(count), c)
}

func (l *LogApi) RuntimeLogDetailView(c *gin.Context) {
	cr := middleware.GetBindUri[models.IDRequest](c)
	item, err := log_service.GetRuntimeLog(uint64(cr.ID))
	if err != nil {
		if err == sql.ErrNoRows {
			res.FailWithMsg("运行日志不存在", c)
			return
		}
		res.FailWithError(err, c)
		return
	}
	res.OkWithData(item, c)
}

func (l *LogApi) LoginLogListView(c *gin.Context) {
	cr := middleware.GetBindQuery[LoginLogListRequest](c)
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
		Username:  cr.Username,
		LoginType: cr.LoginType,
		EventName: cr.EventName,
		Success:   cr.Success,
	})
	if err != nil {
		res.FailWithError(err, c)
		return
	}
	res.OkWithList(list, int(count), c)
}

func (l *LogApi) LoginLogDetailView(c *gin.Context) {
	cr := middleware.GetBindUri[models.IDRequest](c)
	item, err := log_service.GetLoginEvent(uint64(cr.ID))
	if err != nil {
		if err == sql.ErrNoRows {
			res.FailWithMsg("登录事件不存在", c)
			return
		}
		res.FailWithError(err, c)
		return
	}
	res.OkWithData(item, c)
}

func (l *LogApi) ActionAuditListView(c *gin.Context) {
	cr := middleware.GetBindQuery[ActionAuditListRequest](c)
	list, count, err := log_service.ListActionAudits(log_service.ActionAuditQuery{
		PageInfo: common.PageInfo{
			Limit: cr.Limit,
			Page:  cr.Page,
		},
		LogTimeRange: log_service.LogTimeRange{
			StartAt: cr.StartAt,
			EndAt:   cr.EndAt,
		},
		UserID:     cr.UserID,
		IP:         cr.IP,
		ActionName: cr.ActionName,
		TargetType: cr.TargetType,
		TargetID:   cr.TargetID,
		Success:    cr.Success,
	})
	if err != nil {
		res.FailWithError(err, c)
		return
	}
	res.OkWithList(list, int(count), c)
}

func (l *LogApi) ActionAuditDetailView(c *gin.Context) {
	cr := middleware.GetBindUri[models.IDRequest](c)
	item, err := log_service.GetActionAudit(uint64(cr.ID))
	if err != nil {
		if err == sql.ErrNoRows {
			res.FailWithMsg("操作审计日志不存在", c)
			return
		}
		res.FailWithError(err, c)
		return
	}
	res.OkWithData(item, c)
}
