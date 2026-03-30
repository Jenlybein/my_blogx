// 获取评论管理列表
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
	"myblogx/service/follow_service"
	"myblogx/service/redis_service/redis_comment"
	"myblogx/utils/jwts"

	"github.com/gin-gonic/gin"
)

type CommentManListRequest struct {
	common.PageInfo
	ArticleID ctype.ID           `form:"article_id"`
	UserID    ctype.ID           `form:"user_id"`
	Status    enum.CommentStatus `form:"status" binding:"oneof=1 2 3"`
	Type      int8               `form:"type" binding:"required,oneof=1 2 3"`
	// 1 查我文章下的评论 2 查我发的评论 3 管理员查所有评论
}

type CommentManListResponse struct {
	ID           ctype.ID `json:"id"`
	CreatedAt    string   `json:"created_at"`
	Content      string   `json:"content"`
	DiggCount    int      `json:"digg_count"`
	ReplyCount   int      `json:"reply_count"`
	UserID       ctype.ID `json:"user_id"`
	UserNickname string   `json:"user_nickname"`
	UserAvatar   string   `json:"user_avatar"`
	Relation     int8     `json:"relation,omitempty"`
	ArticleID    ctype.ID `json:"article_id"`
	ArticleTitle string   `json:"article_title"`
	ArticleCover string   `json:"article_cover"`
}

func (CommentApi) CommentManListView(c *gin.Context) {
	cr := middleware.GetBindQuery[CommentManListRequest](c)
	claims := jwts.MustGetClaimsByGin(c)

	query := global.DB.Where("")

	switch cr.Type {
	case 1: // 查我文章下的评论
		articleQuery := global.DB.Model(&models.ArticleModel{}).
			Select("id").
			Where("author_id = ?", claims.UserID)
		query = query.Where("article_id IN (?)", articleQuery)
		cr.Status = enum.CommentStatusPublished
	case 2: // 查我发的评论
		cr.UserID = claims.UserID
	case 3: // 管理员查所有评论
		if !claims.IsAdmin() {
			res.FailWithMsg("权限错误", c)
			return
		}
	}

	if cr.ArticleID != 0 {
		query = query.Where("article_id = ?", cr.ArticleID)
	}
	if cr.UserID != 0 {
		query = query.Where("user_id = ?", cr.UserID)
	}

	commentList, count, err := common.ListQuery(models.CommentModel{
		Status: cr.Status,
	}, common.Options{
		PageInfo: cr.PageInfo,
		Likes:    []string{"content"},
		Where:    query,
		Select: []string{
			"id",
			"created_at",
			"content",
			"digg_count",
			"reply_count",
			"user_id",
			"article_id",
			"status",
		},
		DefaultOrder: "created_at desc",
		ExactPreloads: map[string][]string{
			"ArticleModel": {"id", "title", "cover"},
			"UserModel":    {"id", "nickname", "avatar"},
		},
	})
	if err != nil {
		res.FailWithMsg("查询评论失败 "+err.Error(), c)
		return
	}

	commentIDs := make([]ctype.ID, 0, len(commentList))
	for _, item := range commentList {
		commentIDs = append(commentIDs, item.ID)
	}
	replyCountMap := redis_comment.GetBatchCacheReply(commentIDs)
	diggCountMap := redis_comment.GetBatchCacheDigg(commentIDs)

	relationMap := make(map[ctype.ID]relationship_enum.Relation)
	if cr.Type == 1 {
		userIDs := make([]ctype.ID, 0, len(commentList))
		seen := make(map[ctype.ID]struct{}, len(commentList))
		for _, item := range commentList {
			if _, ok := seen[item.UserID]; ok {
				continue
			}
			seen[item.UserID] = struct{}{}
			userIDs = append(userIDs, item.UserID)
		}
		relationMap = follow_service.CalUserRelationshipBatch(claims.UserID, userIDs)
	}

	list := make([]CommentManListResponse, 0, len(commentList))
	for _, item := range commentList {
		item.ReplyCount += replyCountMap[item.ID]
		item.DiggCount += diggCountMap[item.ID]
		list = append(list, CommentManListResponse{
			ID:           item.ID,
			CreatedAt:    item.CreatedAt.Format("2006-01-02 15:04:05"),
			Content:      item.Content,
			DiggCount:    item.DiggCount,
			ReplyCount:   item.ReplyCount,
			UserID:       item.UserID,
			UserNickname: item.UserModel.Nickname,
			UserAvatar:   item.UserModel.Avatar,
			Relation:     int8(relationMap[item.UserID]),
			ArticleID:    item.ArticleID,
			ArticleTitle: item.ArticleModel.Title,
			ArticleCover: item.ArticleModel.Cover,
		})
	}

	res.OkWithList(list, count, c)
}
