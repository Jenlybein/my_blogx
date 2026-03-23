package article_api

import (
	"myblogx/common/res"
	"myblogx/global"
	"myblogx/middleware"
	"myblogx/models"
	"myblogx/models/enum"
	"myblogx/service/message_service"
	"myblogx/service/redis_service/redis_article"
	"myblogx/utils/jwts"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (ArticleApi) ArticleDiggView(c *gin.Context) {
	id := middleware.GetBindUri[models.IDRequest](c)

	var article models.ArticleModel
	if err := global.DB.Take(&article, "id = ? and status = ?", id.ID, enum.ArticleStatusPublished).Error; err != nil {
		res.FailWithMsg("文章不存在", c)
		return
	}

	claims := jwts.MustGetClaimsByGin(c)
	var digg models.ArticleDiggModel
	if err := global.DB.Take(&digg, "article_id = ? and user_id = ?", id.ID, claims.UserID).Error; err == nil {
		if err := global.DB.Delete(&digg).Error; err != nil {
			res.FailWithMsg("取消点赞失败", c)
			return
		}

		redis_article.SetCacheDigg(id.ID, -1)
		res.OkWithMsg("取消点赞成功", c)
		return
	} else if err != gorm.ErrRecordNotFound {
		res.FailWithMsg("查询点赞记录失败", c)
		return
	}

	if err := global.DB.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "article_id"},
			{Name: "user_id"},
		},
		DoUpdates: clause.Assignments(map[string]any{
			"deleted_at": nil,
			"updated_at": time.Now(),
		}),
	}).Create(&models.ArticleDiggModel{
		ArticleID: id.ID,
		UserID:    claims.UserID,
	}).Error; err != nil {
		res.FailWithMsg("点赞失败", c)
		return
	}

	redis_article.SetCacheDigg(id.ID, 1)
	go message_service.InsertArticleDiggMessage(message_service.ArticleDiggMessage{
		ReceiverID:   article.AuthorID,
		ActionUserID: claims.UserID,
		ArticleID:    article.ID,
		ArticleTitle: article.Title,
	})
	res.OkWithMsg("点赞成功", c)
}
