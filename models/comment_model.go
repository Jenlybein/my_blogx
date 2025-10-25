package models

// 评论表
type CommentModel struct {
	Model
	Content        string          `json:"content"`
	UserID         uint            `json:"user_id"`
	UserModel      UserModel       `json:"user_model" gorm:"foreignKey:UserID;references:ID"`
	ArticleID      uint            `json:"article_id"`
	ArticleModel   ArticleModel    `json:"article_model" gorm:"foreignKey:ArticleID;references:ID"`
	ParentID       *uint           `json:"parent_id"` // 父评论ID，0表示一级评论
	ParentModel    *CommentModel   `json:"parent_model" gorm:"foreignKey:ParentID;references:ID"`
	SubCommentList []*CommentModel `json:"sub_comment_list" gorm:"foreignKey:RootParentID;references:ID"` // 子评论列表
	RootParentID   *uint           `json:"root_parent_id"`                                                // 根评论ID，0表示不是回复评论
	DiggCount      int             `json:"digg_count"`                                                    // 点赞数
}
