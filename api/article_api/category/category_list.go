package category

import (
	"myblogx/common"
	"myblogx/common/res"
	"myblogx/middleware"
	"myblogx/models"
	"myblogx/service/user_service"
	"myblogx/utils/jwts"

	"github.com/gin-gonic/gin"
)

// 查询分类列表
func (CategoryApi) CategoryListView(c *gin.Context) {
	cr := middleware.GetBindQuery[CategoryListRequest](c)

	var claim *jwts.MyClaims
	var err error
	if authResult := user_service.MustAuthenticateAccessTokenByGin(c); authResult != nil {
		claim = authResult.Claims
	} else {
		err = user_service.ErrAuthRequired
	}

	preloads := []string{"ArticleList"}

	switch cr.Type {
	case 1:
		if err != nil {
			res.FailWithError(err, c)
			return
		}
		cr.UserID = claim.UserID
	case 2: //
	case 3:
		if err != nil || claim.IsAdmin() == false {
			res.FailWithMsg("权限不足", c)
			return
		}
		preloads = append(preloads, "UserModel")
	}

	_list, count, err := common.ListQuery(models.CategoryModel{
		UserID: cr.UserID,
	}, common.Options{
		PageInfo: cr.PageInfo,
		Likes:    []string{"title"},
		Preloads: preloads,
	})
	if err != nil {
		res.FailWithError(err, c)
		return
	}

	var list = make([]CategoryListResponse, 0)
	for _, item := range _list {
		list = append(list, CategoryListResponse{
			CategoryModel: item,
			ArticleCount:  len(item.ArticleList),
			Nickname:      item.UserModel.Nickname,
			Avatar:        item.UserModel.Avatar,
		})
	}

	res.OkWithList(list, count, c)
}
