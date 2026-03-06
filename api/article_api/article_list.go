package article_api

import (
	"fmt"
	"myblogx/common"
	"myblogx/common/res"
	"myblogx/global"
	"myblogx/middleware"
	"myblogx/models"
	"myblogx/models/enum"
	"myblogx/service/redis_service/redis_article"
	"myblogx/utils/jwts"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ArticleListRequest struct {
	common.PageInfo
	// 1 查自己的文章，2 查别人的文章，3 管理员查文章
	Type       int8               `form:"type" binding:"required,oneof=1 2 3"`
	UserID     uint               `form:"user_id"`
	CategoryID *uint              `form:"category_id"`
	TagID      *uint              `form:"tag_id"`
	Status     enum.ArticleStatus `form:"status"`
}

type ArticleListResponse struct {
	models.ArticleModel
	UserTop       bool   `json:"user_top"`  // 是否置顶
	AdminTop      bool   `json:"admin_top"` // 是否管理员置顶
	CategoryTitle string `json:"category_title"`
	UserNickname  string `json:"user_nickname"`
	UserAvatar    string `json:"user_avatar"`
}

// 排序字段校验
var orderColumnMap = map[string]string{
	"view_count desc":    "article_models.view_count desc",
	"digg_count desc":    "article_models.digg_count desc",
	"comment_count desc": "article_models.comment_count desc",
	"favor_count desc":   "article_models.favor_count desc",
	"view_count asc":     "article_models.view_count asc",
	"digg_count asc":     "article_models.digg_count asc",
	"comment_count asc":  "article_models.comment_count asc",
	"favor_count asc":    "article_models.favor_count asc",
}

// 获取文章列表
func (ArticleApi) ArticleListView(c *gin.Context) {
	cr := middleware.GetBindQuery[ArticleListRequest](c)
	claims, _ := jwts.ParseTokenByGin(c)

	normalized, err := validateRequest(cr, claims, c)
	if err != nil {
		return
	}

	// 处理文章置顶逻辑
	userTopMap, adminTopMap, defaultOrder := handleTopArticles(normalized.UserID)
	// 构造基础查询
	baseQuery := buildArticleListQuery(normalized)

	var count int64
	if err := baseQuery.Session(&gorm.Session{}).
		Count(&count).Error; err != nil {
		res.FailWithMsg("查询文章数量失败", c)
		return
	}

	page := normalized.PageInfo.GetPage(int(count))
	limit := normalized.PageInfo.GetLimit()
	offset := (page - 1) * limit

	order := defaultOrder
	if normalized.PageInfo.Order != "" {
		var ok bool
		order, ok = orderColumnMap[normalized.PageInfo.Order]
		if !ok {
			res.FailWithMsg("排序字段错误", c)
			return
		}
	}

	var articleIDs []uint
	if err := baseQuery.Session(&gorm.Session{}).
		Select("article_models.id").
		Order(order).
		Limit(limit).
		Offset(offset).
		Pluck("article_models.id", &articleIDs).Error; err != nil {
		res.FailWithMsg("查询文章失败", c)
		return
	}

	if len(articleIDs) == 0 {
		res.OkWithList([]ArticleListResponse{}, int(count), c)
		return
	}

	var articleList []models.ArticleModel
	if err := global.DB.Where("id IN ?", articleIDs).
		Preload("CategoryModel").
		Preload("UserModel").
		Preload("Tags", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort desc, id asc")
		}).
		Find(&articleList).Error; err != nil {
		res.FailWithMsg("查询文章失败", c)
		return
	}

	articleMap := make(map[uint]models.ArticleModel, len(articleList))
	for _, item := range articleList {
		articleMap[item.ID] = item
	}

	favorMap := redis_article.GetBatchCacheFavorite(articleIDs)
	diggMap := redis_article.GetBatchCacheDigg(articleIDs)
	viewMap := redis_article.GetBatchCacheView(articleIDs)
	commentMap := redis_article.GetBatchCacheComment(articleIDs)

	responseList := make([]ArticleListResponse, 0, len(articleIDs))
	for _, articleID := range articleIDs {
		model, ok := articleMap[articleID]
		if !ok {
			continue
		}

		model.DiggCount += diggMap[model.ID]
		model.FavorCount += favorMap[model.ID]
		model.ViewCount += viewMap[model.ID]
		model.CommentCount += commentMap[model.ID]

		item := ArticleListResponse{
			ArticleModel: model,
			UserTop:      userTopMap[model.ID],
			AdminTop:     adminTopMap[model.ID],
			UserNickname: model.UserModel.Nickname,
			UserAvatar:   model.UserModel.Avatar,
		}
		if model.CategoryModel != nil {
			item.CategoryTitle = model.CategoryModel.Title
		}
		responseList = append(responseList, item)
	}

	res.OkWithList(responseList, int(count), c)
}

