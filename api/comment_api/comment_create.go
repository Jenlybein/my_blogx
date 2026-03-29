package comment_api

import (
	"errors"
	"myblogx/common/res"
	"myblogx/global"
	"myblogx/middleware"
	"myblogx/models"
	"myblogx/models/ctype"
	"myblogx/models/enum"
	"myblogx/service/message_service"
	"myblogx/service/redis_service/redis_article"
	"myblogx/service/redis_service/redis_comment"
	"myblogx/utils/jwts"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CommentCreateRequest struct {
	Content   string    `json:"content" binding:"required"`
	ArticleID ctype.ID  `json:"article_id" binding:"required"`
	ReplyId   *ctype.ID `json:"reply_id"`
}

func (CommentApi) CommentCreateView(c *gin.Context) {
	cr := middleware.GetBindJson[CommentCreateRequest](c)

	var article models.ArticleModel
	if err := global.DB.Take(&article, cr.ArticleID).Error; err != nil {
		res.FailWithMsg("文章不存在", c)
		return
	}
	if !article.CommentsToggle {
		res.FailWithMsg("该文章已关闭评论", c)
		return
	}

	claims := jwts.MustGetClaimsByGin(c)

	status := enum.CommentStatusExamining
	model := models.CommentModel{
		Content:   cr.Content,
		UserID:    claims.UserID,
		ArticleID: cr.ArticleID,
		Status:    status,
	}
	var rootCommentID ctype.ID

	// 只做两级评论：回复二级评论时，仍挂到同一个一级评论下
	var replyComment models.CommentModel
	if cr.ReplyId != nil {
		if err := global.DB.Take(&replyComment, "id = ? and article_id = ? and status = ?", *cr.ReplyId, cr.ArticleID, enum.CommentStatusPublished).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				res.FailWithMsg("回复的评论不存在", c)
				return
			}
			res.FailWithMsg("查询回复评论失败", c)
			return
		}

		model.ReplyId = *cr.ReplyId
		if replyComment.RootID != 0 {
			model.RootID = replyComment.RootID
		} else {
			model.RootID = replyComment.ID
		}
		rootCommentID = model.RootID
	}

	if err := global.DB.Create(&model).Error; err != nil {
		res.FailWithMsg("评论失败", c)
		return
	}

	// 临时审核（模拟审核过程，之后再修改）
	if (claims != nil && claims.IsAdmin()) || global.Config.Site.Comment.SkipExamining {
		status = enum.CommentStatusPublished
		if err := global.DB.Model(&model).Update("status", status).Error; err != nil {
			res.FailWithMsg("审核失败", c)
			return
		}

		// 给文章创作者发送系统通知
		go message_service.InsertCommentMessage(message_service.ArticleCommentMessage{
			CommentID:    model.ID,
			Content:      cr.Content,
			ReceiverID:   article.AuthorID,
			ActionUserID: claims.UserID,
			ArticleID:    article.ID,
			ArticleTitle: article.Title,
		})

		// 给回复人发送系统通知
		if rootCommentID != 0 {
			go message_service.InsertReplyMessage(message_service.ArticleReplyMessage{
				CommentID:    model.ID,
				Content:      cr.Content,
				ReceiverID:   replyComment.UserID,
				ActionUserID: claims.UserID,
				ArticleID:    article.ID,
				ArticleTitle: article.Title,
			})
		}

		// 只有已发布评论才计入前台计数
		if err := redis_article.SetCacheComment(cr.ArticleID, 1); err != nil {
			global.Logger.Errorf("写入评论计数缓存失败: 文章ID=%d 错误=%v", cr.ArticleID, err)
		}
		if rootCommentID != 0 {
			if err := redis_comment.SetCacheReply(rootCommentID, 1); err != nil {
				global.Logger.Errorf("写入回复数缓存失败: 根评论ID=%d 错误=%v", rootCommentID, err)
			}
		}
		res.OkWithMsg("评论成功", c)
		return
	} else {
		// 进入审核
	}

	res.OkWithMsg("评论已提交，等待审核", c)
}
