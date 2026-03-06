package view_history

import (
	"myblogx/common"
	"myblogx/common/res"
	"myblogx/middleware"
	"myblogx/models"
	"myblogx/utils/jwts"

	"github.com/gin-gonic/gin"
)

func (ViewHistoryApi) ArticleViewHistoryView(c *gin.Context) {
	cr := middleware.GetBindQuery[ArticleViewHistoryRequest](c)
	claims := jwts.GetClaimsByGin(c)

	switch cr.Type {
	case 1:
		cr.UserID = claims.UserID
	case 2:
	}

	_list, count, err := common.ListQuery(models.UserArticleViewHistoryModel{
		UserID: cr.UserID,
	}, common.Options{
		PageInfo: cr.PageInfo,
		Preloads: []string{"UserModel", "ArticleModel"},
	})
	if err != nil {
		res.FailWithError(err, c)
		return
	}

	var list = make([]ArticleViewHistoryResponse, 0)
	for _, item := range _list {
		list = append(list, ArticleViewHistoryResponse{
			UpdatedAt: item.UpdatedAt,
			Title:     item.ArticleModel.Title,
			Cover:     item.ArticleModel.Cover,
			Nickname:  item.UserModel.Nickname,
			Avatar:    item.UserModel.Avatar,
			UserID:    item.UserID,
			ArticleID: item.ArticleID,
		})
	}

	res.OkWithList(list, count, c)
}