func validateRequest(cr ArticleListRequest, claims *jwts.MyClaims, c *gin.Context) (ArticleListRequest, error) {
	switch cr.Type {
	case 1:
		if cr.UserID == 0 {
			res.FailWithMsg("用户 id 不能为空", c)
			return cr, fmt.Errorf("user_id is required")
		}
		if claims == nil && (cr.Page > 1 || cr.Limit > 10) {
			res.FailWithMsg("想查看更多内容，请先登录", c)
			return cr, fmt.Errorf("login required")
		}
		if cr.Status != 0 && cr.Status != enum.ArticleStatusPublished {
			res.FailWithMsg("只能查看已发布的文章", c)
			return cr, fmt.Errorf("invalid status")
		}
	case 2:
		if claims == nil {
			res.FailWithMsg("未登录", c)
			return cr, fmt.Errorf("unauthorized")
		}
		cr.UserID = claims.UserID
	case 3:
		if claims == nil || !claims.IsAdmin() {
			res.FailWithMsg("权限错误", c)
			return cr, fmt.Errorf("forbidden")
		}
	default:
		res.FailWithMsg("查询类型错误", c)
		return cr, fmt.Errorf("invalid type")
	}
	return cr, nil
}

func buildArticleListQuery(cr ArticleListRequest) *gorm.DB {
	query := global.DB.Model(&models.ArticleModel{}).
		Where(&models.ArticleModel{
			AuthorID:   cr.UserID,
			CategoryID: cr.CategoryID,
			Status:     cr.Status,
		})

	if cr.Key != "" {
		query = query.Where("article_models.title LIKE ?", "%"+cr.Key+"%")
	}

	if cr.TagID != nil {
		query = query.Joins("JOIN article_tag_models ON article_tag_models.article_id = article_models.id").
			Where("article_tag_models.tag_id = ?", *cr.TagID)
	}

	return query
}

func handleTopArticles(userID uint) (map[uint]bool, map[uint]bool, string) {
	userTopMap := make(map[uint]bool)
	adminTopMap := make(map[uint]bool)
	order := "article_models.created_at desc"

	if userID == 0 {
		return userTopMap, adminTopMap, order
	}

	var topRows []struct {
		ArticleID uint
	}
	if err := global.DB.Model(&models.UserTopArticleModel{}).
		Select("article_id").
		Where("user_id = ?", userID).
		Order("created_at desc").
		Find(&topRows).Error; err != nil {
		return userTopMap, adminTopMap, order
	}
	if len(topRows) == 0 {
		return userTopMap, adminTopMap, order
	}

	parts := make([]string, 0, len(topRows))
	for _, item := range topRows {
		parts = append(parts, fmt.Sprintf("article_models.id in (%d) desc", item.ArticleID))
		userTopMap[item.ArticleID] = true
	}

	order = fmt.Sprintf("%s, article_models.created_at desc", strings.Join(parts, ","))
	return userTopMap, adminTopMap, order
}
