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

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const maxUserTopArticleCount = 3

var (
	errArticleTopExists = errors.New("article top already exists")
	errArticleTopLimit  = errors.New("article top limit exceeded")
)

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
		if err := tx.Model(&models.UserTopArticleModel{}).
			Where("user_id = ? AND article_id = ?", claims.UserID, article.ID).
			Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return errArticleTopExists
		}

		if cr.Type == 1 && !claims.IsAdmin() {
			if err := tx.Model(&models.UserTopArticleModel{}).
				Where("user_id = ?", claims.UserID).
				Count(&count).Error; err != nil {
				return err
			}
			if count >= maxUserTopArticleCount {
				return errArticleTopLimit
			}
		}

		return tx.Create(&models.UserTopArticleModel{
			UserID:    claims.UserID,
			ArticleID: article.ID,
		}).Error
	})
	if err != nil {
		switch {
		case errors.Is(err, errArticleTopExists):
			res.FailWithMsg("文章已置顶", c)
		case errors.Is(err, errArticleTopLimit):
			res.FailWithMsg("每个用户最多置顶 3 篇自己的文章", c)
		default:
			global.Logger.Errorf("文章置顶失败 article_id=%d user_id=%d type=%d err=%v", article.ID, claims.UserID, cr.Type, err)
			res.FailWithMsg("文章置顶失败", c)
		}
		return
	}

	if err := es_service.UpdateESDocsTop([]uint{article.ID}); err != nil {
		global.Logger.Errorf("更新文章置顶后刷新 ES 失败 article_id=%d err=%v", article.ID, err)
	}

	res.OkWithMsg("文章置顶成功", c)
}
