# AI 助力文章搜索设计方案

## 1. 目标

在现有 `ES + 关键词搜索` 的基础上，增加一层 AI 能力，让用户可以直接输入自然语言问题，例如：

- “有没有讲 Gin 中间件和 JWT 的文章”
- “帮我找作者写过的 ES 同步相关文章”
- “最近有没有 Go 缓存一致性的文章”

系统自动完成：

1. 理解用户真实搜索意图
2. 将自然语言改写成当前搜索系统能执行的结构化条件
3. 复用现有 ES 搜索逻辑召回文章
4. 按需对候选结果做 AI 重排
5. 按需给出一句简短总结

本方案优先保证：

- 不过度解耦
- 尽量复用现有 `api/search_api` 和 `service/es_service`
- 第一版先做“AI 改写搜索词 + 复用现有 ES 搜索”
- 第二版再做“AI 重排”
- 第三版再做“AI 搜索总结”


## 2. 当前项目基础

仓库里已经有足够好的搜索基础，不需要推倒重来。

### 2.1 当前搜索入口

- [api/search_api/article_search.go](./api/search_api/article_search.go)
- [router/search_router.go](./router/search_router.go)

当前已有：

- 普通文章搜索
- 猜你喜欢
- 作者文章搜索
- 自己文章搜索
- 管理员搜索
- 标签过滤
- 分类过滤
- 置顶优先
- 多种排序

### 2.2 当前 ES 可搜索字段

- [service/es_service/article_sync.go](./service/es_service/article_sync.go)

当前文章写入 ES 的字段已经包括：

- `title`
- `abstract`
- `content_parts`
- `content_head`
- `tags`
- `category_id`
- `author_id`
- `status`
- `created_at`
- `view_count`
- `digg_count`
- `comment_count`
- `favor_count`
- `admin_top`
- `author_top`

这意味着第一版 AI 搜索不需要新增索引，不需要单独上向量库，也不需要改 ES mapping。

### 2.3 当前 AI 基础能力

- [service/ai_service/enter.go](./service/ai_service/enter.go)

当前已经有：

- `Message`
- `Request`
- `Chat(msgList []Message)`
- `ChatStream(msgList []Message)`

所以 AI 搜索第一版只需要新增一个“搜索改写”能力文件即可。


## 3. 设计原则

### 3.1 不新增没必要的抽象层

这块不要为了“优雅”额外造：

- `SearchDomainService`
- `SearchRepository`
- `AIOrchestrator`
- `PromptManager`

第一版保持直接：

- API 层放在 `api/search_api`
- AI 改写逻辑放在 `service/ai_service/article_search.go`
- ES 查询继续复用 `api/search_api/article_search_utils_build.go`

### 3.2 AI 不直接拼 ES DSL

AI 只负责输出我们定义好的 JSON 字段，例如：

```json
{
  "query": "Gin 中间件 JWT",
  "tag_list": ["Gin", "JWT"],
  "category_id": 0,
  "sort": 1,
  "need_summary": false
}
```

然后服务端自己调用现有 `buildDefaultArticleSearchQuery`、`buildTagListQuery`、`buildCategoryIDQuery` 等函数。

这样好处很直接：

- 更稳
- 更可控
- 更好测
- 不会把 ES 语法暴露给 AI

### 3.3 所有 AI 输出都要二次校验

尤其是：

- 分类 ID
- 标签列表
- 排序值

必须服务端再校验一遍，绝不能直接信任 AI 输出。


## 4. 第一版范围

第一版只做：

1. 用户输入自然语言
2. AI 把自然语言改写成结构化搜索条件
3. 复用当前 ES 搜索
4. 返回搜索结果
5. 可选返回 AI 改写结果，方便前端展示“已为你理解为：xxx”

第一版先不做：

- 向量搜索
- embedding 存储
- Agent 多轮检索
- 站内答案生成
- 流式输出


## 5. 接口设计

### 5.1 新增接口

建议新增：

