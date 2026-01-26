package article_api

import (
	"errors"
	"myblogx/common/res"
	"myblogx/global"
	"myblogx/middleware"
	"myblogx/models"
	"myblogx/models/enum"
	"myblogx/utils/jwts"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ArticleFavoriteRequest struct {
	ArticleID uint `json:"article_id" binding:"required"`
	FavorID   uint `json:"favor_id"`
}

func (ArticleApi) ArticleFavoriteSaveView(c *gin.Context) {
	cr := middleware.GetBindJson[ArticleFavoriteRequest](c)
	claims := jwts.MustGetClaimsByGin(c)

	var article models.ArticleModel
	if err := global.DB.Take(&article, "id = ? and status = ?", cr.ArticleID, enum.ArticleStatusPublished).Error; err != nil {
		res.FailWithMsg("查询文章失败", c)
		return
	}

	var isFavorited bool
	if err := global.DB.Transaction(func(tx *gorm.DB) error {
		favorite, err := getOrCreateFavoriteID(tx, cr.FavorID, claims.UserID)
		if err != nil {
			return err
		}

		var articleFavorite models.UserArticleFavorModel
		if err = tx.Take(&articleFavorite, "article_id = ? and user_id = ? and favor_id = ?", cr.ArticleID, claims.UserID, favorite.ID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				articleFavorite = models.UserArticleFavorModel{
					ArticleID: cr.ArticleID,
					UserID:    claims.UserID,
					FavorID:   favorite.ID,
				}
				if err = tx.Create(&articleFavorite).Error; err != nil {
					return err
				}
				if err = tx.Model(&favorite).Update("article_count", gorm.Expr("article_count + 1")).Error; err != nil {
					return err
				}
				isFavorited = true
				return nil
			}
			return err
		}

		if err = tx.Delete(&articleFavorite).Error; err != nil {
			return err
		}
		if err = tx.Model(&favorite).Where("article_count > 0").Update("article_count", gorm.Expr("article_count - 1")).Error; err != nil {
			return err
		}
		isFavorited = false
		return nil
	}); err != nil {
		res.FailWithMsg("收藏操作失败", c)
		return
	}

	if isFavorited {
		res.OkWithMsg("收藏成功", c)
	} else {
		res.OkWithMsg("取消收藏成功", c)
	}
}

func getOrCreateFavoriteID(db *gorm.DB, favorID, userID uint) (*models.FavoriteModel, error) {
	var favorite models.FavoriteModel
	if favorID == 0 {
		err := db.Take(&favorite, "is_default = ? and user_id = ?", true, userID).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				favorite = models.FavoriteModel{
					Title:     "默认收藏夹",
					IsDefault: true,
					UserID:    userID,
				}
				if err := db.Create(&favorite).Error; err != nil {
					return nil, errors.New("创建默认收藏夹失败")
				}
				return &favorite, nil
			}
			return nil, errors.New("查询默认收藏夹失败")
		}
		return &favorite, nil
	}

	if err := db.Take(&favorite, "id = ? and user_id = ?", favorID, userID).Error; err != nil {
		return nil, errors.New("收藏夹不存在")
	}
	return &favorite, nil
}
