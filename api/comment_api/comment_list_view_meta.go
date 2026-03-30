package comment_api

import (
	"myblogx/global"
	"myblogx/models"
	"myblogx/models/ctype"
	"myblogx/models/enum/relationship_enum"
	"myblogx/service/follow_service"
	"myblogx/service/user_service"

	"github.com/gin-gonic/gin"
)

// commentViewerIDFromGin 提取当前可选登录用户。
// 评论列表接口允许匿名访问，所以这里不能强制要求鉴权成功。
func commentViewerIDFromGin(c *gin.Context) ctype.ID {
	if authResult := user_service.MustAuthenticateAccessTokenByGin(c); authResult != nil {
		return authResult.Claims.UserID
	}
	return 0
}

// buildCommentDiggMap 生成评论点赞 map
func buildCommentDiggMap(viewerUserID ctype.ID, commentIDs []ctype.ID) map[ctype.ID]bool {
	result := make(map[ctype.ID]bool, len(commentIDs))
	if viewerUserID == 0 || len(commentIDs) == 0 {
		return result
	}

	var diggList []models.CommentDiggModel
	if err := global.DB.Select("comment_id").
		Where("user_id = ? AND comment_id IN ?", viewerUserID, commentIDs).
		Find(&diggList).Error; err != nil {
		return result
	}

	for _, item := range diggList {
		result[item.CommentID] = true
	}
	return result
}

// buildCommentRelationMap 生成评论点赞 map
func buildCommentRelationMap(viewerUserID ctype.ID, userIDs []ctype.ID) map[ctype.ID]relationship_enum.Relation {
	if viewerUserID == 0 {
		return make(map[ctype.ID]relationship_enum.Relation, len(userIDs))
	}
	return follow_service.CalUserRelationshipBatch(viewerUserID, userIDs)
}
