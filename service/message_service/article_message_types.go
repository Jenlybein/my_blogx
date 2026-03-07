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

