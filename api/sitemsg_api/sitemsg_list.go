package sitemsg_api

import (
	"myblogx/common"
	"myblogx/common/res"
	"myblogx/global"
	"myblogx/middleware"
	"myblogx/models"
	"myblogx/models/enum/message_enum"
	"myblogx/utils/jwts"

	"github.com/gin-gonic/gin"
)

func (a *SitemsgApi) SitemsgListView(c *gin.Context) {
	cr := middleware.GetBindQuery[SitemsgListRequest](c)

	var typeList []message_enum.Type
	switch cr.T {
	case 1:
		typeList = append(typeList, message_enum.CommentArticleType, message_enum.CommentReplyType)
	case 2:
		typeList = append(typeList, message_enum.DiggArticleType, message_enum.DiggCommentType, message_enum.FavorArticleType)
	case 3:
		typeList = append(typeList, message_enum.SystemType)
	}

	claims := jwts.MustGetClaimsByGin(c)

	list, count, err := common.ListQuery(models.ArticleMessageModel{
		ReceiverID: claims.UserID,
	}, common.Options{
		PageInfo: cr.PageInfo,
		Where:    global.DB.Where("type in ?", typeList),
	})
	if err != nil {
		res.FailWithError(err, c)
		return
	}

	res.OkWithList(list, count, c)
}
