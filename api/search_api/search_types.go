package search_api

import (
	"myblogx/common"
	"myblogx/models/enum"
	"time"
)

type ArticleSearchRequest struct {
	common.PageInfo
	// 0 猜你喜欢 1 最新发布 2 最多回复 3 最多点赞 4 最多收藏
	Type int8 `form:"type" binding:"required,oneof=0 1 2 3 4"`
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
	UserTop        bool               `json:"user_top"`  // 是否置顶
	AdminTop       bool               `json:"admin_top"` // 是否管理员置顶
	CategoryTitle  string             `json:"category_title"`
	UserNickname   string             `json:"user_nickname"`
	UserAvatar     string             `json:"user_avatar"`
}
