package article_api

import (
	"myblogx/common"
	"myblogx/common/res"
	"myblogx/middleware"
	"myblogx/models"
	"myblogx/models/enum"
	"myblogx/utils/jwts"

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

func (ArticleApi) ArticleListView(c *gin.Context) {
	cr := middleware.GetBindQuery[ArticleListRequest](c)

	switch cr.Type {
	case 1:
		// 查别人的文章
		if cr.UserID == 0 {
			res.FailWithMsg("用户 id 不能为空", c)
			return
		}
		if cr.Page > 1 || cr.Limit > 10 {
			res.FailWithMsg("想查询更多内容，请进行登录", c)
			return
		}
		if cr.Status != 0 && cr.Status != enum.ArticleStatusPublished {
			res.FailWithMsg("只能查已发布的文章", c)
			return
		}
	case 2:
		// 查自己的文章
		claims, err := jwts.ParseTokenByGin(c)
		if err != nil {
			res.FailWithMsg("未登录", c)
			return
		}
		cr.UserID = claims.UserID

	case 3:
		// 管理员查
		claims, err := jwts.ParseTokenByGin(c)
		if !(err == nil && claims.Role == enum.RoleAdmin) {
			res.FailWithMsg("权限错误", c)
			return
		}

	}

	_list, count, _ := common.ListQuery(models.ArticleModel{
		AuthorID:   cr.UserID,
		CategoryID: cr.CategoryID,
		Status:     cr.Status,
	}, common.Options{
		Likes: []string{"title"},
	})

	var list = make([]ArticleListResponse, 0)
	for _, model := range _list {
		// model.Content = ""
		// model.HtmlContent = ""
		list = append(list, ArticleListResponse{
			ArticleModel: model,
		})
	}

	res.OkWithList(list, count, c)
}
