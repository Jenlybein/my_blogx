package models

// 用户文章查看历史表
type UserArticleViewHistoryModel struct {
	Model
	ArticleID    uint         `gorm:"uniqueIndex:idx_article_user" json:"article_id"`
	ArticleModel ArticleModel `gorm:"foreignKey:ArticleID;references:ID" json:"-"`
	UserID       uint         `gorm:"uniqueIndex:idx_article_user" json:"user_id"`
	UserModel    UserModel    `gorm:"foreignKey:UserID;references:ID" json:"-"`
}
