package article_api

import (
	"errors"
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
		if err = tx.Take(&articleFavorite, "article_id = ? and user_id = ? and favor_id = ?", cr.ArticleID, claims.UserID, favorite.ID).Error; err == nil {
			if err = tx.Delete(&articleFavorite).Error; err != nil {
				return err
			}

			isFavorited = false
			return nil
		}
		if err != gorm.ErrRecordNotFound {
			return err
		}

		if err = tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "article_id"},
				{Name: "user_id"},
				{Name: "favor_id"},
			},
			DoUpdates: clause.Assignments(map[string]any{
				"deleted_at": nil,
				"updated_at": time.Now(),
			}),
		}).Create(&models.UserArticleFavorModel{
			ArticleID: cr.ArticleID,
			UserID:    claims.UserID,
			FavorID:   favorite.ID,
		}).Error; err != nil {
			return err
		}

		isFavorited = true
		return nil
	}); err != nil {
		res.FailWithMsg("收藏操作失败", c)
		return
	}

	if isFavorited {
		go message_service.InsertArticleFavorMessage(message_service.ArticleFavorMessage{
			ReceiverID:   article.AuthorID,
			ActionUserID: claims.UserID,
			ArticleID:    cr.ArticleID,
			ArticleTitle: article.Title,
		})

		if err := redis_article.SetCacheFavorite(cr.ArticleID, 1); err != nil {
			global.Logger.Errorf("文章收藏数据加一失败 err: %v", err)
		}
		res.OkWithMsg("收藏成功", c)
	} else {
		if err := redis_article.SetCacheFavorite(cr.ArticleID, -1); err != nil {
			global.Logger.Errorf("文章收藏数据减一失败 err: %v", err)
		}
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
					UserID:    userID,
					Title:     "默认收藏夹",
					IsDefault: true,
				}
				if err := db.Clauses(clause.OnConflict{
					Columns: []clause.Column{
						{Name: "user_id"},
						{Name: "title"},
					},
					DoUpdates: clause.Assignments(map[string]any{
						"is_default": true,
						"deleted_at": nil,
						"updated_at": time.Now(),
					}),
				}).Create(&favorite).Error; err != nil {
					return nil, errors.New("创建默认收藏夹失败")
				}
				if err := db.Take(&favorite, "is_default = ? and user_id = ?", true, userID).Error; err != nil {
					return nil, errors.New("查询默认收藏夹失败")
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
