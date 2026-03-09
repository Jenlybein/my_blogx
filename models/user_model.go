// 用户模型

package models

import (
	"myblogx/models/enum"
	"time"

	"gorm.io/gorm"
)

// 用户表
type UserModel struct {
	Model
	Username       string                  `gorm:"size:32" json:"username"`
	Nickname       string                  `gorm:"size:32" json:"nickname"`
	Avatar         string                  `gorm:"size:256" json:"avatar"`
	Abstract       string                  `gorm:"size:256" json:"abstract"`
	RegisterSource enum.RegisterSourceType `json:"register_source"` // 注册来源
	Password       string                  `gorm:"size:64" json:"-"`
	Email          string                  `gorm:"size:256" json:"email"`
	OpenID         string                  `gorm:"size:64" json:"open_id"` // qq 登录的 openid
	Role           enum.RoleType           `gorm:"default:0" json:"role"`
	IP             string                  `gorm:"size:64" json:"ip"`    // 注册时的 IP
	Addr           string                  `gorm:"size:256" json:"addr"` // 注册时的地址
	UserConfModel  *UserConfModel          `gorm:"foreignKey:UserID;" json:"-"`
}

// 创建用户配置表
func (u *UserModel) AfterCreate(tx *gorm.DB) (err error) {
	u.UserConfModel = &UserConfModel{
		UserID:                   u.ID,
		FavoritesVisibility:      true,
		FollowVisibility:         true,
		FansVisibility:           true,
		HomeStyleID:              1,
		DiggNoticeEnabled:        true,
		CommentNoticeEnabled:     true,
		FavorNoticeEnabled:       true,
		PrivateChatNoticeEnabled: true,
	}
	if err = tx.Create(u.UserConfModel).Error; err != nil {
		return err
	}
	return nil
}

// CodeAge 计算用户注册年龄（单位：年）
func (u *UserModel) CodeAge() int {
	return int(time.Since(u.CreatedAt).Hours() / 24 / 365)
}

type UserConfModel struct {
	UserID                   uint       `gorm:"primaryKey" json:"user_id"`
	UserModel                UserModel  `gorm:"foreignKey:UserID;references:ID" json:"-"`
	LikeTags                 []string   `gorm:"type:longtext;serializer:json" json:"like_tags"`
	UpdatedUsernameDate      *time.Time `json:"updated_username_date"` // 上次修改用户名的时间
	FavoritesVisibility      bool       `json:"favorites_visibility"`  // 收藏夹是否可见
	FollowVisibility         bool       `json:"followers_visibility"`  // 关注是否可见
	FansVisibility           bool       `json:"fans_visibility"`       // 粉丝是否可见
	HomeStyleID              uint       `json:"home_style_id"`         // 首页样式ID
	DiggNoticeEnabled        bool       `json:"digg_notice_enabled"`   // 是否开启点赞通知
	CommentNoticeEnabled     bool       `json:"comment_notice_enabled"`
	FavorNoticeEnabled       bool       `json:"favor_notice_enabled"`
	PrivateChatNoticeEnabled bool       `json:"private_chat_notice_enabled"`
}
