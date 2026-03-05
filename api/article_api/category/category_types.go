package category

import (
	"myblogx/common"
	"myblogx/models"
)

type CategoryRequest struct {
	ID    uint   `json:"id"`
	Title string `json:"title" binding:"required,min=2,max=20"`
}

type CategoryListRequest struct {
	common.PageInfo
	UserID uint `form:"user_id"`
	Type   int8 `form:"type" binding:"required,oneof=1 2 3"` // 1:查自己 2:查别人 3:管理员后台查
}

type CategoryListResponse struct {
	models.CategoryModel
	ArticleCount int    `json:"article_count"`
	Nickname     string `json:"nickname,omitempty"`
	Avatar       string `json:"avatar,omitempty"`
}
