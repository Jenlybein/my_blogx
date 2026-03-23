package top

import (
	"errors"
	"myblogx/common/res"
	"myblogx/global"
	"myblogx/middleware"
	"myblogx/models"
	"myblogx/models/enum"
	"myblogx/service/es_service"
	"myblogx/utils/jwts"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const maxUserTopArticleCount = 3

func (TopApi) ArticleTopSetView(c *gin.Context) {
	cr := middleware.GetBindJson[ArticleTopSetRequest](c)
	claims := jwts.MustGetClaimsByGin(c)

	var article models.ArticleModel
	if err := global.DB.Select("id", "author_id", "status").Take(&article, "id = ?", cr.ArticleID).Error; err != nil {
		res.FailWithMsg("文章不存在", c)
		return
	}

	switch cr.Type {
	case 1:
		if article.AuthorID != claims.UserID {
			res.FailWithMsg("只能置顶自己的文章", c)
			return
		}
		if article.Status != enum.ArticleStatusPublished {
			res.FailWithMsg("文章未发布，无法置顶", c)
			return
		}
	case 2:
		if !claims.IsAdmin() {
			res.FailWithMsg("只有管理员才能执行管理员置顶", c)
			return
		}
	default:
		res.FailWithMsg("置顶类型错误", c)
		return
	}

	err := global.DB.Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Take(&models.UserTopArticleModel{}, "user_id = ? AND article_id = ?", claims.UserID, article.ID).Error; err == nil {
			return errors.New("文章已被置顶")
		} else if err != gorm.ErrRecordNotFound {
			return err
		}

		// 检查用户置顶数量
		if cr.Type == 1 && !claims.IsAdmin() {
			if err := tx.Model(&models.UserTopArticleModel{}).
				Where("user_id = ?", claims.UserID).
				Count(&count).Error; err != nil {
				return err
			}
			if count >= maxUserTopArticleCount {
				return errors.New("文章置顶数量达到限制")
			}
		}

		return tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "user_id"},
				{Name: "article_id"},
			},
			DoUpdates: clause.Assignments(map[string]any{
				"deleted_at": nil,
				"updated_at": time.Now(),
			}),
		}).Create(&models.UserTopArticleModel{
			UserID:    claims.UserID,
			ArticleID: article.ID,
		}).Error
	})
	if err != nil {
		global.Logger.Errorf("文章置顶失败 article_id=%d user_id=%d type=%d err=%v", article.ID, claims.UserID, cr.Type, err)
		res.FailWithError(err, c)
		return
	}

	if err := es_service.UpdateESDocsTop([]uint{article.ID}); err != nil {
		global.Logger.Errorf("更新文章置顶后刷新 ES 失败 article_id=%d err=%v", article.ID, err)
	}

	res.OkWithMsg("文章置顶成功", c)
}
