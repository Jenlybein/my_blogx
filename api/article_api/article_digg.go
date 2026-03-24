package article_api

import (
	"myblogx/common/res"
	"myblogx/global"
	"myblogx/middleware"
	"myblogx/models"
	"myblogx/models/enum"
	dbservice "myblogx/service/db_service"
	"myblogx/service/message_service"
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
	var digg models.ArticleDiggModel
	if err := global.DB.Take(&digg, "article_id = ? and user_id = ?", id.ID, claims.UserID).Error; err == nil {
		// 取消点赞要看条件删除是否真的命中了活记录，避免并发下双成功。
		// 取消点赞必须看本次 Delete 是否真的删掉了活记录，避免并发下双成功。
		deleteResult := global.DB.Where(map[string]any{
			"article_id": id.ID,
			"user_id":    claims.UserID,
		}).Delete(&models.ArticleDiggModel{})
		if deleteResult.Error != nil {
			res.FailWithMsg("取消点赞失败", c)
			return
		}
		if deleteResult.RowsAffected == 0 {
			res.FailWithMsg("点赞状态已变化，请刷新后重试", c)
			return
		}

		redis_article.SetCacheDigg(id.ID, -1)
		res.OkWithMsg("取消点赞成功", c)
		return
	} else if err != gorm.ErrRecordNotFound {
		res.FailWithMsg("查询点赞记录失败", c)
		return
	}

	// 点赞成功与否只看本次恢复/新建是否真的写入，不能再依赖前置查询快照。
	createdOrRestored, err := dbservice.RestoreOrCreateUnique(global.DB, &models.ArticleDiggModel{
		ArticleID: id.ID,
		UserID:    claims.UserID,
	}, []string{"article_id", "user_id"})
	if err != nil {
		res.FailWithMsg("点赞失败", c)
		return
	}
	if !createdOrRestored {
		res.FailWithMsg("请勿重复点赞", c)
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
