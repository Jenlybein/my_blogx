// 用户置顶文章模型

package models

// 用户置顶文章表
type UserTopArticleModel struct {
	Model
	UserID       uint         `gorm:"uniqueIndex:uk_user_top_article,priority:1" json:"user_id"`
	ArticleID    uint         `gorm:"uniqueIndex:uk_user_top_article,priority:2" json:"article_id"`
	UserModel    UserModel    `gorm:"foreignKey:UserID;references:ID" json:"-"`
	ArticleModel ArticleModel `gorm:"foreignKey:ArticleID;references:ID" json:"-"`
}
