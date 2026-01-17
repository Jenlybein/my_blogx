// 文章模型

package models

import (
	_ "embed"
	"myblogx/global"
)

// 文章表
type ArticleModel struct {
	Model
	Title          string    `gorm:"size:256" json:"title"`
	Abstract       string    `gorm:"size:256" json:"abstract"`
	Content        string    `gorm:"type:longtext" json:"content"`
	CategoryID     uint      `gorm:"index" json:"category_id"`
	TagList        []string  `gorm:"type:longtext;serializer:json" json:"tag_list"`
	Cover          string    `gorm:"size:256" json:"cover"`
	AuthorID       uint      `gorm:"index" json:"author_id"`
	UserModel      UserModel `gorm:"foreignKey:AuthorID;references:ID" json:"-"`
	ViewCount      uint      `gorm:"default:0" json:"view_count"`         // 查看次数
	DiggCount      uint      `gorm:"default:0" json:"digg_count"`         // 点赞次数
	CommentCount   uint      `gorm:"default:0" json:"comment_count"`      // 评论次数
	FavorCount     uint      `gorm:"default:0" json:"favor_count"`        // 收藏次数
	CommentsToggle bool      `gorm:"default:true" json:"comments_toggle"` // 是否允许评论
	Status         int       `gorm:"default:0" json:"status"`             // 状态 0: 草稿 1: 审核中 2: 已发布
}

//go:embed mappings/article_mapping.json
var ArticleMapping string

func (ArticleModel) Mapping() string {
	return ArticleMapping
}

func (ArticleModel) Index() string {
	return global.Config.ES.Index
}