- `POST /api/search/ai_articles`

放在现有 `search` 下面，而不是 `ai` 下面。原因很简单：

- 这是搜索能力，不是通用 AI 聊天能力
- 能直接复用 `search_api` 现有上下文
- 路由更清晰

### 5.2 请求结构

文件：

- [api/search_api/search_types.go](./api/search_api/search_types.go)

新增：

```go
type AIArticleSearchRequest struct {
	common.PageInfo
	Query      string `json:"query" binding:"required"`
	Type       int8   `json:"type" binding:"omitempty,oneof=1 2 3 4 5"`
	UserID     uint   `json:"user_id"`
	TopSearch  bool   `json:"top_search"`
	NeedRerank bool   `json:"need_rerank"`
	NeedSummary bool  `json:"need_summary"`
}
```

说明：

- `Query` 是用户自然语言输入
- `Type` 含义继续沿用当前文章搜索
- `UserID` 只在“作者文章搜索”时使用
- `NeedRerank` 第一版可以先保留字段但默认不启用
- `NeedSummary` 第一版可以先支持 `false`

### 5.3 响应结构

新增：

```go
type AIArticleSearchRewriteResponse struct {
	Query      string   `json:"query"`
	TagList    []string `json:"tag_list"`
	CategoryID uint     `json:"category_id"`
	Sort       int8     `json:"sort"`
}

type AIArticleSearchResponse struct {
	Rewrite AIArticleSearchRewriteResponse `json:"rewrite"`
	List    []SearchListResponse          `json:"list"`
	Count   int                           `json:"count"`
	Summary string                        `json:"summary,omitempty"`
}
```


## 6. API 层实现

### 6.1 新文件

新增文件：

- `api/search_api/article_ai_search.go`

实现一个新方法：

```go
func (SearchApi) AIArticleSearchView(c *gin.Context)
```

### 6.2 处理流程

代码流程建议如下：

```go
func (SearchApi) AIArticleSearchView(c *gin.Context) {
	cr := middleware.GetBindJson[AIArticleSearchRequest](c)

	page := cr.Page
	if page <= 0 {
		page = 1
	}

	claims, _ := jwts.ParseTokenByGin(c)

	rewrite, err := ai_service.RewriteArticleSearch(ai_service.RewriteArticleSearchRequest{
		UserID:      getRewriteUserID(cr, claims),
		Query:       cr.Query,
		SearchType:  cr.Type,
		UserIDScope: cr.UserID,
	})
	if err != nil {
		res.FailWithMsg(err.Error(), c)
		return
	}

	query, extraBody, err := buildAIArticleSearchESQuery(cr, rewrite, claims)
	if err != nil {
		res.FailWithMsg(err.Error(), c)
		return
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

	list := extractArticleSearchResults(data)
	total, _ := data["total"].(float64)

	res.OkWithData(AIArticleSearchResponse{
		Rewrite: AIArticleSearchRewriteResponse{
			Query:      rewrite.Query,
			TagList:    rewrite.TagList,
			CategoryID: rewrite.CategoryID,
			Sort:       rewrite.Sort,
		},
		List:  list,
		Count: int(total),
	}, c)
}
```

### 6.3 不要重复造查询构造器

这里不要再新建一套 `AIQueryBuilder`。

直接复用当前已有的：

- `buildDefaultArticleSearchQuery`
- `buildSelfArticleSearchQuery`
- `buildAdminArticleSearchQuery`
- `buildUserIDQuery`
- `buildTagListQuery`
- `buildCategoryIDQuery`
- `buildArticleSearchExtraBody`
- `buildAuthorAdminTopQuery`
- `buildAdminTopQuery`

只需要在 `api/search_api` 内部补一个辅助函数：

```go
func buildAIArticleSearchESQuery(
	cr AIArticleSearchRequest,
	rewrite ai_service.ArticleSearchRewrite,
	claims *jwts.MyClaims,
) (map[string]any, map[string]any, error)
```

这个函数只做两件事：

