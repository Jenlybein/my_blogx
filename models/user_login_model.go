// 用户登录模型

package models

import "myblogx/models/ctype"

// 用户登录表(记录用户登录信息)
type UserLoginModel struct {
	Model
	UserID    ctype.ID  `json:"user_id"`
	UserModel UserModel `gorm:"foreignKey:UserID" json:"user_model"`
	IP        string    `gorm:"size:32" json:"ip"`
	Addr      string    `gorm:"size:64" json:"addr"`
	UA        string    `gorm:"size:128" json:"ua"` // 用户代理
}
