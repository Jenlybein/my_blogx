// 用户文章收藏模型

package models

import "myblogx/models/ctype"

// 用户收藏表
type UserArticleFavorModel struct {
	Model
	ArticleID     ctype.ID      `gorm:"uniqueIndex:uk_user_article_favor,priority:1;index:idx_article_user,priority:1" json:"article_id"`
	UserID        ctype.ID      `gorm:"uniqueIndex:uk_user_article_favor,priority:2;index:idx_article_user,priority:2" json:"user_id"`
	FavorID       ctype.ID      `gorm:"uniqueIndex:uk_user_article_favor,priority:3;index:idx_article_user,priority:3" json:"favor_id"`
	UserModel     UserModel     `gorm:"foreignKey:UserID;references:ID" json:"-"`
	ArticleModel  ArticleModel  `gorm:"foreignKey:ArticleID;references:ID" json:"-"`
	FavoriteModel FavoriteModel `gorm:"foreignKey:FavorID;references:ID" json:"-"`
}
