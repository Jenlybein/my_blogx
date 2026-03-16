package search_api

import (
	"myblogx/common"
	"myblogx/models/enum"
	"time"
)

type ArticleSearchRequest struct {
	// 0 默认搜索 1 猜你喜欢 2 最新发布 3 最多回复
	// 4 最多点赞 5 最多收藏 6 最多浏览 7 标签匹配
	// 8 作者文章
	Type int8 `form:"type" binding:"required,oneof=0 1 2 3 4 5 6 7 8"`
	common.PageInfo
	TagList   []string `form:"tag_list"`
	UserID    uint     `form:"user_id"`
	TopSearch bool     `form:"top_search"` // 是否启用置顶优先搜索
}

type SearchListResponse struct {
	ID             uint               `json:"id"`
	CreatedAt      time.Time          `json:"created_at"`
	UpdatedAt      time.Time          `json:"updated_at"`
	Title          string             `json:"title"`
	Abstract       string             `json:"abstract"`
	HtmlContent    string             `json:"html_content"`
	Cover          string             `json:"cover"`
	ViewCount      int                `json:"view_count"`
	DiggCount      int                `json:"digg_count"`
	CommentCount   int                `json:"comment_count"`
	FavorCount     int                `json:"favor_count"`
	CommentsToggle bool               `json:"comments_toggle"`
	Status         enum.ArticleStatus `json:"status"`
	Tags           []string           `json:"tags"`
	UserTop        bool               `json:"user_top,omitempty"`  // 是否置顶
	AdminTop       bool               `json:"admin_top,omitempty"` // 是否管理员置顶
	CategoryTitle  string             `json:"category_title"`
	UserNickname   string             `json:"user_nickname"`
	UserAvatar     string             `json:"user_avatar"`
}
