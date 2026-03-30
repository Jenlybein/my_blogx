package favorite

import (
	"myblogx/common"
	"myblogx/common/res"
	"myblogx/global"
	"myblogx/middleware"
	"myblogx/models"
	"myblogx/models/ctype"
	"myblogx/service/user_service"
	"myblogx/utils/jwts"

	"github.com/gin-gonic/gin"
)

// 查询收藏夹列表
func (FavoriteApi) FavoriteListView(c *gin.Context) {
	cr := middleware.GetBindQuery[FavoriteListRequest](c)

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
	case 2:
		if cr.UserID == 0 {
			res.FailWithMsg("用户 id 不能为空", c)
			return
		}
		// 查询目标用户隐私设置，判断是否公开收藏夹
		var userConf models.UserConfModel
		if err := global.DB.Take(&userConf, "user_id = ?", cr.UserID).Error; err != nil {
			res.FailWithMsg("用户不存在", c)
			return
		}
		if !userConf.FavoritesVisibility {
			res.FailWithMsg("收藏夹不公开", c)
			return
		}
	case 3:
		if err != nil || claim.IsAdmin() == false {
			res.FailWithMsg("权限不足", c)
			return
		}
		preloads = append(preloads, "UserModel")
	}

	_list, count, err := common.ListQuery(models.FavoriteModel{
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

	// 判断是否已收藏
	hasArticleMap := make(map[ctype.ID]bool, len(_list))
	if cr.Type == 1 && cr.ArticleID != 0 && len(_list) > 0 {
		favoriteIDs := make([]ctype.ID, 0, len(_list))
		for _, item := range _list {
			favoriteIDs = append(favoriteIDs, item.ID)
		}

		var relationList []models.UserArticleFavorModel
		if err = global.DB.Select("favor_id").
			Where("user_id = ? AND article_id = ? AND favor_id IN ?", claim.UserID, cr.ArticleID, favoriteIDs).
			Find(&relationList).Error; err != nil {
			res.FailWithError(err, c)
			return
		}
		for _, relation := range relationList {
			hasArticleMap[relation.FavorID] = true
		}
	}

	var list = make([]FavoriteListResponse, 0)
	for _, item := range _list {
		list = append(list, FavoriteListResponse{
			FavoriteModel: item,
			ArticleCount:  len(item.ArticleList),
			Nickname:      item.UserModel.Nickname,
			Avatar:        item.UserModel.Avatar,
			HasArticle:    hasArticleMap[item.ID],
		})
	}

	res.OkWithList(list, count, c)
}
