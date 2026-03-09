package message_service

type ArticleCommentMessage struct {
	CommentID uint
	Content   string

	ReceiverID   uint
	ActionUserID uint

	ArticleID    uint
	ArticleTitle string
}

type ArticleReplyMessage struct {
	CommentID uint
	Content   string

	ReceiverID   uint
	ActionUserID uint

	ArticleID    uint
	ArticleTitle string
}

type ArticleDiggMessage struct {
	ReceiverID   uint
	ActionUserID uint

	ArticleID    uint
	ArticleTitle string
}

type CommentDiggMessage struct {
	CommentID uint
	Content   string

	ReceiverID   uint
	ActionUserID uint

	ArticleID    uint
	ArticleTitle string
}

type ArticleFavorMessage struct {
	ReceiverID   uint
	ActionUserID uint

	ArticleID    uint
	ArticleTitle string
}

type SystemMessage struct {
	ReceiverID   uint
	ActionUserID *uint

	Content string

	// 额外提示的链接
	LinkTitle string `gorm:"size:128" json:"link_title"`
	LinkHerf  string `gorm:"size:256" json:"link_herf"`
}
