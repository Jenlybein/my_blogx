package search_api

import (
	"myblogx/common/res"
	"myblogx/middleware"
	"myblogx/service/search_service"
	"myblogx/utils/jwts"

	"github.com/gin-gonic/gin"
)

func (SearchApi) ArticleSearchView(c *gin.Context) {
	cr := middleware.GetBindQuery[ArticleSearchRequest](c)
	claims, _ := jwts.ParseTokenByGin(c)
	list, count, err := search_service.SearchArticles(cr, claims)
	if err != nil {
		res.FailWithMsg(err.Error(), c)
		return
	}
	res.OkWithList(list, count, c)
}
