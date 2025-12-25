// 用户置顶文章模型

package models

import "time"

// 用户置顶文章表
type UserTopArticleModel struct {
	UserID       uint         `gorm:"uniqueIndex:idx_article_user" json:"user_id"`
	UserModel    UserModel    `gorm:"foreignKey:UserID;references:ID" json:"-"`
	ArticleID    uint         `gorm:"uniqueIndex:idx_article_user" json:"article_id"`
	ArticleModel ArticleModel `gorm:"foreignKey:ArticleID;references:ID" json:"-"`
	CreatedAt    time.Time    `json:"created_at"`
}
