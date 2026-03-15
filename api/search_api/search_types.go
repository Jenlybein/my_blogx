package search_api

import (
	"myblogx/common"
	"myblogx/models"
)

type ArticleSearchRequest struct {
	common.PageInfo
	// 0 猜你喜欢 1 最新发布 2 最多回复 3 最多点赞 4 最多收藏
	Type int8 `form:"type" binding:"required,oneof=0 1 2 3 4"`
}

type ArticleSearchResponse struct {
	models.ArticleModel
	Highlight map[string][]string `json:"highlight"`
}
