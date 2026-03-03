package comment_api

import (
	"myblogx/common"
	"myblogx/common/res"
	"myblogx/global"
	"myblogx/middleware"
	"myblogx/models"
	"myblogx/models/enum"
	"myblogx/service/redis_service/redis_comment"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CommentReplyListRequest struct {
	common.PageInfo
	ArticleID uint `form:"article_id" binding:"required"`
	RootID    uint `form:"root_id" binding:"required"`
}

type CommentReplyListResponse struct {
	models.CommentModel
	UserNickname      string `json:"user_nickname"`
	UserAvatar        string `json:"user_avatar"`
	ReplyUserNickname string `json:"reply_user_nickname"`
}

func (CommentApi) CommentReplyListView(c *gin.Context) {
	cr := middleware.GetBindQuery[CommentReplyListRequest](c)

	var root models.CommentModel
	if err := global.DB.Select("id", "article_id", "reply_id", "root_id", "reply_count").
		Take(&root, "id = ? and article_id = ? and status = ?", cr.RootID, cr.ArticleID, enum.CommentStatusPublished).Error; err != nil {
		res.FailWithMsg("一级评论不存在", c)
		return
	}
	if root.ReplyId != 0 || root.RootID != 0 {
		res.FailWithMsg("必须是一级评论", c)
		return
	}

	query := global.DB.Model(&models.CommentModel{}).Where(models.CommentModel{
		ArticleID: cr.ArticleID,
		RootID:    cr.RootID,
		Status:    enum.CommentStatusPublished,
	})

	var total int64
	if err := query.Count(&total).Error; err != nil {
		res.FailWithMsg("查询二级评论失败 "+err.Error(), c)
		return
	}
	count := int(total)

	var list []models.CommentModel
	err := query.
		Select("id", "content", "user_id", "article_id", "reply_id", "root_id", "digg_count", "reply_count", "status", "created_at", "updated_at").
		Order("created_at asc").
		Limit(cr.GetLimit()).
		Offset(cr.GetOffset(count)).
		Preload("UserModel", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "nickname", "avatar")
		}).
		Preload("ParentModel", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "user_id")
		}).
		Preload("ParentModel.UserModel", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "nickname")
		}).
		Find(&list).Error
	if err != nil {
		res.FailWithMsg("查询二级评论失败 "+err.Error(), c)
		return
	}

	commentIDs := make([]uint, 0, len(list))
	for _, item := range list {
		commentIDs = append(commentIDs, item.ID)
	}
	replyCountMap := redis_comment.GetBatchCacheReply(commentIDs)

	responseList := make([]CommentReplyListResponse, 0, len(list))
	for _, item := range list {
		item.ReplyCount += replyCountMap[item.ID]
		resp := CommentReplyListResponse{
			CommentModel: item,
			UserNickname: item.UserModel.Nickname,
			UserAvatar:   item.UserModel.Avatar,
		}
		if item.ParentModel != nil {
			resp.ReplyUserNickname = item.ParentModel.UserModel.Nickname
		}
		responseList = append(responseList, resp)
	}

	rootReplyCount := root.ReplyCount + redis_comment.GetCacheReply(cr.RootID)
	res.OkWithData(map[string]any{
		"root_id":     cr.RootID,
		"reply_count": rootReplyCount,
		"list":        responseList,
		"count":       count,
	}, c)
}
