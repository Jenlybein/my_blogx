package ai_api

import "myblogx/service/ai_service/ai_metainfo"

type AIBaseRequest struct {
	Content string `json:"content" binding:"required"`
}

type AIBaseResponse struct {
	Content string `json:"content" binding:"required"`
}

type AIArticleMetaInfoResponse struct {
	Title    string                  `json:"title"`
	Abstract string                  `json:"abstract"`
	Category *ai_metainfo.Metainfos  `json:"category"`
	Tags     []ai_metainfo.Metainfos `json:"tags"`
}
