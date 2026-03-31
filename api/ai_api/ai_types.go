package ai_api

import (
	"myblogx/service/ai_service/ai_diagnose"
	"myblogx/service/ai_service/ai_metainfo"
	"myblogx/service/ai_service/ai_overwrite"
	"myblogx/service/ai_service/ai_scoring"
)

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

type AIArticleScoringRequest struct {
	Title   string `json:"title"`
	Content string `json:"content" binding:"required"`
}

type AIArticleScoringResponse = ai_scoring.ArticleScoreResponse

type AIOverwriteRequest struct {
	Mode          string `json:"mode" binding:"required,oneof=polish grammar_fix style_transform"`
	SelectionText string `json:"selection_text" binding:"required,max=3500"`
	PrefixText    string `json:"prefix_text" binding:"max=800"`
	SuffixText    string `json:"suffix_text" binding:"max=800"`
	ArticleTitle  string `json:"article_title" binding:"required,max=50"`
	TargetStyle   string `json:"target_style" binding:"max=30"`
}

type AIDiagnoseRequest struct {
	SelectionText string `json:"selection_text" binding:"required,max=3500"`
	PrefixText    string `json:"prefix_text" binding:"max=800"`
	SuffixText    string `json:"suffix_text" binding:"max=800"`
	ArticleTitle  string `json:"article_title" binding:"required,max=50"`
}

type AIDiagnoseResponse = ai_diagnose.DiagnoseResponse

func (r AIOverwriteRequest) toServiceRequest() ai_overwrite.RewriteRequest {
	return ai_overwrite.RewriteRequest{
		Mode:          r.Mode,
		SelectionText: r.SelectionText,
		PrefixText:    r.PrefixText,
		SuffixText:    r.SuffixText,
		ArticleTitle:  r.ArticleTitle,
		TargetStyle:   r.TargetStyle,
	}
}

func (r AIDiagnoseRequest) toServiceRequest() ai_diagnose.DiagnoseRequest {
	return ai_diagnose.DiagnoseRequest{
		SelectionText: r.SelectionText,
		PrefixText:    r.PrefixText,
		SuffixText:    r.SuffixText,
		ArticleTitle:  r.ArticleTitle,
	}
}