1. 把 AI 输出转成现有搜索条件
2. 调用现有搜索构造函数


## 7. AI 服务层实现

### 7.1 新文件

新增文件：

- `service/ai_service/article_search.go`

### 7.2 结构体定义

建议定义：

```go
type ArticleSearchOption struct {
	ID    uint   `json:"id"`
	Title string `json:"title"`
}

type RewriteArticleSearchRequest struct {
	UserID      uint   `json:"user_id"`
	Query       string `json:"query"`
	SearchType  int8   `json:"search_type"`
	UserIDScope uint   `json:"user_id_scope"`
}

type ArticleSearchRewrite struct {
	Query      string   `json:"query"`
	TagList    []string `json:"tag_list"`
	CategoryID uint     `json:"category_id"`
	Sort       int8     `json:"sort"`
	NeedSummary bool    `json:"need_summary"`
}
```

### 7.3 Prompt 设计

建议在文件里直接放一个 `var articleSearchRewritePrompt`，不要单独抽 prompt 包。

示例：

```go
var articleSearchRewritePrompt = `
你是一个博客站内文章搜索助手。
你需要把用户的自然语言搜索请求，转换为文章搜索参数。

规则：
1. 只输出合法 JSON，不要输出解释。
2. query 字段必须是适合全文检索的核心关键词短语。
3. tag_list 只能从给定标签中选，最多 3 个。
4. category_id 只能从给定分类中选；没有合适分类就填 0。
5. sort 只能是 1~6：
   1 默认相关度
   2 最新发布
   3 最多回复
   4 最多点赞
   5 最多收藏
   6 最多浏览
6. 如果用户没有明显表达排序偏好，sort 返回 1。
7. 严格输出：
{
  "query": "",
  "tag_list": [],
  "category_id": 0,
  "sort": 1,
  "need_summary": false
}

分类候选：%s
标签候选：%s
`
```

### 7.4 分类和标签候选加载

和当前 `article_metainfo` 保持同风格，不要再新开 repository 层。

直接在 `article_search.go` 里写两个小函数：

```go
func loadArticleSearchCategoryOptions(userID uint) ([]ArticleSearchOption, error)
func loadArticleSearchTagOptions() ([]ArticleSearchOption, error)
```

查询逻辑：

- 分类：按 `user_id = ?`
- 标签：按 `is_enabled = true`

### 7.5 AI 改写主函数

核心函数：

```go
func RewriteArticleSearch(req RewriteArticleSearchRequest) (*ArticleSearchRewrite, error)
```

建议流程：

1. 校验 `req.Query`
2. 加载分类候选
3. 加载标签候选
4. 用 `[]Message` 调 `Chat`
5. 提取 JSON
6. 二次校验结果
7. 返回结构化改写结果

伪代码：

```go
func RewriteArticleSearch(req RewriteArticleSearchRequest) (*ArticleSearchRewrite, error) {
	if strings.TrimSpace(req.Query) == "" {
		return nil, errors.New("搜索内容不能为空")
	}

	categoryOptions, err := loadArticleSearchCategoryOptions(req.UserID)
	if err != nil {
		return nil, err
	}
	tagOptions, err := loadArticleSearchTagOptions()
	if err != nil {
		return nil, err
	}

	msgList := []Message{
		{
			Role: "system",
			Content: fmt.Sprintf(
				articleSearchRewritePrompt,
				mustJSONString(categoryOptions),
				mustJSONString(tagOptions),
			),
		},
		{
			Role:    "user",
			Content: req.Query,
		},
	}

	reply, err := Chat(msgList)
	if err != nil {
		return nil, err
	}

	return normalizeArticleSearchRewrite(reply, categoryOptions, tagOptions)
}
```

### 7.6 AI 输出校验

这里必须做二次校验。

建议实现：

```go
func normalizeArticleSearchRewrite(
	raw string,
	categoryOptions []ArticleSearchOption,
	tagOptions []ArticleSearchOption,
) (*ArticleSearchRewrite, error)
```

