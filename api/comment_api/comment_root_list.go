// 获取文章一级评论列表
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
)

type CommentRootListRequest struct {
	common.PageInfo
	ArticleID uint `form:"article_id" binding:"required"`
}

type CommentRootListResponse struct {
	models.CommentModel
	UserNickname string `json:"user_nickname"`
	UserAvatar   string `json:"user_avatar"`
}

func (CommentApi) CommentRootListView(c *gin.Context) {
	cr := middleware.GetBindQuery[CommentRootListRequest](c)

	var article models.ArticleModel
	if err := global.DB.Select("id").Take(&article, cr.ArticleID).Error; err != nil {
		res.FailWithMsg("文章不存在", c)
		return
	}

	// 注意：reply_id/root_id 的零值条件不能依赖结构体过滤，需要显式 Where。
	list, count, err := common.ListQuery(models.CommentModel{
		ArticleID: cr.ArticleID,
		Status:    enum.CommentStatusPublished,
	}, common.Options{
		PageInfo:     cr.PageInfo,
		DefaultOrder: "created_at desc",
		Where:        global.DB.Where("reply_id = 0 AND root_id = 0"),
		Select: []string{
			"id",
			"created_at",
			"updated_at",
			"content",
			"user_id",
			"article_id",
			"reply_id",
			"root_id",
			"digg_count",
			"reply_count",
			"status",
		},
		ExactPreloads: map[string][]string{
			"UserModel": {"id", "nickname", "avatar"},
		},
	})
	if err != nil {
		res.FailWithMsg("查询一级评论失败 "+err.Error(), c)
		return
	}

	commentIDs := make([]uint, 0, len(list))
	for _, item := range list {
		commentIDs = append(commentIDs, item.ID)
	}
	replyCountMap := redis_comment.GetBatchCacheReply(commentIDs)

	responseList := make([]CommentRootListResponse, 0, len(list))
	for _, item := range list {
		item.ReplyCount += replyCountMap[item.ID]
		responseList = append(responseList, CommentRootListResponse{
			CommentModel: item,
			UserNickname: item.UserModel.Nickname,
			UserAvatar:   item.UserModel.Avatar,
		})
	}

	res.OkWithData(map[string]any{
		"list":  responseList,
		"count": count,
	}, c)
}
