package comment_api

import (
	"errors"
	"myblogx/common/res"
	"myblogx/global"
	"myblogx/middleware"
	"myblogx/models"
	"myblogx/models/enum"
	"myblogx/service/redis_service/redis_comment"
	"myblogx/utils/jwts"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func (CommentApi) CommentDiggView(c *gin.Context) {
	id := middleware.GetBindUri[models.IDRequest](c)

	var comment models.CommentModel
	if err := global.DB.Take(&comment, "id = ? and status = ?", id.ID, enum.CommentStatusPublished).Error; err != nil {
		res.FailWithMsg("评论不存在", c)
		return
	}

	claims := jwts.MustGetClaimsByGin(c)
	if err := global.DB.Take(&models.CommentDiggModel{}, "comment_id = ? and user_id = ?", id.ID, claims.UserID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if err := global.DB.Create(&models.CommentDiggModel{
				CommentID: id.ID,
				UserID:    claims.UserID,
			}).Error; err != nil {
				res.FailWithMsg("点赞失败", c)
				return
			}
			if err := redis_comment.SetCacheDigg(id.ID, 1); err != nil {
				global.Logger.Errorf("写入评论点赞缓存失败 comment_id=%d err=%v", id.ID, err)
			}
			res.OkWithMsg("点赞成功", c)
			return
		}
		res.FailWithMsg("查询点赞记录失败", c)
		return
	}

	if err := global.DB.Delete(&models.CommentDiggModel{}, "comment_id = ? and user_id = ?", id.ID, claims.UserID).Error; err != nil {
		res.FailWithMsg("取消点赞失败", c)
		return
	}
	if err := redis_comment.SetCacheDigg(id.ID, -1); err != nil {
		global.Logger.Errorf("回写评论点赞缓存失败 comment_id=%d err=%v", id.ID, err)
	}
	res.OkWithMsg("取消点赞成功", c)
}
