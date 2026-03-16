package search_api

import (
	"myblogx/common/res"
	"myblogx/middleware"
	"myblogx/models"
	"myblogx/service/es_service"
	"myblogx/utils/jwts"

	"github.com/gin-gonic/gin"
)

func (SearchApi) ArticleSearchView(c *gin.Context) {
	cr := middleware.GetBindQuery[ArticleSearchRequest](c)
	page := cr.Page
	if page <= 0 {
		page = 1
	}

	query := buildDefaultArticleSearchQuery(cr.Key)
	extraBody := buildArticleSearchExtraBody("created_at")
	topMap := map[uint]int{}

	switch cr.Type {
	case 0:
		extraBody = buildArticleSearchExtraBody("created_at")
	case 1:
		claims, err := jwts.ParseTokenByGin(c)
		if err != nil || claims == nil {
			break
		}
		query = buildLikeTagsQuery(query, claims.UserID)
	case 2:
		extraBody = buildArticleSearchExtraBody("created_at")
	case 3:
		extraBody = buildArticleSearchExtraBody("comment_count")
	case 4:
		extraBody = buildArticleSearchExtraBody("digg_count")
	case 5:
		extraBody = buildArticleSearchExtraBody("favor_count")
	case 6:
		extraBody = buildArticleSearchExtraBody("view_count")
	case 7:
		query = buildTagListQuery(query, cr.TagList)
	case 8:
		query = buildUserIDQuery(query, cr.UserID)
	}

	if cr.Key == "" && cr.Type == 8 {
		query, topMap = buildAuthorAdminTopQuery(query, cr.UserID)
	} else if cr.Key == "" {
		query, topMap = buildAdminTopQuery(query)
	}

	resp := es_service.Search[map[string]any](
		models.ArticleModel{}.Index(),
		page,
		cr.GetLimit(),
		query,
		extraBody,
	)
	if !resp.Success {
		res.FailWithMsg(resp.Msg, c)
		return
	}

	data, ok := resp.Data.(map[string]any)
	if !ok {
		res.FailWithMsg("搜索结果格式错误", c)
		return
	}

	total, _ := data["total"].(float64)
	res.OkWithList(extractArticleSearchResults(data, topMap), int(total), c)
}
