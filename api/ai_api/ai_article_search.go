package ai_api

import (
	"fmt"
	"myblogx/common"
	"myblogx/common/res"
	"myblogx/middleware"
	"myblogx/service/ai_service"
	"myblogx/service/search_service"
	"strings"

	"github.com/gin-gonic/gin"
)

func (AIApi) AIArticleSearchView(c *gin.Context) {
	cr := middleware.GetBindJson[AIArticleMetaInfoRequest](c)

	rewrite, err := ai_service.RewriteArticleSearch(cr.Content)
	if err != nil {
		res.FailWithMsg(err.Error(), c)
		return
	}
	if rewrite.Intent != "search" {
		res.FailWithMsg("当前输入不是文章搜索意图", c)
		return
	}

	fmt.Println(rewrite)

	key := buildAIArticleSearchKey(rewrite.Query)
	if key == "" {
		res.FailWithMsg("搜索关键词不能为空", c)
		return
	}

	list := make([]search_service.SearchListResponse, 0, 20)
	seen := make(map[uint]struct{}, 20)

	if len(rewrite.TagList) > 0 {
		tagList, _, err := search_service.SearchArticles(search_service.ArticleSearchRequest{
			Type:     1,
			Sort:     rewrite.Sort,
			TagList:  rewrite.TagList,
			Key:      key,
			PageInfo: common.PageInfo{Page: 1, Limit: 10},
		}, nil)
		if err != nil {
			res.FailWithMsg(err.Error(), c)
			return
		}
		list = appendUniqueSearchResults(list, seen, tagList)
	}

	queryList, _, err := search_service.SearchArticles(search_service.ArticleSearchRequest{
		Type:     1,
		Sort:     rewrite.Sort,
		Key:      key,
		PageInfo: common.PageInfo{Page: 1, Limit: 10},
	}, nil)
	if err != nil {
		res.FailWithMsg(err.Error(), c)
		return
	}
	list = appendUniqueSearchResults(list, seen, queryList)

	res.OkWithData(list, c)
}

func buildAIArticleSearchKey(queryList []string) string {
	result := ""
	for _, item := range queryList {
		item = strings.Join(strings.Fields(strings.TrimSpace(item)), " ")
		if item == "" {
			continue
		}
		if result == "" {
			result = item
			continue
		}
		result += " " + item
	}
	return result
}

func appendUniqueSearchResults(
	list []search_service.SearchListResponse,
	seen map[uint]struct{},
	appendList []search_service.SearchListResponse,
) []search_service.SearchListResponse {
	for _, item := range appendList {
		if _, ok := seen[item.ID]; ok {
			continue
		}
		seen[item.ID] = struct{}{}
		list = append(list, item)
	}
	return list
}
