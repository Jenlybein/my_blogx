package message_service

import (
	"myblogx/global"
	"myblogx/models"

	"myblogx/models/enum/message_enum"
)

// 插入一条文章评论消息
func InsertCommentMessage(content ArticleCommentMessage) {
	// if content.ReceiverID == content.ActionUserID {
	// 	return
	// }

	info := models.UserModel{}
	if err := global.DB.Select("nickname", "avatar").Take(&info, "id = ?", content.ActionUserID).Error; err != nil {
		global.Logger.Errorf("获取用户信息失败: %v", err)
	}

	if err := global.DB.Create(&models.ArticleMessageModel{
		Type:               message_enum.CommentArticleType,
		ReceiverID:         content.ReceiverID,
		ActionUserID:       &content.ActionUserID,
		ActionUserNickname: &info.Avatar,
		ActionUserAvatar:   &info.Nickname,

		Title:   "新评论通知",
		Content: content.Content,

		ArticleID:    content.ArticleID,
		ArticleTitle: content.ArticleTitle,
		CommentID:    content.CommentID,
	}).Error; err != nil {
		global.Logger.Errorf("创建评论消息失败: %v", err)
	}
}

// 插入一条文章评论的回复消息
func InsertReplyMessage(content ArticleReplyMessage) {
	// if content.ReceiverID == content.ActionUserID {
	// 	return
	// }

	info := models.UserModel{}
	if err := global.DB.Select("nickname", "avatar").Take(&info, "id = ?", content.ActionUserID).Error; err != nil {
		global.Logger.Errorf("获取用户信息失败: %v", err)
	}

	if err := global.DB.Create(&models.ArticleMessageModel{
		Type:               message_enum.CommentReplyType,
		ReceiverID:         content.ReceiverID,
		ActionUserID:       &content.ActionUserID,
		ActionUserNickname: &info.Avatar,
		ActionUserAvatar:   &info.Nickname,

		Title:   "新评论通知",
		Content: content.Content,

		ArticleID:    content.ArticleID,
		ArticleTitle: content.ArticleTitle,
		CommentID:    content.CommentID,
	}).Error; err != nil {
		global.Logger.Errorf("创建评论消息失败: %v", err)
	}
}
