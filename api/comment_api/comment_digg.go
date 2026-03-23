package comment_api

import (
	"myblogx/common/res"
	"myblogx/global"
	"myblogx/middleware"
	"myblogx/models"
	"myblogx/models/enum"
	"myblogx/service/message_service"
	"myblogx/service/redis_service/redis_comment"
	"myblogx/utils/jwts"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (CommentApi) CommentDiggView(c *gin.Context) {
	id := middleware.GetBindUri[models.IDRequest](c)

	var comment models.CommentModel
	if err := global.DB.Preload("ArticleModel", func(db *gorm.DB) *gorm.DB { return db.Select("id", "title") }).Take(&comment, "id = ? and status = ?", id.ID, enum.CommentStatusPublished).Error; err != nil {
		res.FailWithMsg("评论不存在", c)
		return
	}

	claims := jwts.MustGetClaimsByGin(c)
	var digg models.CommentDiggModel
	if err := global.DB.Take(&digg, "comment_id = ? and user_id = ?", id.ID, claims.UserID).Error; err == nil {
		if err := global.DB.Delete(&digg).Error; err != nil {
			res.FailWithMsg("取消点赞失败", c)
			return
		}
		if err := redis_comment.SetCacheDigg(id.ID, -1); err != nil {
			global.Logger.Errorf("回写评论点赞缓存失败 comment_id=%d err=%v", id.ID, err)
		}
		res.OkWithMsg("取消点赞成功", c)
		return
	} else if err != gorm.ErrRecordNotFound {
		res.FailWithMsg("查询点赞记录失败", c)
		return
	}

	if err := global.DB.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "comment_id"},
			{Name: "user_id"},
		},
		DoUpdates: clause.Assignments(map[string]any{
			"deleted_at": nil,
			"updated_at": time.Now(),
		}),
	}).Create(&models.CommentDiggModel{
		CommentID: id.ID,
		UserID:    claims.UserID,
	}).Error; err != nil {
		res.FailWithMsg("点赞失败", c)
		return
	}
	if err := redis_comment.SetCacheDigg(id.ID, 1); err != nil {
		global.Logger.Errorf("写入评论点赞缓存失败 comment_id=%d err=%v", id.ID, err)
	}
	go message_service.InsertCommentDiggMessage(message_service.CommentDiggMessage{
		ReceiverID:   comment.UserID,
		ActionUserID: claims.UserID,
		CommentID:    comment.ID,
		Content:      comment.Content,
		ArticleID:    comment.ArticleID,
		ArticleTitle: comment.ArticleModel.Title,
	})
	res.OkWithMsg("点赞成功", c)
}
