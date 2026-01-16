// 操作日志服务

package log_service

import (
	"encoding/json"
	"fmt"
	"myblogx/core"
	"myblogx/global"
	"myblogx/models"
	"myblogx/models/enum"
	io_utils "myblogx/utils/io_util"
	"myblogx/utils/jwts"
	"net/http"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

type ActionLog struct {
	c                  *gin.Context
	level              enum.LogLevelType
	title              string
	requestBody        []byte
	responseBody       []byte
	log                *models.LogModel
	showRequest        bool
	showResponse       bool
	showRequestHeader  bool
	showResponseHeader bool
	itemList           []string
	responseHeader     http.Header
	isMiddleware       bool
}

func (ac *ActionLog) SetShowRequest() {
	ac.showRequest = true
}

func (ac *ActionLog) SetShowResponse() {
	ac.showResponse = true
}

func (ac *ActionLog) ShowRequestHeader() {
	ac.showRequestHeader = true
}

func (ac *ActionLog) ShowResponseHeader() {
	ac.showResponseHeader = true
}

func (ac *ActionLog) SetLevel(level enum.LogLevelType) {
	ac.level = level
}

func (ac *ActionLog) SetTitle(title string) {
	ac.title = title
}

func (ac *ActionLog) SetLink(label string, href string) {
	ac.itemList = append(ac.itemList, fmt.Sprintf("[%s](%s)", label, href))
}

func (ac *ActionLog) SetImage(src string) {
	ac.itemList = append(ac.itemList, fmt.Sprintf("![%s](%s)", src, src))
}

func (ac *ActionLog) setItem(label string, value any, LogLevelType enum.LogLevelType) {
	var v string
	t := reflect.TypeOf(value)
	switch t.Kind() {
	case reflect.Struct, reflect.Map, reflect.Slice:
		byteData, err := json.Marshal(value)
		if err != nil {
			global.Logger.Errorf("JSON 序列化失败: %v", err)
		}
		v = string(byteData)
	default:
		v = fmt.Sprintf("%v", value)
	}

	item := fmt.Sprintf("[%s]%s: %s", LogLevelType.String(), label, v)
	ac.itemList = append(ac.itemList, item)
}
func (ac *ActionLog) SetItemInfo(label string, value any) {
	ac.setItem(label, value, enum.LogInfoLevel)
}
func (ac *ActionLog) SetItemWarn(label string, value any) {
	ac.setItem(label, value, enum.LogWarnLevel)
}
func (ac *ActionLog) SetItemError(label string, value any) {
	ac.setItem(label, value, enum.LogErrorLevel)
}

func (ac *ActionLog) SetError(label string, err error) {
	msg := errors.WithStack(err)
	global.Logger.Errorf("%s: %v", label, err.Error())
	ac.itemList = append(ac.itemList, fmt.Sprintf("[%s:%T]%s: %+v", label, err, err, msg))
}

func (ac *ActionLog) SetResponseHeader(c *gin.Context) {
	ac.responseHeader = c.Writer.Header()
}

func (ac *ActionLog) SetRequest(c *gin.Context) {
	byteData, err := io_utils.GetBody(&c.Request.Body)
	if err != nil {
		global.Logger.Errorf("读取请求体失败: %v", err)
	}
	ac.requestBody = byteData
}

func (ac *ActionLog) SetResponse(data []byte) {
	ac.responseBody = data
}

func (ac *ActionLog) MiddlewareSave() {
	_saveLog, _ := ac.c.Get("isSaveLog")
	saveLog, _ := _saveLog.(bool)
	if !saveLog {
		return
	}

	if ac.log == nil {
		// 创建
		ac.isMiddleware = true
		ac.Save()
		return
	}
	// 在视图里 Save 过，更新
	// 响应头
	if ac.showResponseHeader {
		byteData, _ := json.Marshal(ac.responseHeader)
		ac.itemList = append(ac.itemList, "响应头: "+string(byteData))
	}

	// 设置响应
	if ac.showResponse {
		ac.itemList = append(ac.itemList, "响应体: "+string(ac.responseBody))
	}
	ac.Save()
}

func (ac *ActionLog) Save() (id uint) {
	if ac.log != nil {
		// 日志已存在，直接更新
		newContent := strings.Join(ac.itemList, "\n")
		content := ac.log.Content + "\n" + newContent
		global.DB.Model(ac.log).Updates(map[string]any{
			"content": content,
		})
		ac.itemList = []string{}
		return ac.log.ID
	}

	var newItemList []string

	// 请求头
	if ac.showRequestHeader {
		byteData, _ := json.Marshal(ac.c.Request.Header)
		newItemList = append(newItemList, "请求头: "+string(byteData))
	}

	// 设置请求
	if ac.showRequest {
		newItemList = append(newItemList, "请求体: "+string(ac.requestBody))
	}

	if ac.isMiddleware {
		// 响应头
		if ac.showResponseHeader {
			byteData, _ := json.Marshal(ac.responseHeader)
			ac.itemList = append(ac.itemList, "响应头: "+string(byteData))
		}

		// 设置响应
		if ac.showResponse {
			ac.itemList = append(ac.itemList, "响应体: "+string(ac.responseBody))
		}
	}

	// 中间的一些 content
	newItemList = append(newItemList, ac.itemList...)

	ip := ac.c.ClientIP()
	addr := core.GetIpAddr(ip)

	// 解析 jwt token 中的 userID
	userID := uint(0)
	claims, err := jwts.ParseTokenByGin(ac.c)
	if err != nil {
		global.Logger.Errorf("解析 token 失败: %v", err)
	} else {
		userID = uint(claims.UserID)
	}

	log := models.LogModel{
		LogType: enum.ActionLogType,
		Title:   ac.title,
		Content: strings.Join(newItemList, "\n"),
		Level:   ac.level,
		UserID:  userID,
		IP:      ip,
		Addr:    addr,
	}

	err = global.DB.Create(&log).Error
	if err != nil {
		global.Logger.Errorf("日志创建失败: %v", err)
		return
	}
	ac.log = &log
	ac.itemList = []string{}
	return log.ID
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

	c.Set("isSaveLog", true)

	return log
}
