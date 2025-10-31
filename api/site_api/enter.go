package site_api

import (
	"net/http"
	"time"

	"myblogx/models/enum"
	"myblogx/service/log_service"

	"github.com/gin-gonic/gin"
)

type SiteApi struct {
}

func (s SiteApi) SiteInfoView(c *gin.Context) {
	log_service.NewLoginSuccess(c, enum.PasswordLoginType)
	log_service.NewLoginFail(c, enum.PasswordLoginType, "用户不存在", "用户名", "123456")
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "站点信息",
	})
}

type SiteUpdateRequest struct {
	Name string `json:"name" binding:"required"`
}

func (s SiteApi) SiteUpdateView(c *gin.Context) {
	log := log_service.GetLog(c)

	log.SetShowRequest()
	log.SetShowResponse()
	log.ShowRequestHeader()
	log.ShowResponseHeader()

	log.SetTitle("更新站点信息")
	log.SetItemInfo("站点时间", time.Now())

	log.SetImage("https://wx2.sinaimg.cn/mw690/896ff430gy1hqs9fgf4yhj20u01t0q7p.jpg")
	log.SetLink("百度一下", "https://www.baidu.com")

	var cr SiteUpdateRequest
	err := c.ShouldBindJSON(&cr)
	if err != nil {
		log.SetError("参数绑定失败", err)
	}

	log.SetItemInfo("结构体", cr)
	log.SetItemInfo("切片", []string{"b", "c", "d"})
	log.SetItemInfo("映射", map[string]string{"a": "1", "b": "2"})
	log.SetItemInfo("整数", 100)
	log.SetItemInfo("字符串", "hello")

	id := log.Save()

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "站点信息",
		"id":   id,
	})

}
