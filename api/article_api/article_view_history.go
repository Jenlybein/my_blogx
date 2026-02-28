package article_api

import (
	"fmt"
	"myblogx/common"
	"myblogx/common/res"
	"myblogx/global"
	"myblogx/middleware"
	"myblogx/models"
	"myblogx/utils/jwts"
	"time"

	"github.com/gin-gonic/gin"
)

type ArticleViewHistoryRequest struct {
	common.PageInfo
	UserID uint `form:"user_id"`
	Type   int8 `form:"type" binding:"required,oneof=1 2"` // 1: 自己的浏览记录 2: 其他人的浏览记录
}

type ArticleViewHistoryResponse struct {
	UpdatedAt time.Time `json:"updated_at"`
	Title     string    `json:"title"`
	Cover     string    `json:"cover"`
	Nickname  string    `json:"nickname"`
	Avatar    string    `json:"avatar"`
	UserID    uint      `json:"user_id"`
	ArticleID uint      `json:"article_id"`
}

func (ArticleApi) ArticleViewHistoryView(c *gin.Context) {
	cr := middleware.GetBindQuery[ArticleViewHistoryRequest](c)
	claims := jwts.GetClaimsByGin(c)

	switch cr.Type {
	case 1:
		cr.UserID = claims.UserID
	case 2:
	}

	_list, count, _ := common.ListQuery(models.UserArticleViewHistoryModel{
		UserID: cr.UserID,
	}, common.Options{
		PageInfo: cr.PageInfo,
		Preloads: []string{"UserModel", "ArticleModel"},
	})

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

func (ArticleApi) ArticleViewHistoryRemoveView(c *gin.Context) {
	cr := middleware.GetBindJson[models.RemoveRequest](c)
	claims := jwts.GetClaimsByGin(c)

	var list []models.UserArticleViewHistoryModel
	if err := global.DB.Find(&list, "user_id = ? and article_id IN ?", claims.UserID, cr.IDList).Error; err != nil {
		res.FailWithError(err, c)
		return
	}
	if len(list) > 0 {
		if err := global.DB.Delete(&list).Error; err != nil {
			res.FailWithMsg(fmt.Sprintf("删除访问历史失败:%v", err), c)
			return
		}
	}

	res.OkWithMsg(fmt.Sprintf("访问历史删除成功，共%d条", len(list)), c)
}
