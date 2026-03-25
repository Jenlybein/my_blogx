package favorite

import (
	"errors"
	"myblogx/common"
	"myblogx/common/res"
	"myblogx/global"
	"myblogx/middleware"
	"myblogx/models"
	"myblogx/models/ctype"
	"myblogx/models/enum"
	"myblogx/service/redis_service/redis_article"
	"myblogx/service/user_service"
	"myblogx/utils/jwts"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var favoriteArticleOrderMap = map[string]string{
	"created_at desc": "user_article_favor_models.created_at desc",
	"created_at asc":  "user_article_favor_models.created_at asc",
}

// FavoriteArticlesView 查询某个收藏夹内的文章列表。
// 1. 收藏夹所有者和管理员可直接查看
// 2. 其他人只有在收藏夹公开时才能查看
// 3. 列表按收藏时间排序，支持按文章标题关键字筛选
func (FavoriteApi) FavoriteArticlesView(c *gin.Context) {
	cr := middleware.GetBindQuery[FavoriteArticlesRequest](c)
	var claims *jwts.MyClaims
	if authResult := user_service.MustAuthenticateAccessTokenByGin(c); authResult != nil {
		claims = authResult.Claims
	}

	favoriteModel, err := getAccessibleFavorite(c, cr.FavoriteID, claims)
	if err != nil {
		return
	}

	baseQuery := global.DB.Model(&models.UserArticleFavorModel{}).
		Joins("JOIN article_models ON article_models.id = user_article_favor_models.article_id").
		Where("user_article_favor_models.favor_id = ?", favoriteModel.ID).
		Where("article_models.status = ?", enum.ArticleStatusPublished)

	if cr.Key != "" {
		baseQuery = baseQuery.Where("article_models.title LIKE ?", "%"+cr.Key+"%")
	}

	articleIDs, count, err := common.PageIDQuery(baseQuery, common.IDPageOptions{
		PageInfo:     cr.PageInfo,
		IDColumn:     "user_article_favor_models.article_id",
		OrderMap:     favoriteArticleOrderMap,
		DefaultOrder: "user_article_favor_models.created_at desc",
	})
	if err != nil {
		if errors.Is(err, common.ErrInvalidOrder) {
			res.FailWithMsg(err.Error(), c)
			return
		}
		res.FailWithMsg("查询收藏夹文章失败", c)
		return
	}

	if len(articleIDs) == 0 {
		res.OkWithList([]FavoriteArticleResponse{}, count, c)
		return
	}

	var favorArticles []models.UserArticleFavorModel
	if err = global.DB.
		Where("favor_id = ? AND article_id IN ?", favoriteModel.ID, articleIDs).
		Preload("ArticleModel", func(db *gorm.DB) *gorm.DB {
			return db.Select(
				"id",
				"title",
				"abstract",
				"cover",
				"view_count",
				"digg_count",
				"comment_count",
				"favor_count",
				"status",
				"author_id",
			)
		}).
		Preload("ArticleModel.UserModel", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "nickname", "avatar")
		}).
		Find(&favorArticles).Error; err != nil {
		res.FailWithMsg("查询收藏夹文章失败", c)
		return
	}

	favorMap := make(map[ctype.ID]models.UserArticleFavorModel, len(favorArticles))
	for _, item := range favorArticles {
		favorMap[item.ArticleID] = item
	}

	favorCountMap := redis_article.GetBatchCacheFavorite(articleIDs)
	diggMap := redis_article.GetBatchCacheDigg(articleIDs)
	viewMap := redis_article.GetBatchCacheView(articleIDs)
	commentMap := redis_article.GetBatchCacheComment(articleIDs)

	list := make([]FavoriteArticleResponse, 0, len(articleIDs))
	for _, articleID := range articleIDs {
		item, ok := favorMap[articleID]
		if !ok {
			continue
		}
		article := item.ArticleModel
		article.ViewCount += viewMap[article.ID]
		article.DiggCount += diggMap[article.ID]
		article.CommentCount += commentMap[article.ID]
		article.FavorCount += favorCountMap[article.ID]

		list = append(list, FavoriteArticleResponse{
			FavoritedAt:   item.CreatedAt,
			ArticleID:     article.ID,
			Title:         article.Title,
			Abstract:      article.Abstract,
			Cover:         article.Cover,
			ViewCount:     article.ViewCount,
			DiggCount:     article.DiggCount,
			CommentCount:  article.CommentCount,
			FavorCount:    article.FavorCount,
			UserNickname:  article.UserModel.Nickname,
			UserAvatar:    article.UserModel.Avatar,
			ArticleStatus: article.Status,
		})
	}

	res.OkWithList(list, count, c)
}

func getAccessibleFavorite(c *gin.Context, favoriteID ctype.ID, claims *jwts.MyClaims) (*models.FavoriteModel, error) {
	var favoriteModel models.FavoriteModel
	if err := global.DB.Take(&favoriteModel, "id = ?", favoriteID).Error; err != nil {
		res.FailWithMsg("收藏夹不存在", c)
		return nil, err
	}

	if claims != nil {
		if claims.IsAdmin() || claims.UserID == favoriteModel.UserID {
			return &favoriteModel, nil
		}
	}

	var userConf models.UserConfModel
	if err := global.DB.Take(&userConf, "user_id = ?", favoriteModel.UserID).Error; err != nil {
		res.FailWithMsg("用户不存在", c)
		return nil, err
	}
	if !userConf.FavoritesVisibility {
		res.FailWithMsg("收藏夹不公开", c)
		return nil, errors.New("favorite not visible")
	}
	return &favoriteModel, nil
}
