package models

// 全局通知表
type GlobalNotificationModel struct {
	Model
	Title   string `gorm:"size:64" json:"title"`    // 通知标题
	Icon    string `gorm:"size:64" json:"icon"`     // 通知图标
	Content string `gorm:"size:128" json:"content"` // 通知内容
	Herf    string `gorm:"size:256" json:"herf"`    // 通知链接
}
