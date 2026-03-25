package article_api

import (
	"errors"
	"fmt"
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
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// orderColumnMap 将前端允许传入的排序参数映射为实际 SQL 排序表达式。
// 这里只开放计数字段排序，避免把任意字符串直接拼进 Order。
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

// ArticleListView 使用“两段查询”返回文章列表：
// 1. 先按筛选条件、排序规则分页取出当前页文章 ID
// 2. 再按这些 ID 回表查询完整文章信息并预加载关联
//
// 这样做的原因是文章列表存在标签筛选、置顶排序、关联预加载等需求，
// 直接一条 SQL 同时处理 join、分页、排序和 preload，行为更容易失稳。
func (ArticleApi) ArticleListView(c *gin.Context) {
	cr := middleware.GetBindQuery[ArticleListRequest](c)
	var claims *jwts.MyClaims
	if authResult := user_service.MustAuthenticateAccessTokenByGin(c); authResult != nil {
		claims = authResult.Claims
	}

	normalized, err := validateRequest(cr, claims, c)
	if err != nil {
		return
	}

	// 置顶文章会拼进默认排序里，因此先拿到置顶映射和默认排序表达式。
	userTopMap, adminTopMap, defaultOrder := handleTopArticles(normalized.UserID)
	// 基础查询只负责筛选文章，不负责预加载关联。
	baseQuery := buildArticleListQuery(normalized)

	// 第一阶段：只分页获取当前页文章 ID，避免 join 后直接查整行带来的分页和排序问题。
	articleIDs, count, err := common.PageIDQuery(baseQuery, common.IDPageOptions{
		PageInfo:     normalized.PageInfo,
		IDColumn:     "article_models.id",
		OrderMap:     orderColumnMap,
		DefaultOrder: defaultOrder,
	})
	if err != nil {
		if errors.Is(err, common.ErrInvalidOrder) {
			res.FailWithMsg(err.Error(), c)
			return
		}
		res.FailWithMsg("查询文章失败", c)
		return
	}

	if len(articleIDs) == 0 {
		res.OkWithList([]ArticleListResponse{}, count, c)
		return
	}

	// 第二阶段：按主键集合回表查询详情，并预加载分类、作者、标签等展示信息。
	var articleList []models.ArticleModel
	if err := global.DB.Select(
		"ID",
		"CreatedAt",
		"UpdatedAt",
		"Title",
		"Abstract",
		"Content",
		"Cover",
		"ViewCount",
		"DiggCount",
		"CommentCount",
		"FavorCount",
		"CommentsToggle",
		"Status",
	).Where("id IN ?", articleIDs).
		Preload("CategoryModel", func(db *gorm.DB) *gorm.DB { return db.Select("id", "title") }).
		Preload("UserModel", func(db *gorm.DB) *gorm.DB { return db.Select("id", "nickname", "avatar") }).
		Preload("Tags", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "title", "sort").Order("sort desc, id asc")
		}).
		Find(&articleList).Error; err != nil {
		res.FailWithMsg("查询文章失败", c)
		return
	}

	// 回表查询不保证和 articleIDs 顺序完全一致，因此先转成 map，再按 articleIDs 顺序组装响应。
	articleMap := make(map[ctype.ID]models.ArticleModel, len(articleList))
	for _, item := range articleList {
		articleMap[item.ID] = item
	}

	// 文章计数采用“数据库基础值 + Redis 增量”模式，这里统一叠加最新变化量。
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
			ID:             model.ID,
			CreatedAt:      model.CreatedAt,
			UpdatedAt:      model.UpdatedAt,
			Title:          model.Title,
			Abstract:       model.Abstract,
			Content:        model.Content,
			Cover:          model.Cover,
			ViewCount:      model.ViewCount,
			DiggCount:      model.DiggCount,
			CommentCount:   model.CommentCount,
			FavorCount:     model.FavorCount,
			CommentsToggle: model.CommentsToggle,
			Status:         model.Status,
			UserTop:        userTopMap[model.ID],
			AdminTop:       adminTopMap[model.ID],
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

	res.OkWithList(responseList, count, c)
}

