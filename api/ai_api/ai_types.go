package ai_api

import "myblogx/service/ai_service"

type AIArticleMetaInfoRequest struct {
	Content string `json:"content" binding:"required"`
}

type AIArticleMetaInfoResponse struct {
	Title    string                 `json:"title"`
	Abstract string                 `json:"abstract"`
	Category *ai_service.Metainfos  `json:"category"`
	Tags     []ai_service.Metainfos `json:"tags"`
}
