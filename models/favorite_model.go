// 收藏模型

package models

// 收藏表
type FavoriteModel struct {
	Model
	UserID       uint                    `gorm:"index" json:"user_id"`
	Title        string                  `gorm:"size:32" json:"title"`
	Cover        string                  `gorm:"size:256" json:"cover"`
	Abstract     string                  `gorm:"size:256" json:"abstract"`
	ArticleCount int                     `gorm:"default:0" json:"article_count"`
	IsDefault    bool                    `gorm:"default:false" json:"is_default"`
	UserModel    UserModel               `gorm:"foreignKey:UserID;references:ID" json:"-"`
	ArticleList  []UserArticleFavorModel `gorm:"foreignKey:FavorID" json:"-"`
}
