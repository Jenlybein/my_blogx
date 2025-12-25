// 站点API模块

package site_api

import (
	"myblogx/common/res"
	"myblogx/service/log_service"

	"github.com/gin-gonic/gin"
)

type SiteApi struct {
}

func (s SiteApi) SiteInfoView(c *gin.Context) {
	res.OkWithData("xxxxx", c)
}

type SiteUpdateRequest struct {
	Name string `json:"name" binding:"required"`
	Age  int    `json:"age" binding:"required"`
}

func (s SiteApi) SiteUpdateView(c *gin.Context) {
	log := log_service.GetLog(c)

	var cr SiteUpdateRequest
	err := c.ShouldBindJSON(&cr)
	if err != nil {
		log.SetError("参数绑定失败", err)
		res.FailWithError(err, c)
		return
	}

	res.OkWithMsg("站点信息更新成功", c)
}
