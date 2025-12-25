// 分类模型

package models

// 分类表
type CategoryModel struct {
	Model
	Title     string    `gorm:"size:256" json:"title"`
	UserID    uint      `gorm:"index" json:"user_id"`
	UserModel UserModel `gorm:"foreignKey:UserID;references:ID" json:"-"`
}
