package article_api

import (
	"fmt"
	"myblogx/common"
	"myblogx/common/res"
	"myblogx/global"
	"myblogx/middleware"
	"myblogx/models"
	"myblogx/models/enum"
	"myblogx/utils/jwts"
	"strings"

	"github.com/gin-gonic/gin"
)

type ArticleListRequest struct {
	common.PageInfo
	// 1 查别人的 2 查自己的 3 管理员查
	Type       int8               `form:"type" binding:"required,oneof=1 2 3"`
	UserID     uint               `form:"user_id"`
	CategoryID *uint              `form:"category_id"`
	Status     enum.ArticleStatus `form:"status"`
}

type ArticleListResponse struct {
	models.ArticleModel
	UserTop  bool `json:"user_top"`  // 是否为用户置顶
	AdminTop bool `json:"admin_top"` // 是否为管理员置顶
}

// 排序字段校验
var orderColumnMap = map[string]bool{
	"view_count desc":    true,
	"digg_count desc":    true,
	"comment_count desc": true,
	"favor_count desc":   true,
	"view_count asc":     true,
	"digg_count asc":     true,
	"comment_count asc":  true,
	"favor_count asc":    true,
}

// ArticleListView 获取文章列表
func (ArticleApi) ArticleListView(c *gin.Context) {
	cr := middleware.GetBindQuery[ArticleListRequest](c)
	claims, _ := jwts.ParseTokenByGin(c)

	// 验证请求参数和权限
	if err := validateRequest(cr, claims, c); err != nil {
		return
	}

	// 处理置顶文章逻辑
	userTopMap, adminTopMap, order := handleTopArticles(cr.UserID)

	// 查询文章列表
	_list, count, err := common.ListQuery(models.ArticleModel{
		AuthorID:   cr.UserID,
		CategoryID: cr.CategoryID,
		Status:     cr.Status,
	}, common.Options{
		Likes:        []string{"title"},
		PageInfo:     cr.PageInfo,
		DefaultOrder: order,
		OrderMap:     orderColumnMap,
	})
	if err != nil {
		res.FailWithMsg(err.Error(), c)
		return
	}

	// 构建响应数据
	responseList := make([]ArticleListResponse, 0, len(_list))
	for _, model := range _list {
		responseList = append(responseList, ArticleListResponse{
			ArticleModel: model,
			UserTop:      userTopMap[model.ID],
			AdminTop:     adminTopMap[model.ID],
		})
	}

	res.OkWithList(responseList, count, c)
}

// validateRequest 验证请求参数和权限
func validateRequest(cr ArticleListRequest, claims *jwts.MyClaims, c *gin.Context) error {
	switch cr.Type {
	case 1:
		// 查别人的文章
		if cr.UserID == 0 {
			res.FailWithMsg("用户 id 不能为空", c)
			return fmt.Errorf("用户 id 不能为空")
		}
		if claims == nil && (cr.Page > 1 || cr.Limit > 10) {
			res.FailWithMsg("想查询更多内容，请进行登录", c)
			return fmt.Errorf("想查询更多内容，请进行登录")
		}
		if cr.Status != 0 && cr.Status != enum.ArticleStatusPublished {
			res.FailWithMsg("只能查已发布的文章", c)
			return fmt.Errorf("只能查已发布的文章")
		}
	case 2:
		// 查自己的文章
		if claims == nil {
			res.FailWithMsg("未登录", c)
			return fmt.Errorf("未登录")
		}
		cr.UserID = claims.UserID
	case 3:
		// 管理员查
		if claims == nil || !(claims.Role == enum.RoleAdmin) {
			res.FailWithMsg("权限错误", c)
			return fmt.Errorf("权限错误")
		}
	}
	return nil
}

// handleTopArticles 处理置顶文章逻辑
func handleTopArticles(userID uint) (userTopMap map[uint]bool, adminTopMap map[uint]bool, order string) {
	userTopMap = make(map[uint]bool)
	adminTopMap = make(map[uint]bool)
	order = "created_at desc"

	if userID == 0 {
		return
	}

	// 查询用户置顶文章，只选择需要的字段
	var userTopArticleList []struct {
		ArticleID uint
		Role      enum.RoleType
	}

	// 使用连接查询获取置顶文章和用户角色信息，避免预加载
	global.DB.Preload("UserModel").Order("created_at desc").Find(&userTopArticleList, "user_id = ?", userID)

	if len(userTopArticleList) > 0 {
		var topArticleList []string
		for _, item := range userTopArticleList {
			topArticleList = append(topArticleList, fmt.Sprintf("id in (%d) desc", item.ArticleID))
			if item.Role == enum.RoleAdmin {
				adminTopMap[item.ArticleID] = true
			}
			userTopMap[item.ArticleID] = true
		}

		order = fmt.Sprintf("%s, created_at desc", strings.Join(topArticleList, ","))
	}

	return
}
