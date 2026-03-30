// 获取文章一级评论列表
package comment_api

import (
	"myblogx/common"
	"myblogx/common/res"
	"myblogx/global"
	"myblogx/middleware"
	"myblogx/models"
	"myblogx/models/ctype"
	"myblogx/models/enum"
	"myblogx/models/enum/relationship_enum"
	"myblogx/service/redis_service/redis_comment"
	"time"

	"github.com/gin-gonic/gin"
)

type CommentRootListRequest struct {
	common.PageInfo
	ArticleID ctype.ID `form:"article_id" binding:"required"`
}

type CommentRootListResponse struct {
	ID           ctype.ID           `json:"id"`
	CreatedAt    time.Time          `json:"created_at"`
	Content      string             `json:"content"`
	UserID       ctype.ID           `json:"user_id"`
	ReplyId      ctype.ID           `json:"reply_id"`
	RootID       ctype.ID           `json:"root_id"`
	DiggCount    int                `json:"digg_count"`
	ReplyCount   int                `json:"reply_count"`
	IsDigg       bool               `json:"is_digg"`
	Relation     int8               `json:"relation"`
	Status       enum.CommentStatus `json:"status"`
	UserNickname string             `json:"user_nickname"`
	UserAvatar   string             `json:"user_avatar"`
}

func (CommentApi) CommentRootListView(c *gin.Context) {
	cr := middleware.GetBindQuery[CommentRootListRequest](c)

	var article models.ArticleModel
	if err := global.DB.Select("id").Take(&article, cr.ArticleID).Error; err != nil {
		res.FailWithMsg("文章不存在", c)
		return
	}

	// reply_id/root_id 的零值条件不能依赖结构体过滤，需要显式 Where。
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

	// 查询缓存内存的点赞、回复数增量
	commentIDs := make([]ctype.ID, 0, len(list))
	for _, item := range list {
		commentIDs = append(commentIDs, item.ID)
	}
	replyCountMap := map[ctype.ID]int{}
	diggCountMap := map[ctype.ID]int{}
	if len(commentIDs) > 0 {
		replyCountMap = redis_comment.GetBatchCacheReply(commentIDs)
		diggCountMap = redis_comment.GetBatchCacheDigg(commentIDs)
	}

	// 批量查询点赞，好友关系
	viewerUserID := commentViewerIDFromGin(c)
	userIDs := make([]ctype.ID, 0, len(list))
	for _, item := range list {
		userIDs = append(userIDs, item.UserID)
	}
	isDiggMap := buildCommentDiggMap(viewerUserID, commentIDs)
	relationMap := buildCommentRelationMap(viewerUserID, userIDs)

	// 组装响应
	responseList := make([]CommentRootListResponse, 0, len(list))
	for _, item := range list {
		item.ReplyCount += replyCountMap[item.ID]
		item.DiggCount += diggCountMap[item.ID]
		relation := relationship_enum.RelationStranger
		if got, ok := relationMap[item.UserID]; ok {
			relation = got
		}
		responseList = append(responseList, CommentRootListResponse{
			ID:           item.ID,
			CreatedAt:    item.CreatedAt,
			Content:      item.Content,
			UserID:       item.UserID,
			ReplyId:      item.ReplyId,
			RootID:       item.RootID,
			DiggCount:    item.DiggCount,
			ReplyCount:   item.ReplyCount,
			IsDigg:       isDiggMap[item.ID],
			Relation:     int8(relation),
			Status:       item.Status,
			UserNickname: item.UserModel.Nickname,
			UserAvatar:   item.UserModel.Avatar,
		})
	}

	res.OkWithData(map[string]any{
		"list":  responseList,
		"count": count,
	}, c)
}
