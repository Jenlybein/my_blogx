package article_api

import (
	"errors"
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
			// 取消收藏必须看本次 Delete 是否真的删掉了活记录，避免并发下双成功。
			deleteResult := tx.Where(map[string]any{
				"article_id": cr.ArticleID,
				"user_id":    claims.UserID,
				"favor_id":   favorite.ID,
			}).Delete(&models.UserArticleFavorModel{})
			if deleteResult.Error != nil {
				return deleteResult.Error
			}
			if deleteResult.RowsAffected == 0 {
				return errors.New("收藏状态已变化，请刷新后重试")
			}

			isFavorited = false
			return nil
		}
		if err != gorm.ErrRecordNotFound {
			return err
		}

		// 收藏成功与否只看本次恢复/新建是否真正生效。
		createdOrRestored, err := dbservice.RestoreOrCreateUnique(tx, dbservice.UniqueWriteOptions{
			Value: &models.UserArticleFavorModel{
				ArticleID: cr.ArticleID,
				UserID:    claims.UserID,
				FavorID:   favorite.ID,
			},
			Match: []string{"article_id", "user_id", "favor_id"},
		})
		if err != nil {
			return err
		}
		if !createdOrRestored {
			return errors.New("请勿重复收藏")
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
				_, err := dbservice.RestoreOrCreateUnique(db, dbservice.UniqueWriteOptions{
					Value: &models.FavoriteModel{
						UserID:    userID,
						Title:     "默认收藏夹",
						IsDefault: true,
					},
					Match: []string{"user_id", "title"},
				})
				if err != nil {
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
