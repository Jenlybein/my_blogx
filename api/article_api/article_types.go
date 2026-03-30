package article_api

import (
	"myblogx/common"
	"myblogx/models/ctype"
	"myblogx/models/enum"
	"time"
)

type ArticleCreateRequest struct {
	Title          string             `json:"title" binding:"required"`
	Abstract       string             `json:"abstract"`
	Content        string             `json:"content" binding:"required"`
	CategoryID     *ctype.ID          `json:"category_id"`
	TagIDs         []ctype.ID         `json:"tag_ids"`
	Cover          string             `json:"cover"`
	CommentsToggle bool               `json:"comments_toggle"`
	Status         enum.ArticleStatus `json:"status" binding:"required,oneof=1 2"`
}

type ArticleDetailResponse struct {
	ID             ctype.ID           `gorm:"primaryKey" json:"id"`
	CreatedAt      time.Time          `json:"created_at"`
	UpdatedAt      time.Time          `json:"updated_at"`
	Title          string             `json:"title"`
	Abstract       string             `json:"abstract"`
	Content        string             `json:"content"`
	Cover          string             `json:"cover"`
	ViewCount      int                `json:"view_count"`
	DiggCount      int                `json:"digg_count"`
	CommentCount   int                `json:"comment_count"`
	FavorCount     int                `json:"favor_count"`
	CommentsToggle bool               `json:"comments_toggle"`
	Status         enum.ArticleStatus `json:"status"`
	Tags           []string           `json:"tags"`
	AuthorAvatar   string             `json:"author_avatar"`
	AuthorNickname string             `json:"author_name"`
	AuthorUsername string             `json:"author_username"`
	CategoryName   string             `json:"category_name"`
	IsDigg         bool               `json:"is_digg"`
	IsFavor        bool               `json:"is_favor"`
}

type ArticleExamineRequest struct {
	Status enum.ArticleStatus `json:"status" binding:"required,oneof=3 4"`
	Reason string             `json:"reason"`
}

type ArticleFavoriteRequest struct {
	ArticleID ctype.ID `json:"article_id" binding:"required"`
	FavorID   ctype.ID `json:"favor_id"`
}

type ArticleListRequest struct {
	common.PageInfo
	// 1 查自己的文章，2 查别人的文章，3 管理员查文章
	Type       int8               `form:"type" binding:"required,oneof=1 2 3"`
	UserID     ctype.ID           `form:"user_id"`
	CategoryID *ctype.ID          `form:"category_id"`
	TagID      *ctype.ID          `form:"tag_id"`
	Status     enum.ArticleStatus `form:"status"`
}

type ArticleListResponse struct {
	ID             ctype.ID           `json:"id"`
	CreatedAt      time.Time          `json:"created_at"`
	UpdatedAt      time.Time          `json:"updated_at"`
	Title          string             `json:"title"`
	Abstract       string             `json:"abstract"`
	Content        string             `json:"content"`
	Cover          string             `json:"cover"`
	ViewCount      int                `json:"view_count"`
	DiggCount      int                `json:"digg_count"`
	CommentCount   int                `json:"comment_count"`
	FavorCount     int                `json:"favor_count"`
	CommentsToggle bool               `json:"comments_toggle"`
	Status         enum.ArticleStatus `json:"status"`
	Tags           []string           `json:"tags"`
	UserTop        bool               `json:"user_top"`  // 是否置顶
	AdminTop       bool               `json:"admin_top"` // 是否管理员置顶
	CategoryTitle  string             `json:"category_title"`
	UserNickname   string             `json:"user_nickname"`
	UserAvatar     string             `json:"user_avatar"`
}

type ArticleUpdateRequest struct {
	Title          *string     `json:"title"`
	Abstract       *string     `json:"abstract"`
	Content        *string     `json:"content"`
	CategoryID     *ctype.ID   `json:"category_id"`
	TagIDs         *[]ctype.ID `json:"tag_ids"`
	Cover          *string     `json:"cover"`
	CommentsToggle *bool       `json:"comments_toggle"`
}

type ArticleViewCountRequest struct {
	ArticleID ctype.ID `json:"article_id" binding:"required"`
}
