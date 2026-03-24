package ai_service

import (
	"myblogx/models/ctype"
	"myblogx/utils/markdown"
	"time"
)

type AISearchList struct {
	ID             ctype.ID               `json:"id"`
	CreatedAt      time.Time              `json:"created_at"`
	Title          string                 `json:"title"`
	Abstract       string                 `json:"abstract,omitempty"`
	Content        string                 `json:"content,omitempty"`
	Part           []markdown.ContentPart `json:"part,omitempty"`
	ViewCount      int                    `json:"view_count"`
	DiggCount      int                    `json:"digg_count"`
	CommentCount   int                    `json:"comment_count"`
	FavorCount     int                    `json:"favor_count"`
	Tags           []string               `json:"tags"`
}