校验规则：

1. `query` 去空格，不能为空
2. `sort` 不在 1~6 之间时，回退为 1
3. `category_id` 不在候选列表中时，回退为 0
4. `tag_list` 只保留候选标签中的标题，最多 3 个
5. 标签去重

注意：

- 标签建议按 `Title` 匹配即可，不需要再引入标签 ID 到搜索响应里
- 分类可以直接用 ID，最终还是走 `buildCategoryIDQuery`


## 8. ES 查询复用方案

### 8.1 查询构造策略

AI 改写完成后，不要新写一份 ES DSL。

直接复用当前搜索逻辑：

```go
query := buildDefaultArticleSearchQuery(rewrite.Query)

if len(rewrite.TagList) > 0 {
	query = buildTagListQuery(query, rewrite.TagList)
}
if rewrite.CategoryID != 0 {
	query = buildCategoryIDQuery(query, rewrite.CategoryID)
}

extraBody := buildArticleSearchExtraBody(sortFieldByAI(rewrite.Sort), rewrite.Query)
```

### 8.2 排序映射

补一个极小的辅助函数就够了：

```go
func buildAIArticleSearchExtraBody(sort int8, key string) map[string]any {
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
```


## 9. 第二版：AI 重排

第一版上线稳定后再做。

### 9.1 目标

解决“关键词都命中了，但排序不够像人理解”的问题。

例如：

- 用户搜“Gin 中间件鉴权”
- ES 会召回很多 Gin 文章
- AI 重排后，把“真正讲鉴权中间件”的放前面

### 9.2 做法

1. 先用 ES 查出前 20 篇
2. 抽每篇文章的精简信息
3. 交给 AI 返回排序后的文章 ID 列表
4. 服务端按返回顺序重组列表

### 9.3 候选输入结构

建议喂给 AI 的候选项：

```json
[
  {
    "id": 12,
    "title": "Gin 中间件设计实战",
    "abstract": "介绍日志、鉴权和恢复中间件写法",
    "tags": ["Go", "Gin", "中间件"],
    "content_head": "本文围绕 Gin 中间件链路展开..."
  }
]
```

### 9.4 重排输出格式

```json
{
  "ids": [12, 3, 9, 20]
}
```

### 9.5 新增函数

文件仍然放在：

- `service/ai_service/article_search.go`

新增：

```go
func RerankArticleSearch(query string, list []ArticleSearchRerankItem) ([]uint, error)
```

注意这里仍然不要单独再拆新 service。


## 10. 第三版：AI 搜索总结

### 10.1 目标

对搜索结果做一段非常短的总结，适合展示在列表顶部。

例如：

- “找到 6 篇和 Gin 中间件相关的文章，主要集中在鉴权、日志和异常恢复。”

### 10.2 触发时机

只有在：

- 用户输入明显是问句
- 且结果数 > 0
- 且 `need_summary = true`

时才调用。

### 10.3 新增函数

```go
func SummarizeArticleSearch(query string, list []ArticleSearchRerankItem) (string, error)
```


## 11. 建议的具体改动文件

第一版落地时，建议只改这些文件。

### 11.1 API 层

- 新增 `api/search_api/article_ai_search.go`
- 修改 `api/search_api/search_types.go`
- 可选修改 `api/search_api/enter.go`，如果需要新增注释

### 11.2 Router 层

- 修改 `router/search_router.go`

新增路由：

```go
group.POST("ai_articles", mw.BindJson[search_api.AIArticleSearchRequest], app.AIArticleSearchView)
```

### 11.3 AI 服务层

- 新增 `service/ai_service/article_search.go`

### 11.4 测试

- 新增 `service/ai_service/article_search_test.go`
- 新增 `api/search_api/article_ai_search_test.go`


## 12. 具体测试方案

### 12.1 `service/ai_service/article_search_test.go`

至少覆盖：

