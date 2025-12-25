// 日志模型

package models

import "myblogx/models/enum"

// 日志表
type LogModel struct {
	Model
	LogType     enum.LogType      `json:"log_type"`                     // 日志类型
	Title       string            `gorm:"size:64" json:"title"`         // 日志标题
	Content     string            `gorm:"type:longtext" json:"content"` // 日志内容
	Level       enum.LogLevelType `json:"level"`                        // 日志级别
	UserID      uint              `json:"user_id"`                      // 用户ID
	UserModel   UserModel         `json:"user_model" gorm:"foreignKey:UserID;references:ID"`
	IP          string            `gorm:"size:32" json:"ip"`           // 操作IP
	Addr        string            `gorm:"size:64" json:"addr"`         // 操作地址
	IsRead      bool              `json:"is_read"`                     // 是否已读
	LoginStatus bool              `json:"login_status"`                // 登录状态
	Username    string            `gorm:"size:32" json:"username"`     // 登录日志的用户名
	Password    string            `gorm:"size:32" json:"password"`     // 登录日志的密码
	LoginType   enum.LoginType    `json:"login_type"`                  // 登录类型
	ServiceName string            `gorm:"size:32" json:"service_name"` // 服务名称
}
