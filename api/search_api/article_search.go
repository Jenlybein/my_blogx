package search_api

import (
	"myblogx/common/res"
	"myblogx/middleware"
	"myblogx/models"
	"myblogx/models/enum"
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

	switch cr.Type {
	case 1:
		// 普通搜索使用默认全局搜索即可，这里不追加额外查询条件。
	case 2:
		claims, err := jwts.ParseTokenByGin(c)
		if err != nil || claims == nil {
			break
		}
		query = buildLikeTagsQuery(query, claims.UserID)
	case 3:
		query = buildUserIDQuery(query, cr.UserID)
	case 4:
		claims, err := jwts.ParseTokenByGin(c)
		if err != nil || claims == nil {
			res.FailWithMsg("未登录", c)
			return
		}
		if cr.Status == enum.ArticleStatusDeleted {
			res.FailWithMsg("不能搜索已删除的文章", c)
			return
		}
		cr.UserID = claims.UserID
		query = buildSelfArticleSearchQuery(cr.Key, claims.UserID, cr.Status)
	case 5:
		claims, err := jwts.ParseTokenByGin(c)
		if err != nil || claims == nil || !claims.IsAdmin() {
			res.FailWithMsg("权限错误", c)
			return
		}
		query = buildAdminArticleSearchQuery(cr.Key, cr.Status)
	}

	if len(cr.TagList) > 0 {
		query = buildTagListQuery(query, cr.TagList)
	}

	if cr.CategoryID != 0 {
		if cr.Type != 3 && cr.Type != 4 {
			res.FailWithMsg("只有作者文章和自己文章支持按分类搜索", c)
			return
		}
		if cr.UserID == 0 {
			res.FailWithMsg("按分类搜索必须传 user_id", c)
			return
		}
		query = buildCategoryIDQuery(query, cr.CategoryID)
	}

	var extraBody map[string]any
	switch cr.Sort {
	case 1:
		extraBody = buildArticleSearchExtraBody("", cr.Key)
	case 2:
		extraBody = buildArticleSearchExtraBody("created_at", cr.Key)
	case 3:
		extraBody = buildArticleSearchExtraBody("comment_count", cr.Key)
	case 4:
		extraBody = buildArticleSearchExtraBody("digg_count", cr.Key)
	case 5:
		extraBody = buildArticleSearchExtraBody("favor_count", cr.Key)
	case 6:
		extraBody = buildArticleSearchExtraBody("view_count", cr.Key)
	}

	if cr.TopSearch && (cr.Type == 3 || cr.Type == 4) {
		query = buildAuthorAdminTopQuery(query)
	} else if cr.TopSearch {
		query = buildAdminTopQuery(query)
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
	res.OkWithList(extractArticleSearchResults(data), int(total), c)
}
