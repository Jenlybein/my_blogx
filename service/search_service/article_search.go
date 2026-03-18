package search_service

import (
	"errors"
	"myblogx/models"
	"myblogx/models/enum"
	"myblogx/service/es_service"
	"myblogx/utils/jwts"
)

func SearchArticles(cr ArticleSearchRequest, claims *jwts.MyClaims) ([]SearchListResponse, int, error) {
	page := cr.Page
	if page <= 0 {
		page = 1
	}

	query := buildDefaultArticleSearchQuery(cr.Key)

	switch cr.Type {
	case 1:
		// 普通搜索使用默认全局搜索即可，这里不追加额外查询条件。
	case 2:
		if claims != nil {
			query = buildLikeTagsQuery(query, claims.UserID)
		}
	case 3:
		query = buildUserIDQuery(query, cr.UserID)
	case 4:
		if claims == nil {
			return nil, 0, errors.New("未登录")
		}
		if cr.Status == enum.ArticleStatusDeleted {
			return nil, 0, errors.New("不能搜索已删除的文章")
		}
		cr.UserID = claims.UserID
		query = buildSelfArticleSearchQuery(cr.Key, claims.UserID, cr.Status)
	case 5:
		if claims == nil || !claims.IsAdmin() {
			return nil, 0, errors.New("权限错误")
		}
		query = buildAdminArticleSearchQuery(cr.Key, cr.Status)
	default:
		return nil, 0, errors.New("搜索类型错误")
	}

	if len(cr.TagList) > 0 {
		query = buildTagListQuery(query, cr.TagList)
	}

	if cr.CategoryID != 0 {
		if cr.Type != 3 && cr.Type != 4 {
			return nil, 0, errors.New("只有作者文章和自己文章支持按分类搜索")
		}
		if cr.UserID == 0 {
			return nil, 0, errors.New("按分类搜索必须传 user_id")
		}
		query = buildCategoryIDQuery(query, cr.CategoryID)
	}

	extraBody := buildArticleSearchExtraBodyBySort(cr.Sort, cr.Key)

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
		return nil, 0, errors.New(resp.Msg)
	}

	data, ok := resp.Data.(map[string]any)
	if !ok {
		return nil, 0, errors.New("搜索结果格式错误")
	}

	total, _ := data["total"].(float64)
	return extractArticleSearchResults(data), int(total), nil
}

func buildArticleSearchExtraBodyBySort(sort int8, key string) map[string]any {
	switch sort {
	case 2:
		return buildArticleSearchExtraBody("created_at", key)
	case 3:
		return buildArticleSearchExtraBody("comment_count", key)
	case 4:
		return buildArticleSearchExtraBody("digg_count", key)
	case 5:
		return buildArticleSearchExtraBody("favor_count", key)
	case 6:
		return buildArticleSearchExtraBody("view_count", key)
	default:
		return buildArticleSearchExtraBody("", key)
	}
}
