package search_api

import (
	"myblogx/common/res"
	"myblogx/middleware"
	"myblogx/service/search_service"
	"myblogx/service/user_service"
	"myblogx/utils/jwts"

	"github.com/gin-gonic/gin"
)

func (SearchApi) ArticleSearchView(c *gin.Context) {
	cr := middleware.GetBindQuery[search_service.ArticleSearchRequest](c)

	var claims *jwts.MyClaims
	if authResult := user_service.MustAuthenticateAccessTokenByGin(c); authResult != nil {
		claims = authResult.Claims
	}

	list, count, err := search_service.SearchArticles(cr, claims)
	if err != nil {
		res.FailWithMsg(err.Error(), c)
		return
	}
	res.OkWithList(list, count, c)
}