1. AI 返回合法 JSON，能正确解析
2. AI 返回多余文字，能正确提取 JSON
3. AI 返回不存在的分类 ID，被回退为 0
4. AI 返回不存在的标签，被过滤
5. AI 返回重复标签，被去重
6. AI 返回非法排序值，被回退为 1

### 12.2 `api/search_api/article_ai_search_test.go`

使用：

- sqlite
- 假 AI HTTP 服务
- 假 ES 搜索响应或直接复用当前 ES 测试方案

至少覆盖：

1. 正常搜索成功
2. query 为空时失败
3. AI 改写失败时失败
4. ES 查询失败时失败
5. AI 返回非法分类/标签时仍能成功搜索


## 13. 失败降级策略

### 13.1 第一版建议

AI 改写失败时，直接降级到普通搜索，而不是整个接口报错。

建议逻辑：

```go
rewrite, err := ai_service.RewriteArticleSearch(...)
if err != nil {
	rewrite = &ai_service.ArticleSearchRewrite{
		Query:      strings.TrimSpace(cr.Query),
		TagList:    nil,
		CategoryID: 0,
		Sort:       1,
	}
}
```

这样 AI 挂了，搜索功能仍然能用。

### 13.2 是否把错误返回给前端

建议第一版不要把 AI 内部错误直接暴露给前端。

前端只需要知道：

- 当前是 AI 理解后的结果
- 或本次已降级为普通搜索

可以加一个响应字段：

```go
Fallback bool `json:"fallback"`
```


## 14. 日志与观测

建议在 API 层和 AI 服务层增加必要日志。

### 14.1 记录改写结果

```go
global.Logger.Infof(
	"AI文章搜索改写 query=%q rewrite_query=%q tag_list=%v category_id=%d sort=%d",
	cr.Query,
	rewrite.Query,
	rewrite.TagList,
	rewrite.CategoryID,
	rewrite.Sort,
)
```

### 14.2 不要记录全文候选内容

避免日志过大，也避免泄漏敏感内容。


## 15. 第二阶段再考虑的增强项

这些不是第一版必须做，但后面可以逐步演进。

### 15.1 ES 向量检索

如果后面要做真正的语义搜索，可以在文章索引增加：

- `title_vector`
- `abstract_vector`
- `content_head_vector`

然后在 ES 里做 `knn + bool filter` 混合召回。

但这一步只有在当前“AI 改写 + 关键词检索”效果不够时再做。

### 15.2 用户搜索历史

可记录：

- 原始 query
- AI 改写后的 query
- 最终点击的 article_id

后续可做：

- 热门搜索词
- 搜索改写优化
- 个性化推荐


## 16. 推荐落地顺序

### 第 1 天

1. 新增 `service/ai_service/article_search.go`
2. 完成 AI 改写逻辑
3. 补 service 层测试

### 第 2 天

1. 新增 `api/search_api/article_ai_search.go`
2. 接上路由
3. 补 API 层测试

### 第 3 天

1. 增加失败降级
2. 加日志
3. 调 prompt

### 第 4 天

1. 观察搜索日志
2. 再决定是否做 AI 重排


## 17. 最终推荐结论

最适合你当前仓库的实现，不是直接上重型 RAG，而是：

1. 先做 `AI 改写自然语言搜索词`
2. 复用现有 ES 搜索逻辑
3. 稳定后再做 `AI 重排`
4. 最后再做 `AI 总结`

这样有几个明显优点：

- 和你现有代码最契合
- 改动小
- 成本低
- 可控性强
- 容易测试
- 不会因为“AI 搜索”把整个搜索体系复杂化


## 18. 第一版最终建议的文件清单

建议最终新增或修改：

- `api/search_api/search_types.go`
- `api/search_api/article_ai_search.go`
- `router/search_router.go`
- `service/ai_service/article_search.go`
- `service/ai_service/article_search_test.go`
- `api/search_api/article_ai_search_test.go`

如果只按这个范围做，已经足够上线第一版 AI 文章搜索。
