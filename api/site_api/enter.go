package site_api

import (
	"fmt"
	"net/http"

	"myblogx/models/enum"
	"myblogx/service/log_service"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type SiteApi struct {
}

func (s SiteApi) SiteInfoView(c *gin.Context) {
	fmt.Println("1")
	log_service.NewLoginSuccess(c, enum.PasswordLoginType)
	log_service.NewLoginFail(c, enum.PasswordLoginType, "用户不存在", "用户名", "123456")
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "站点信息",
	})
}

type SiteUpdateRequest struct {
	Name string `json:"name"`
}

func (s SiteApi) SiteUpdateView(c *gin.Context) {
	log := log_service.GetLog(c)

	var cr SiteUpdateRequest
	err := c.ShouldBindJSON(&cr)
	if err != nil {
		logrus.Errorf("参数绑定失败: %v", err)
	}
	fmt.Println(cr)

	log.Save()

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "站点信息",
	})
}
