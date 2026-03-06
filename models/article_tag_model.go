package models

import "time"

// ArticleTagModel 文章和公共标签的关系表
type ArticleTagModel struct {
	ArticleID uint      `gorm:"primaryKey;index:idx_article_tag_article,priority:1" json:"article_id"`
	TagID     uint      `gorm:"primaryKey;index:idx_article_tag_tag,priority:1" json:"tag_id"`
	CreatedAt time.Time `json:"created_at"`
	Article   ArticleModel
	Tag       TagModel
}
