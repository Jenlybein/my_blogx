// 文章点赞模型

package models

import "time"

// 用户点赞表
type ArticleDiggModel struct {
	Model
	ArticleID    uint         `gorm:"uniqueIndex:idx_article_user" json:"article_id"`
	ArticleModel ArticleModel `gorm:"foreignKey:ArticleID;references:ID" json:"-"`
	UserID       uint         `gorm:"uniqueIndex:idx_article_user" json:"user_id"`
	UserModel    UserModel    `gorm:"foreignKey:UserID;references:ID" json:"-"`
	CreatedAt    time.Time    `json:"created_at"`
}
