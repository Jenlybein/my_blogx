package log_service

import (
	"encoding/json"
	"fmt"
	"myblogx/global"
	"myblogx/models"
	"myblogx/models/enum"
	"reflect"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type RuntimeLog struct {
	level           enum.LogLevelType
	title           string
	itemList        []string
	serviceName     string
	RuntimeDateType RuntimeDateType
}

func (ac *RuntimeLog) Save() {
	// 判断是更新还是创建
	var log models.LogModel

	global.DB.Find(
		&log,
		fmt.Sprintf("service_name = ? and log_type = ? and created_at >= date_sub(now(), %s)", ac.RuntimeDateType.GetSqlTime()),
		ac.serviceName, enum.RuntimeLogType)

	if log.ID != 0 {
		// 日志已存在，直接更新
		newContent := strings.Join(ac.itemList, "\n")
		content := log.Content + "\n" + newContent
		global.DB.Model(&log).Updates(map[string]any{
			"content": content,
		})
		ac.itemList = []string{}
		return
	}
	err := global.DB.Create(&models.LogModel{
		LogType:     enum.RuntimeLogType,
		Title:       ac.title,
		Content:     strings.Join(ac.itemList, "\n"),
		Level:       ac.level,
		ServiceName: ac.serviceName,
	}).Error
	if err != nil {
		logrus.Errorf("保存运行时日志失败: %v", err)
		return
	}
	ac.itemList = []string{}
}

func (ac *RuntimeLog) SetLevel(level enum.LogLevelType) {
	ac.level = level
}

func (ac *RuntimeLog) SetTitle(title string) {
	ac.title = title
}

func (ac *RuntimeLog) SetLink(label string, href string) {
	ac.itemList = append(ac.itemList, fmt.Sprintf("[%s](%s)", label, href))
}

func (ac *RuntimeLog) SetImage(src string) {
	ac.itemList = append(ac.itemList, fmt.Sprintf("![%s](%s)", src, src))
}

func (ac *RuntimeLog) setItem(label string, value any, LogLevelType enum.LogLevelType) {
	var v string
	t := reflect.TypeOf(value)
	switch t.Kind() {
	case reflect.Struct, reflect.Map, reflect.Slice:
		byteData, err := json.Marshal(value)
		if err != nil {
			logrus.Errorf("JSON 序列化失败: %v", err)
		}
		v = string(byteData)
	default:
		v = fmt.Sprintf("%v", value)
	}

	item := fmt.Sprintf("[%s]%s: %s", LogLevelType.String(), label, v)
	ac.itemList = append(ac.itemList, item)
}
func (ac *RuntimeLog) SetItem(label string, value any) {
	ac.setItem(label, value, enum.LogInfoLevel)
}
func (ac *RuntimeLog) SetItemInfo(label string, value any) {
	ac.setItem(label, value, enum.LogInfoLevel)
}
func (ac *RuntimeLog) SetItemWarn(label string, value any) {
	ac.setItem(label, value, enum.LogWarnLevel)
}
func (ac *RuntimeLog) SetItemError(label string, value any) {
	ac.setItem(label, value, enum.LogErrorLevel)
}

func (ac *RuntimeLog) SetError(label string, err error) {
	msg := errors.WithStack(err)
	logrus.Errorf("%s: %v", label, err.Error())
	ac.itemList = append(ac.itemList, fmt.Sprintf("[%s:%T]%s: %+v", label, err, err, msg))
}

func (ac *RuntimeLog) SetNowTime() {
	ac.itemList = append(ac.itemList, fmt.Sprintf("当前时间: %s", time.Now().Format("2006-01-02 15:04:05")))
}

type RuntimeDateType int8

const (
	RuntimeDateHour  RuntimeDateType = 1
	RuntimeDateDay   RuntimeDateType = 2 // 按天分割
	RuntimeDateWeek  RuntimeDateType = 3
	RuntimeDateMonth RuntimeDateType = 4
)

func (r RuntimeDateType) GetSqlTime() string {
	switch r {
	case RuntimeDateHour:
		return "interval 1 HOUR"
	case RuntimeDateDay:
		return "interval 1 DAY"
	case RuntimeDateWeek:
		return "interval 1 WEEK"
	case RuntimeDateMonth:
		return "interval 1 MONTH"
	}
	return "interval 1 DAY"
}

func NewRuntimeLog(serviceName string, dateType RuntimeDateType) *RuntimeLog {
	return &RuntimeLog{
		serviceName:     serviceName,
		RuntimeDateType: dateType,
	}
}
