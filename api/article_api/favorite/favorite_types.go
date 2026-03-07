package favorite

import (
	"myblogx/common"
	"myblogx/models"
	"myblogx/models/enum"
	"time"
)

type FavoriteRequest struct {
	ID       uint   `json:"id"`
	Title    string `json:"title" binding:"required,min=2,max=32"`
	Cover    string `json:"cover"`
	Abstract string `json:"abstract" binding:"required,max=256"`
}

type FavoriteListRequest struct {
	common.PageInfo
	UserID uint `form:"user_id"`
	Type   int8 `form:"type" binding:"required,oneof=1 2 3"` // 1:查自己 2:查别人 3:管理员后台查
}

type FavoriteListResponse struct {
	models.FavoriteModel
	ArticleCount int    `json:"article_count"`
	Nickname     string `json:"nickname,omitempty"`
	Avatar       string `json:"avatar,omitempty"`
}

type FavoriteArticlesRequest struct {
	common.PageInfo
	FavoriteID uint `form:"favorite_id" binding:"required"`
}

type FavoriteArticleResponse struct {
	FavoritedAt   time.Time          `json:"favorited_at"`
	ArticleID     uint               `json:"article_id"`
	Title         string             `json:"title"`
	Abstract      string             `json:"abstract"`
	Cover         string             `json:"cover"`
	ViewCount     int                `json:"view_count"`
	DiggCount     int                `json:"digg_count"`
	CommentCount  int                `json:"comment_count"`
	FavorCount    int                `json:"favor_count"`
	UserNickname  string             `json:"user_nickname"`
	UserAvatar    string             `json:"user_avatar"`
	ArticleStatus enum.ArticleStatus `json:"article_status"`
}

type FavoriteRemovePatchModel struct {
	FavoriteID uint   `json:"favorite_id" binding:"required"`
	Articles   []uint `json:"articles" binding:"required"`
}
