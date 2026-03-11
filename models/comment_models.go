// 评论模型

package models

import (
	"myblogx/models/enum"
	"time"
)

// 评论表
type CommentModel struct {
	Model
	Content      string             `json:"content" gorm:"type:text;not null"`
	UserID       uint               `json:"user_id"`
	UserModel    UserModel          `json:"user_model" gorm:"foreignKey:UserID;references:ID"`
	ArticleID    uint               `json:"article_id" gorm:"not null;index:idx_article_root"`
	ReplyId      uint               `json:"reply_id" gorm:"default:0"`                       // 回复的评论id，0表示一级评论
	RootID       uint               `json:"root_id" gorm:"default:0;index:idx_article_root"` // 根评论ID，0表示本身就是一级评论
	DiggCount    int                `json:"digg_count" gorm:"default:0"`                     // 点赞数
	ReplyCount   int                `json:"reply_count" gorm:"default:0"`
	Status       enum.CommentStatus `gorm:"default:0" json:"status"`
	ArticleModel ArticleModel       `json:"article_model" gorm:"foreignKey:ArticleID;references:ID"`
	ParentModel  *CommentModel      `json:"parent_model" gorm:"foreignKey:ReplyId;references:ID"`
}

// 用户点赞表
type CommentDiggModel struct {
	CommentID    uint         `gorm:"primaryKey" json:"comment_id"`
	UserID       uint         `gorm:"primaryKey" json:"user_id"`
	CreatedAt    time.Time    `json:"created_at"`
	UserModel    UserModel    `gorm:"foreignKey:UserID;references:ID" json:"-"`
	CommentModel CommentModel `gorm:"foreignKey:CommentID;references:ID" json:"-"`
}
