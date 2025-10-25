package models

// 日志表
type LogModel struct {
	Model
	LogType   int8      `json:"log_type"`                // 日志类型
	Title     string    `gorm:"size:64" json:"title"`    // 日志标题
	Content   string    `gorm:"size:128" json:"content"` // 日志内容
	Level     int8      `json:"level"`                   // 日志级别
	UserID    uint      `json:"user_id"`                 // 用户ID
	UserModel UserModel `json:"user_model" gorm:"foreignKey:UserID;references:ID"`
	IP        string    `gorm:"size:32" json:"ip"`   // 操作IP
	Addr      string    `gorm:"size:64" json:"addr"` // 操作地址
	IsRead    bool      `json:"is_read"`             // 是否已读
}
