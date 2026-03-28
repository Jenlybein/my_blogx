// 站点API模块

package site_api

import (
	"myblogx/common/res"
	"myblogx/global"
	"myblogx/service/site_service"

	"github.com/gin-gonic/gin"
)

type SiteApi struct {
}

// 敏感信息占位符
var sensitive_place_holder = "******"

// 站点 qq 登录地址
func (SiteApi) SiteInfoQQView(c *gin.Context) {
	res.OkWithData(global.Config.QQ.Url(), c)
}

// AI 信息获取
func (SiteApi) SiteInfoAIView(c *gin.Context) {
	ai := site_service.GetRuntimeAI()
	res.OkWithData(SiteAIResponse{
		Enable:   ai.Enable,
		Nickname: ai.Nickname,
		Avatar:   ai.Avatar,
		Abstract: ai.Abstract,
	}, c)
}

// SEO 信息获取，适合 Nuxt 等前端在服务端渲染或页面切换时拉取。
func (SiteApi) SiteSEOView(c *gin.Context) {
	site := site_service.GetRuntimeSite()
	res.OkWithData(SiteSEOResponse{
		SiteTitle:    site.SiteInfo.Title,
		ProjectTitle: site.Project.Title,
		Logo:         site.SiteInfo.Logo,
		Icon:         site.Project.Icon,
		Keywords:     site.Seo.Keywords,
		Description:  site.Seo.Description,
	}, c)
}
