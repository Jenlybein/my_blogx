package search_service

import (
	"myblogx/common"
	"myblogx/models/ctype"
	"myblogx/models/enum"
	"myblogx/utils/markdown"
	"time"
)

type ArticleSearchRequest struct {
	// Type
	// 1 普通搜索 2 猜你喜欢 3 作者文章 4 自己文章 5 管理员搜
	// Sort
	// 1 默认搜索 2 最新发布 3 最多回复
	// 4 最多点赞 5 最多收藏 6 最多浏览
	common.PageInfo
	Type       int8               `form:"type" binding:"required,oneof=1 2 3 4 5"`
	Sort       int8               `form:"sort" binding:"required,oneof=1 2 3 4 5 6"`
	TagList    []string           `form:"tag_list"`
	CategoryID ctype.ID           `form:"category_id"`
	UserID     ctype.ID           `form:"user_id"`
	TopSearch  bool               `form:"top_search"` // 是否启用置顶优先搜索
	Status     enum.ArticleStatus `form:"status"`
	Key        string             `form:"key"`
}

type SearchListResponse struct {
	ID             ctype.ID               `json:"id"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
	Title          string                 `json:"title"`
	Abstract       string                 `json:"abstract,omitempty"`
	Content        string                 `json:"content,omitempty"`
	Part           []markdown.ContentPart `json:"part,omitempty"`
	Cover          string                 `json:"cover"`
	ViewCount      int                    `json:"view_count"`
	DiggCount      int                    `json:"digg_count"`
	CommentCount   int                    `json:"comment_count"`
	FavorCount     int                    `json:"favor_count"`
	CommentsToggle bool                   `json:"comments_toggle"`
	Status         enum.ArticleStatus     `json:"status"`
	Tags           []string               `json:"tags"`
	UserTop        bool                   `json:"user_top,omitempty"`  // 是否置顶
	AdminTop       bool                   `json:"admin_top,omitempty"` // 是否管理员置顶
	CategoryTitle  string                 `json:"category_title"`
	UserNickname   string                 `json:"user_nickname"`
	UserAvatar     string                 `json:"user_avatar"`
}
