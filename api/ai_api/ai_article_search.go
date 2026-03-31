package ai_api

import (
	"myblogx/common"
	"myblogx/common/res"
	"myblogx/middleware"
	"myblogx/models/ctype"
	"myblogx/service/ai_service/ai_search"
	"myblogx/service/search_service"
	"strings"

	"github.com/gin-gonic/gin"
)

func (AIApi) AIArticleSearchListView(c *gin.Context) {
	cr := middleware.GetBindJson[AIBaseRequest](c)

	rewrite, err := ai_search.RewriteArticleSearch(cr.Content)
	if err != nil {
		res.FailWithMsg(err.Error(), c)
		return
	}
	if rewrite.Intent != "search" {
		res.FailWithMsg("当前输入不是文章搜索意图", c)
		return
	}

	// fmt.Println(rewrite)

	key := buildAIArticleSearchKey(rewrite.Query)
	if key == "" {
		res.FailWithMsg("搜索关键词不能为空", c)
		return
	}

	list, err := searchAIArticleList(rewrite)
	if err != nil {
		res.FailWithMsg(err.Error(), c)
		return
	}

	res.OkWithData(list, c)
}

func (AIApi) AIArticleSearchLLMView(c *gin.Context) {
	cr := middleware.GetBindJson[AIBaseRequest](c)
	prepareAISSE(c)

	// 意图识别与搜索重写
	rewrite, err := ai_search.RewriteArticleSearch(cr.Content)
	if err != nil {
		res.SSEFail(err.Error(), c)
		return
	}

	// 非搜索意图，直接回复
	if rewrite.Intent != "search" {
		res.SSEOk(AIBaseResponse{
			Content: rewrite.Content,
		}, c)
		return
	}

	// 搜索意图，进行搜索
	key := buildAIArticleSearchKey(rewrite.Query)
	if key == "" {
		res.SSEFail("搜索关键词不能为空", c)
		return
	}

	list, err := searchAIArticleList(rewrite)
	if err != nil {
		res.SSEFail(err.Error(), c)
		return
	}

	contentChan, errChan, err := ai_search.AnalyzeArticleSearchStream(cr.Content, list)
	if err != nil {
		res.SSEFail(err.Error(), c)
		return
	}

	for contentChan != nil || errChan != nil {
		select {
		// 接收消息
		case text, ok := <-contentChan:
			if !ok {
				contentChan = nil
				continue
			}
			res.SSEOk(AIBaseResponse{
				Content: text,
			}, c)
		// 接收错误
		case streamErr, ok := <-errChan:
			if !ok {
				errChan = nil
				continue
			}
			if streamErr != nil {
				res.SSEFail(streamErr.Error(), c)
				return
			}
		}
	}
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
	seen map[ctype.ID]struct{},
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

func searchAIArticleList(rewrite *ai_search.ArticleSearchRewrite) ([]search_service.SearchListResponse, error) {
	key := buildAIArticleSearchKey(rewrite.Query)
	if key == "" {
		return nil, nil
	}

	list := make([]search_service.SearchListResponse, 0, 20)
	seen := make(map[ctype.ID]struct{}, 20)

	if len(rewrite.TagList) > 0 {
		tagList, _, err := search_service.SearchArticles(search_service.ArticleSearchRequest{
			Type:     1,
			Sort:     rewrite.Sort,
			TagList:  rewrite.TagList,
			Key:      key,
			PageInfo: common.PageInfo{Page: 1, Limit: 10},
		}, nil)
		if err != nil {
			return nil, err
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
		return nil, err
	}
	list = appendUniqueSearchResults(list, seen, queryList)

	return list, nil
}