// validateRequest 负责把不同查询类型的权限和参数要求收敛成统一规则。
// 返回的 cr 可能会被修正，例如 type=2 时会强制改写为当前登录用户的 user_id。
func validateRequest(cr ArticleListRequest, claims *jwts.MyClaims, c *gin.Context) (ArticleListRequest, error) {
	switch cr.Type {
	case 1:
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

// buildArticleListQuery 只拼接文章主查询条件。
// 如果按标签筛选，必须使用 join article_tag_models 让标签关系参与主查询过滤；
// 仅 Preload("Tags") 只能回填标签数据，不能筛掉不满足标签条件的文章。
func buildArticleListQuery(cr ArticleListRequest) *gorm.DB {
	query := global.DB.Model(&models.ArticleModel{}).
		Where(&models.ArticleModel{
			CategoryID: cr.CategoryID,
			Status:     cr.Status,
		})

	if cr.UserID != 0 {
		query = query.Where("article_models.author_id = ?", cr.UserID)
	}

	if cr.Key != "" {
		query = query.Where("article_models.title LIKE ?", "%"+cr.Key+"%")
	}

	if cr.TagID != nil {
		query = query.Joins("JOIN article_tag_models ON article_tag_models.article_id = article_models.id").
			Where("article_tag_models.tag_id = ?", *cr.TagID)
	}

	return query
}

// handleTopArticles 计算作者置顶 / 管理员置顶的状态映射和默认排序表达式。
// 默认情况下管理员置顶优先；查看某位作者的文章时，再叠加该作者自己的置顶顺序。
func handleTopArticles(userID ctype.ID) (map[ctype.ID]bool, map[ctype.ID]bool, string) {
	userTopMap := make(map[ctype.ID]bool)
	adminTopMap := make(map[ctype.ID]bool)
	orderParts := make([]string, 0)
	orderedArticleMap := make(map[ctype.ID]struct{})

	appendOrder := func(articleID ctype.ID) {
		if articleID == 0 {
			return
		}
		if _, ok := orderedArticleMap[articleID]; ok {
			return
		}
		orderedArticleMap[articleID] = struct{}{}
		orderParts = append(orderParts, fmt.Sprintf("article_models.id in (%d) desc", articleID))
	}

	var adminTopRows []struct {
		ArticleID ctype.ID
	}
	if err := global.DB.Model(&models.UserTopArticleModel{}).
		Select("user_top_article_models.article_id").
		Joins("JOIN user_models ON user_models.id = user_top_article_models.user_id").
		Where("user_models.role = ?", enum.RoleAdmin).
		Order("user_top_article_models.created_at desc").
		Find(&adminTopRows).Error; err == nil {
		for _, item := range adminTopRows {
			adminTopMap[item.ArticleID] = true
			appendOrder(item.ArticleID)
		}
	}

	if userID != 0 {
		var userTopRows []struct {
			ArticleID ctype.ID
		}
		if err := global.DB.Model(&models.UserTopArticleModel{}).
			Select("user_top_article_models.article_id").
			Joins("JOIN article_models ON article_models.id = user_top_article_models.article_id").
			Where("user_top_article_models.user_id = ? AND article_models.author_id = ?", userID, userID).
			Order("user_top_article_models.created_at desc").
			Find(&userTopRows).Error; err == nil {
			for _, item := range userTopRows {
				userTopMap[item.ArticleID] = true
				appendOrder(item.ArticleID)
			}
		}
	}

	if len(orderParts) == 0 {
		return userTopMap, adminTopMap, "article_models.created_at desc"
	}

	order := fmt.Sprintf("%s, article_models.created_at desc", strings.Join(orderParts, ","))
	return userTopMap, adminTopMap, order
}
