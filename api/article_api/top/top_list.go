package top

import (
	"myblogx/common/res"
	"myblogx/global"
	"myblogx/middleware"
	"myblogx/models"
	"myblogx/models/ctype"
	"myblogx/models/enum"
	"myblogx/service/redis_service/redis_article"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func (TopApi) ArticleTopListView(c *gin.Context) {
	cr := middleware.GetBindQuery[ArticleTopListRequest](c)

	if cr.Type == 1 && cr.UserID == 0 {
		res.FailWithMsg("请选择作者", c)
		return
	}

	baseQuery := buildArticleTopListQuery(cr)
	var articleIDs []ctype.ID
	if err := baseQuery.
		Order("top_list.top_created_at desc, article_models.id desc").
		Pluck("article_models.id", &articleIDs).Error; err != nil {
		res.FailWithMsg("查询置顶文章失败", c)
		return
	}

	if len(articleIDs) == 0 {
		res.OkWithList([]ArticleTopListResponse{}, 0, c)
		return
	}

	articleList, err := loadTopArticlesByIDs(articleIDs)
	if err != nil {
		res.FailWithMsg("查询置顶文章失败", c)
		return
	}

	articleMap := make(map[ctype.ID]models.ArticleModel, len(articleList))
	for _, item := range articleList {
		articleMap[item.ID] = item
	}

	counters := redis_article.GetBatchCounters(articleIDs)

	extraTopMap, err := loadExtraTopMap(cr.Type, articleIDs)
	if err != nil {
		res.FailWithMsg("查询置顶文章失败", c)
		return
	}

	responseList := make([]ArticleTopListResponse, 0, len(articleIDs))
	for _, articleID := range articleIDs {
		model, ok := articleMap[articleID]
		if !ok {
			continue
		}

		model.DiggCount += counters.DiggMap[model.ID]
		model.FavorCount += counters.FavorMap[model.ID]
		model.ViewCount += counters.ViewMap[model.ID]
		model.CommentCount += counters.CommentMap[model.ID]

		userTop := cr.Type == 1
		adminTop := cr.Type == 2
		if cr.Type == 1 {
			adminTop = extraTopMap[model.ID]
		} else {
			userTop = extraTopMap[model.ID]
		}

		item := ArticleTopListResponse{
			ID:             model.ID,
			CreatedAt:      model.CreatedAt,
			UpdatedAt:      model.UpdatedAt,
			Title:          model.Title,
			Abstract:       model.Abstract,
			Cover:          model.Cover,
			ViewCount:      model.ViewCount,
			DiggCount:      model.DiggCount,
			CommentCount:   model.CommentCount,
			FavorCount:     model.FavorCount,
			CommentsToggle: model.CommentsToggle,
			Status:         model.Status,
			UserTop:        userTop,
			AdminTop:       adminTop,
			UserNickname:   model.UserModel.Nickname,
			UserAvatar:     model.UserModel.Avatar,
		}
		if model.CategoryModel != nil {
			item.CategoryTitle = model.CategoryModel.Title
		}
		if model.Tags != nil {
			for _, tag := range model.Tags {
				item.Tags = append(item.Tags, tag.Title)
			}
		}
		responseList = append(responseList, item)
	}

	res.OkWithList(responseList, len(responseList), c)
}

func buildArticleTopListQuery(cr ArticleTopListRequest) *gorm.DB {
	subQuery := global.DB.Model(&models.UserTopArticleModel{}).
		Select("user_top_article_models.article_id, MAX(user_top_article_models.created_at) AS top_created_at")

	switch cr.Type {
	case 1:
		subQuery = subQuery.
			Joins("JOIN article_models ON article_models.id = user_top_article_models.article_id").
			Where("user_top_article_models.user_id = ? AND article_models.author_id = ?", cr.UserID, cr.UserID)
	case 2:
		subQuery = subQuery.
			Joins("JOIN user_models ON user_models.id = user_top_article_models.user_id").
			Where("user_models.role = ?", enum.RoleAdmin)
	}

	subQuery = subQuery.Group("user_top_article_models.article_id")

	return global.DB.Table("(?) AS top_list", subQuery).
		Joins("JOIN article_models ON article_models.id = top_list.article_id").
		Where("article_models.status = ?", enum.ArticleStatusPublished)
}

func loadTopArticlesByIDs(articleIDs []ctype.ID) ([]models.ArticleModel, error) {
	var articleList []models.ArticleModel
	err := global.DB.Select(
		"id",
		"created_at",
		"updated_at",
		"title",
		"abstract",
		"cover",
		"view_count",
		"digg_count",
		"comment_count",
		"favor_count",
		"comments_toggle",
		"status",
	).Where("id IN ?", articleIDs).
		Preload("CategoryModel", func(db *gorm.DB) *gorm.DB { return db.Select("id", "title") }).
		Preload("UserModel", func(db *gorm.DB) *gorm.DB { return db.Select("id", "nickname", "avatar") }).
		Preload("Tags", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "title", "sort").Order("sort desc, id asc")
		}).
		Find(&articleList).Error
	return articleList, err
}

func loadExtraTopMap(topType int, articleIDs []ctype.ID) (map[ctype.ID]bool, error) {
	extraTopMap := make(map[ctype.ID]bool, len(articleIDs))
	if len(articleIDs) == 0 {
		return extraTopMap, nil
	}

	var rows []struct {
		ArticleID ctype.ID
	}

	query := global.DB.Model(&models.UserTopArticleModel{}).
		Select("DISTINCT user_top_article_models.article_id").
		Where("user_top_article_models.article_id IN ?", articleIDs)

	switch topType {
	case 1:
		query = query.
			Joins("JOIN user_models ON user_models.id = user_top_article_models.user_id").
			Where("user_models.role = ?", enum.RoleAdmin)
	case 2:
		query = query.
			Joins("JOIN article_models ON article_models.id = user_top_article_models.article_id").
			Where("user_top_article_models.user_id = article_models.author_id")
	default:
		return extraTopMap, nil
	}

	if err := query.Find(&rows).Error; err != nil {
		return extraTopMap, err
	}

	for _, row := range rows {
		extraTopMap[row.ArticleID] = true
	}
	return extraTopMap, nil
}
