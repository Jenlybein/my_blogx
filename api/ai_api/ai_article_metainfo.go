package ai_api

import (
	"myblogx/common/res"
	"myblogx/middleware"
	"myblogx/service/ai_service"
	"myblogx/utils/jwts"

	"github.com/gin-gonic/gin"
)

func (AIApi) AIArticleMetaInfoView(c *gin.Context) {
	cr := middleware.GetBindJson[AIArticleMetaInfoRequest](c)
	claims := jwts.MustGetClaimsByGin(c)

	data, err := ai_service.GenerateArticleMetainfo(claims.UserID, cr.Content)
	if err != nil {
		res.FailWithMsg(err.Error(), c)
		return
	}

	res.OkWithData(data, c)
}
