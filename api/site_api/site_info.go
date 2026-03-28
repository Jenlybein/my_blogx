package site_api

import (
	"myblogx/common/res"
	"myblogx/global"
	"myblogx/middleware"
	"myblogx/service/redis_service/redis_site"
	"myblogx/service/site_service"

	"github.com/gin-gonic/gin"
)

// 获取站点基本配置信息-任何用户
func (s SiteApi) SiteInfoView(c *gin.Context) {
	cr := middleware.GetBindUri[SiteInfoRequest](c)

	var data any

	switch cr.Name {
	// 站点版本
	case "site":
		redis_site.SetFlow()
		rep := site_service.GetRuntimeSite()
		rep.About.Version = global.Version
		data = rep
	case "seo":
		site := site_service.GetRuntimeSite()
		data = SiteSEOResponse{
			SiteTitle:    site.SiteInfo.Title,
			ProjectTitle: site.Project.Title,
			Logo:         site.SiteInfo.Logo,
			Icon:         site.Project.Icon,
			Keywords:     site.Seo.Keywords,
			Description:  site.Seo.Description,
		}
	default:
		res.FailWithMsg("站点信息不存在", c)
		return
	}

	res.OkWithData(data, c)
}

// 获取站点配置信息-管理员
func (s SiteApi) SiteInfoAdminView(c *gin.Context) {
	cr := middleware.GetBindUri[SiteInfoRequest](c)

	var data any

	switch cr.Name {
	case "site":
		rep := site_service.GetRuntimeSite()
		rep.About.Version = global.Version
		data = rep
	case "ai":
		rep := site_service.GetRuntimeAI()
		rep.SecretKey = sensitive_place_holder
		data = rep
	default:
		res.FailWithMsg("站点信息不存在", c)
		return
	}

	res.OkWithData(data, c)
}
