package site_api

import (
	"myblogx/common/res"
	"myblogx/global"
	"myblogx/middleware"

	"github.com/gin-gonic/gin"
)

// 获取站点基本配置信息-任何用户
func (s SiteApi) SiteInfoView(c *gin.Context) {
	cr := middleware.GetBindUri[SiteInfoRequest](c)

	var data any

	switch cr.Name {
	// 站点版本
	case "site":
		rep := global.Config.Site
		rep.About.Version = global.Version
		data = rep
	}

	res.OkWithData(data, c)
}

// 获取站点配置信息-管理员
func (s SiteApi) SiteInfoAdminView(c *gin.Context) {
	cr := middleware.GetBindUri[SiteInfoRequest](c)

	var data any

	switch cr.Name {
	case "email":
		rep := global.Config.Email
		rep.AuthCode = sensitive_place_holder
		data = rep
	case "qq":
		rep := global.Config.QQ
		rep.AppKey = sensitive_place_holder
		data = rep
	case "qiniu":
		rep := global.Config.QiNiu
		rep.SecretKey = sensitive_place_holder
		data = rep
	case "ai":
		rep := global.Config.AI
		rep.SecretKey = sensitive_place_holder
		data = rep
	default:
		res.FailWithMsg("站点信息不存在", c)
		return
	}

	res.OkWithData(data, c)
}
