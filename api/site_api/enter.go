// 站点API模块

package site_api

import (
	"myblogx/common/res"
	"myblogx/global"

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
	res.OkWithData(SiteAIResponse{
		Enable:   global.Config.AI.Enable,
		Nickname: global.Config.AI.Nickname,
		Avatar:   global.Config.AI.Avatar,
		Abstract: global.Config.AI.Abstract,
	}, c)
}
