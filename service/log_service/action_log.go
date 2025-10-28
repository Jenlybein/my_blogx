package log_service

import (
	"bytes"
	"io"
	"myblogx/core"
	"myblogx/global"
	"myblogx/models"
	"myblogx/models/enum"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type ActionLog struct {
	c            *gin.Context
	level        enum.LogLevelType
	title        string
	requestBody  []byte
	responseBody []byte
	log          *models.LogModel
}

func (ac *ActionLog) SetLevel(level enum.LogLevelType) {
	ac.level = level
}

func (ac *ActionLog) SetTitle(title string) {
	ac.title = title
}

func (ac *ActionLog) SetRequest(c *gin.Context) {
	byteData, err := io.ReadAll(c.Request.Body)
	if err != nil {
		logrus.Errorf("读取请求体失败: %v", err)
	}
	ac.requestBody = byteData
	// 请求内容阅后即焚，所以需要恢复请求体内容
	c.Request.Body = io.NopCloser(bytes.NewBuffer(byteData))
}

func (ac *ActionLog) SetResponse(data []byte) {
	ac.responseBody = data
}

func (ac *ActionLog) Save() {
	if ac.log != nil {
		// 日志已存在，直接更新
		global.DB.Model(ac.log).Updates(map[string]any{
			"title": "更新",
		})
		return
	}

	ip := ac.c.ClientIP()
	addr := core.GetIpAddr(ip)
	userID := uint(1)

	log := models.LogModel{
		LogType: enum.ActionLogType,
		Title:   ac.title,
		Content: "",
		Level:   ac.level,
		UserID:  userID,
		IP:      ip,
		Addr:    addr,
	}

	err := global.DB.Create(&log).Error
	if err != nil {
		logrus.Errorf("日志创建失败: %v", err)
		return
	}
	ac.log = &log
}

func NewActionLogByGin(c *gin.Context) *ActionLog {
	return &ActionLog{c: c}
}

func GetLog(c *gin.Context) *ActionLog {
	_log, ok := c.Get("log")
	if !ok {
		return NewActionLogByGin(c)
	}
	log, ok := _log.(*ActionLog)
	if !ok {
		return NewActionLogByGin(c)
	}
	return log
}
