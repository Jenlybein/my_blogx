// 站点API模块

package site_api

import (
	"myblogx/common/res"
	"myblogx/global"

	"github.com/gin-gonic/gin"
)

type SiteApi struct {
}

// 站点信息请求参数
type SiteInfoRequest struct {
	Name string `uri:"name"`
}

// 敏感信息占位符
var sensitive_place_holder = "******"

// 站点 qq 登录地址
func (SiteApi) SiteInfoQQView(c *gin.Context) {
	res.OkWithData(global.Config.QQ.Url(), c)
}
