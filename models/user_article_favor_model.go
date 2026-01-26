// 用户文章收藏模型

package models

import "time"

// 用户收藏表
type UserArticleFavorModel struct {
	Model
	ArticleID     uint          `gorm:"primaryKey" json:"article_id"`
	UserID        uint          `gorm:"primaryKey" json:"user_id"`
	FavorID       uint          `gorm:"primaryKey" json:"favor_id"`
	CreatedAt     time.Time     `json:"created_at"`
	UserModel     UserModel     `gorm:"foreignKey:UserID;references:ID" json:"-"`
	ArticleModel  ArticleModel  `gorm:"foreignKey:ArticleID;references:ID" json:"-"`
	FavoriteModel FavoriteModel `gorm:"foreignKey:FavorID;references:ID" json:"-"`
}
