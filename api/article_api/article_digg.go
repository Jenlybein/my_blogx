package article_api

import (
	"errors"
	"myblogx/common/res"
	"myblogx/global"
	"myblogx/middleware"
	"myblogx/models"
	"myblogx/models/enum"
	"myblogx/service/redis_service/redis_article"
	"myblogx/utils/jwts"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func (ArticleApi) ArticleDiggView(c *gin.Context) {
	id := middleware.GetBindUri[models.IDRequest](c)

	var article models.ArticleModel
	if err := global.DB.Take(&article, "id = ? and status = ?", id.ID, enum.ArticleStatusPublished).Error; err != nil {
		res.FailWithMsg("文章不存在", c)
		return
	}

	claims := jwts.MustGetClaimsByGin(c)
	if err := global.DB.Take(&models.ArticleDiggModel{}, "article_id = ? and user_id = ?", id.ID, claims.UserID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if err := global.DB.Create(&models.ArticleDiggModel{
				ArticleID: id.ID,
				UserID:    claims.UserID,
			}).Error; err != nil {
				res.FailWithMsg("点赞失败", c)
				return
			}
			redis_article.SetCacheDigg(id.ID, 1)
			res.OkWithMsg("点赞成功", c)
			return
		}
		res.FailWithMsg("查询点赞记录失败", c)
		return
	} else {
		// 如果已点赞
		if err := global.DB.Delete(&models.ArticleDiggModel{}, "article_id = ? and user_id = ?", id.ID, claims.UserID).Error; err != nil {
			res.FailWithMsg("取消点赞失败", c)
			return
		}
		redis_article.SetCacheDigg(id.ID, -1)
		res.OkWithMsg("取消点赞成功", c)
		return
	}
}
