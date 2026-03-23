// 用户文章查看历史模型

package models

// 用户文章查看历史表
type UserArticleViewHistoryModel struct {
	Model
	ArticleID    uint         `gorm:"uniqueIndex:uk_user_article_history,priority:1" json:"article_id"`
	UserID       uint         `gorm:"uniqueIndex:uk_user_article_history,priority:2" json:"user_id"`
	ArticleModel ArticleModel `gorm:"foreignKey:ArticleID;references:ID" json:"-"`
	UserModel    UserModel    `gorm:"foreignKey:UserID;references:ID" json:"-"`
}
