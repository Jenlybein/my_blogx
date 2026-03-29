package message_service

import (
	"myblogx/global"
	"myblogx/models"
	"myblogx/models/ctype"

	"myblogx/models/enum/message_enum"
)

// 插入一条文章评论消息
func InsertCommentMessage(content ArticleCommentMessage) {
	// if content.ReceiverID == content.ActionUserID {
	// 	return
	// }

	nickname, avatar := getActionUserInfo(content.ActionUserID)

	if err := global.DB.Create(&models.ArticleMessageModel{
		Type:               message_enum.CommentArticleType,
		ReceiverID:         content.ReceiverID,
		ActionUserID:       &content.ActionUserID,
		ActionUserNickname: &nickname,
		ActionUserAvatar:   &avatar,

		Content: content.Content,

		ArticleID:    content.ArticleID,
		ArticleTitle: content.ArticleTitle,
		CommentID:    content.CommentID,
	}).Error; err != nil {
		global.Logger.Errorf("创建评论消息失败: %v", err)
		return
	}
}

// 插入一条文章评论的回复消息
func InsertReplyMessage(content ArticleReplyMessage) {
	// if content.ReceiverID == content.ActionUserID {
	// 	return
	// }

	nickname, avatar := getActionUserInfo(content.ActionUserID)

	if err := global.DB.Create(&models.ArticleMessageModel{
		Type:               message_enum.CommentReplyType,
		ReceiverID:         content.ReceiverID,
		ActionUserID:       &content.ActionUserID,
		ActionUserNickname: &nickname,
		ActionUserAvatar:   &avatar,

		Content: content.Content,

		ArticleID:    content.ArticleID,
		ArticleTitle: content.ArticleTitle,
		CommentID:    content.CommentID,
	}).Error; err != nil {
		global.Logger.Errorf("创建回复消息失败: %v", err)
		return
	}
}

// 插入一条文章点赞消息
func InsertArticleDiggMessage(content ArticleDiggMessage) {
	// if content.ReceiverID == content.ActionUserID {
	// 	return
	// }

	if err := global.DB.Take(&models.ArticleMessageModel{}, "action_user_id = ? and type = ? and article_id = ?", content.ActionUserID, message_enum.DiggArticleType, content.ArticleID).Error; err == nil {
		return
	}

	nickname, avatar := getActionUserInfo(content.ActionUserID)

	if err := global.DB.Create(&models.ArticleMessageModel{
		Type:               message_enum.DiggArticleType,
		ReceiverID:         content.ReceiverID,
		ActionUserID:       &content.ActionUserID,
		ActionUserNickname: &nickname,
		ActionUserAvatar:   &avatar,
		ArticleID:          content.ArticleID,
		ArticleTitle:       content.ArticleTitle,
	}).Error; err != nil {
		global.Logger.Errorf("创建文章点赞消息失败: %v", err)
		return
	}
}

// 插入一条评论点赞消息
func InsertCommentDiggMessage(content CommentDiggMessage) {
	// if content.ReceiverID == content.ActionUserID {
	// 	return
	// }

	if err := global.DB.Take(&models.ArticleMessageModel{}, "action_user_id = ? and type = ? and comment_id = ?", content.ActionUserID, message_enum.DiggCommentType, content.CommentID).Error; err == nil {
		return
	}

	nickname, avatar := getActionUserInfo(content.ActionUserID)

	if err := global.DB.Create(&models.ArticleMessageModel{
		Type:               message_enum.DiggCommentType,
		CommentID:          content.CommentID,
		Content:            content.Content,
		ReceiverID:         content.ReceiverID,
		ActionUserID:       &content.ActionUserID,
		ActionUserNickname: &nickname,
		ActionUserAvatar:   &avatar,
		ArticleID:          content.ArticleID,
		ArticleTitle:       content.ArticleTitle,
	}).Error; err != nil {
		global.Logger.Errorf("创建评论点赞消息失败: %v", err)
		return
	}
}

// 插入一条文章收藏消息
func InsertArticleFavorMessage(content ArticleFavorMessage) {
	// if content.ReceiverID == content.ActionUserID {
	// 	return
	// }

	if err := global.DB.Take(&models.ArticleMessageModel{}, "action_user_id = ? and type = ? and article_id = ?", content.ActionUserID, message_enum.FavorArticleType, content.ArticleID).Error; err == nil {
		return
	}

	nickname, avatar := getActionUserInfo(content.ActionUserID)

	if err := global.DB.Create(&models.ArticleMessageModel{
		Type:               message_enum.FavorArticleType,
		ReceiverID:         content.ReceiverID,
		ActionUserID:       &content.ActionUserID,
		ActionUserNickname: &nickname,
		ActionUserAvatar:   &avatar,
		ArticleID:          content.ArticleID,
		ArticleTitle:       content.ArticleTitle,
	}).Error; err != nil {
		global.Logger.Errorf("创建评论收藏消息失败: %v", err)
		return
	}
}

// 插入一条系统消息
func InsertSystemMessage(content SystemMessage) {
	if content.ReceiverID == 0 {
		global.Logger.Errorf("创建系统消息失败: 接收者ID不能为空")
		return
	}

	msg := models.ArticleMessageModel{
		Type:       message_enum.SystemType,
		ReceiverID: content.ReceiverID,
		Content:    content.Content,
		LinkTitle:  content.LinkTitle,
		LinkHerf:   content.LinkHerf,
	}

	if content.ActionUserID != nil {
		msg.ActionUserID = content.ActionUserID
		nickname, avatar := getActionUserInfo(*content.ActionUserID)
		msg.ActionUserNickname = &nickname
		msg.ActionUserAvatar = &avatar
	}

	if err := global.DB.Create(&msg).Error; err != nil {
		global.Logger.Errorf("创建系统消息失败: %v", err)
		return
	}
}

func getActionUserInfo(actionUserID ctype.ID) (nickname string, avatar string) {
	info := models.UserModel{}
	if err := global.DB.Select("nickname", "avatar").Take(&info, "id = ?", actionUserID).Error; err != nil {
		global.Logger.Errorf("获取用户信息失败: %v", err)
		return
	}

	return info.Nickname, info.Avatar
}
