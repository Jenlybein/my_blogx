package site_api

import (
	"myblogx/common/res"
	"myblogx/conf"
	"myblogx/middleware"
	"myblogx/service/log_service"
	"myblogx/service/site_service"

	"github.com/gin-gonic/gin"
)

// 数据结构映射表
var confMap = map[string]any{
	"site": &conf.Site{},
	"ai":   &conf.AI{},
}

// 更新站点运行时配置。
// 这里只允许修改数据库中的运行时配置，不再修改 settings.yaml 和前端 HTML 文件。
func (s SiteApi) SiteUpdateView(c *gin.Context) {
	cr := middleware.GetBindUri[SiteInfoRequest](c)
	var auditInput log_service.GinAuditInput

	targetStruct, ok := confMap[cr.Name]
	if !ok {
		res.FailWithMsg("站点信息不存在", c)
		return
	}
	if err := c.ShouldBindJSON(targetStruct); err != nil {
		res.FailWithError(err, c)
		return
	}

	switch s := targetStruct.(type) {
	case *conf.Site:
		if err := site_service.UpdateRuntimeSite(*s); err != nil {
			res.FailWithError(err, c)
			return
		}
		auditInput = log_service.GinAuditInput{
			ActionName:        "site_runtime_update",
			TargetType:        "runtime_site_config",
			TargetID:          "site",
			Success:           true,
			Message:           "更新站点运行时配置成功",
			RequestBody:       *s,
			UseRawRequestBody: true,
			UseRawRequestHead: true,
		}
	case *conf.AI:
		current := site_service.GetRuntimeAI()
		if s.SecretKey == sensitive_place_holder {
			s.SecretKey = current.SecretKey
		}
		if err := site_service.UpdateRuntimeAI(*s); err != nil {
			res.FailWithError(err, c)
			return
		}
		auditInput = log_service.GinAuditInput{
			ActionName: "site_runtime_update",
			TargetType: "runtime_site_config",
			TargetID:   "ai",
			Success:    true,
			Message:    "更新 AI 运行时配置成功",
			RequestBody: map[string]any{
				"enable":     s.Enable,
				"model":      s.ChatModel,
				"base_url":   s.BaseURL,
				"nickname":   s.Nickname,
				"avatar":     s.Avatar,
				"abstract":   s.Abstract,
				"secret_key": "***",
			},
			UseRawRequestBody: true,
			UseRawRequestHead: true,
		}
	default:
		res.FailWithMsg("站点信息不存在", c)
		return
	}

	res.OkWithMsg("站点配置更新成功", c)
	log_service.EmitActionAuditFromGin(c, auditInput)
}
