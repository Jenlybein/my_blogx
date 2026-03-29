package article_api

import (
	"errors"
	"myblogx/common/res"
	"myblogx/global"
	"myblogx/middleware"
	"myblogx/models"
	"myblogx/models/ctype"
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
	if err := global.DB.Select("id", "author_id", "title").
		Take(&article, "id = ? and status = ?", cr.ArticleID, enum.ArticleStatusPublished).Error; err != nil {
		res.FailWithMsg("查询文章失败", c)
		return
	}

	var isFavorited bool
	if err := global.DB.Transaction(func(tx *gorm.DB) error {
		favorite, err := getOrCreateFavoriteID(tx, cr.FavorID, claims.UserID)
		if err != nil {
			return err
		}

		isFavorited, err = switchArticleFavorite(tx, cr.ArticleID, claims.UserID, favorite.ID)
		return err
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
			global.Logger.Errorf("文章收藏数据加一失败: 错误=%v", err)
		}
		res.OkWithMsg("收藏成功", c)
	} else {
		if err := redis_article.SetCacheFavorite(cr.ArticleID, -1); err != nil {
			global.Logger.Errorf("文章收藏数据减一失败: 错误=%v", err)
		}
		res.OkWithMsg("取消收藏成功", c)
	}
}

// getOrCreateFavoriteID 获取收藏夹；如果获取的是默认收藏夹，不存在就新建
func getOrCreateFavoriteID(db *gorm.DB, favorID, userID ctype.ID) (*models.FavoriteModel, error) {
	var favorite models.FavoriteModel
	if favorID == 0 {
		if err := db.Where("is_default = ? AND user_id = ?", true, userID).
			Attrs(models.FavoriteModel{
				UserID:    userID,
				Title:     "默认收藏夹",
				IsDefault: true,
			}).
			FirstOrCreate(&favorite).Error; err != nil {
			return nil, errors.New("创建默认收藏夹失败")
		}
	} else {
		if err := db.Select("id").Take(&favorite, "id = ? and user_id = ?", favorID, userID).Error; err != nil {
			return nil, errors.New("收藏夹不存在")
		}
	}
	return &favorite, nil
}

// switchArticleFavorite 只处理单个收藏关系的切换，避免收藏夹解析和副作用逻辑混在一起。
func switchArticleFavorite(tx *gorm.DB, articleID, userID, favorID ctype.ID) (bool, error) {
	var articleFavorite models.UserArticleFavorModel
	if err := tx.Select("id").
		Take(&articleFavorite, "article_id = ? and user_id = ? and favor_id = ?", articleID, userID, favorID).Error; err == nil {
		// 取消收藏必须看本次 Delete 是否真的删掉了活记录，避免并发下双成功。
		deleteResult := tx.Where("article_id = ? and user_id = ? and favor_id = ?", articleID, userID, favorID).
			Delete(&models.UserArticleFavorModel{})
		if deleteResult.Error != nil {
			return false, deleteResult.Error
		}
		if deleteResult.RowsAffected == 0 {
			return false, errors.New("收藏状态已变化，请刷新后重试")
		}
		return false, nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return false, err
	}

	// 收藏成功与否只看本次恢复/新建是否真正生效。
	createdOrRestored, err := dbservice.RestoreOrCreateUnique(tx, &models.UserArticleFavorModel{
		ArticleID: articleID,
		UserID:    userID,
		FavorID:   favorID,
	}, []string{"article_id", "user_id", "favor_id"})
	if err != nil {
		return false, err
	}
	if !createdOrRestored {
		return false, errors.New("请勿重复收藏")
	}
	return true, nil
}
