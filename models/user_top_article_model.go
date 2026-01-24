// 用户置顶文章模型

package models

import "time"

// 用户置顶文章表
type UserTopArticleModel struct {
	UserID       uint         `gorm:"primaryKey" json:"user_id"`
	ArticleID    uint         `gorm:"primaryKey" json:"article_id"`
	UserModel    UserModel    `gorm:"foreignKey:UserID;references:ID" json:"-"`
	ArticleModel ArticleModel `gorm:"foreignKey:ArticleID;references:ID" json:"-"`
	CreatedAt    time.Time    `json:"created_at"`
}
