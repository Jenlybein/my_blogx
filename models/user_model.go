package models

import (
	"time"
)

// 用户表
type UserModel struct {
	Model
	Username       string `gorm:"size:32" json:"username"`
	Password       string `gorm:"size:64" json:"password"`
	Nickname       string `gorm:"size:32" json:"nickname"`
	Avatar         string `gorm:"size:256" json:"avatar"`
	Abstract       string `gorm:"size:256" json:"abstract"`
	RegisterSource string `json:"register_source"`
	CodeAge        int    `json:"code_age"`
	LikeTags       string `gorm:"type:longtext;serializer:json" json:"like_tags"`
	Email          string `gorm:"size:256" json:"email"`
	OpenID         string `gorm:"size:64" json:"open_id"`
	Role           int    `gorm:"default:0" json:"role"` // 角色 0:管理员 1:普通用户 2:访客
}

type UserConfModel struct {
	UserID              uint      `gorm:"primaryKey" json:"user_id"`
	UserModel           UserModel `gorm:"foreignKey:UserID;references:ID" json:"-"`
	LikeTags            []string  `gorm:"type:longtext;serializer:json" json:"like_tags"`
	UpdatedUsernameDate time.Time `json:"updated_username_date"` // 上次修改用户名的时间
	FavoritesVisibility bool      `json:"favorites_visibility"`  // 收藏夹是否可见
	FollowersVisibility bool      `json:"followers_visibility"`  // 关注是否可见
	FansVisibility      bool      `json:"fans_visibility"`       // 粉丝是否可见
	HomeStyleID         uint      `json:"home_style_id"`         // 首页样式ID
}
