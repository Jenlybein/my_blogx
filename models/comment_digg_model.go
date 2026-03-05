package models

import "time"

type CommentDiggModel struct {
	CommentID    uint         `gorm:"primaryKey" json:"comment_id"`
	UserID       uint         `gorm:"primaryKey" json:"user_id"`
	CreatedAt    time.Time    `json:"created_at"`
	UserModel    UserModel    `gorm:"foreignKey:UserID;references:ID" json:"-"`
	CommentModel CommentModel `gorm:"foreignKey:CommentID;references:ID" json:"-"`
}
