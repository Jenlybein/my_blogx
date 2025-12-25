// 用户文章收藏模型

package models

import "time"

// 用户收藏表
type UserArticleFavorModel struct {
	Model
	ArticleID    uint         `gorm:"uniqueIndex:idx_article_user_favor" json:"article_id"`
	ArticleModel ArticleModel `gorm:"foreignKey:ArticleID;references:ID" json:"-"`
	UserID       uint         `gorm:"uniqueIndex:idx_article_user_favor" json:"user_id"`
	UserModel    UserModel    `gorm:"foreignKey:UserID;references:ID" json:"-"`
	FavorID      uint         `gorm:"uniqueIndex:idx_article_user_favor" json:"favor_id"`
	FavorModel   FavorModel   `gorm:"foreignKey:FavorID;references:ID" json:"-"`
	CreatedAt    time.Time    `json:"created_at"`
}
