// 文章点赞模型

package models

import "time"

// 用户点赞表
type ArticleDiggModel struct {
	ArticleID    uint         `gorm:"primaryKey" json:"article_id"`
	UserID       uint         `gorm:"primaryKey" json:"user_id"`
	CreatedAt    time.Time    `json:"created_at"`
	ArticleModel ArticleModel `gorm:"foreignKey:ArticleID;references:ID" json:"-"`
	UserModel    UserModel    `gorm:"foreignKey:UserID;references:ID" json:"-"`
}
