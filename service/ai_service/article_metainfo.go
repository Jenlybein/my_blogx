package ai_service

import (
	"encoding/json"
	"errors"
	"fmt"
	"myblogx/global"
	"myblogx/models"
	"myblogx/utils/markdown"
	"strings"
)

const (
	articleMetainfoTitleLimit    = 30
	articleMetainfoAbstractLimit = 200
	articleMetainfoMaxTags       = 3
)

var (
	articleMetainfoPrompt = `
你是一个帮助分析文章元信息的人工智能助手。你必须基于用户提供的文章内容，只输出合法 JSON，不要输出 Markdown 代码块、解释或补充说明。

请完成以下任务：
1. 生成一个精炼且贴合原文的标题，30 字以内。
2. 生成一个精炼且贴合原文的摘要，200 字以内。
3. 从已有分类中选择 0 或 1 个最合适的分类；如果没有合适分类或没有可用分类，返回 null。已有分类：%s
4. 从已有标签中选择 0~3 个最合适的标签；如果没有合适标签或没有可用标签，返回空数组。已有标签：%s
5. 分类和标签只能从给定候选里选择，不能编造 id 或 title。
6. 严格按照下面的 JSON 结构输出：
{
  "title": "",
  "abstract": "",
  "category": {"id": 1, "title": ""},
  "tags": [{"id": 1, "title": ""}]
}
`
)

// ArticleMetainfo 是文章元信息推荐时使用的候选项。
type Metainfos struct {
	ID    uint   `json:"id"`
	Title string `json:"title"`
}

// MetainfoRequest 是生成文章元信息时的请求参数。
type MetainfoRequest struct {
	UserID  uint   `json:"user_id"`
	Content string `json:"content"`
}

// MetainfoResponse 是 AI 生成并校验后的文章元信息。
type MetainfoResponse struct {
	Title    string      `json:"title"`
	Abstract string      `json:"abstract"`
	Category *Metainfos  `json:"category"`
	Tags     []Metainfos `json:"tags"`
}

// GenerateArticleMetainfo 根据文章内容生成标题、摘要、分类和标签建议。
func GenerateArticleMetainfo(uid uint, content string) (*MetainfoResponse, error) {
	if uid == 0 {
		return nil, errors.New("用户 ID 不能为空")
	}
	if strings.TrimSpace(content) == "" {
		return nil, errors.New("文章内容不能为空")
	}
	if global.Config == nil {
		return nil, errors.New("系统配置未初始化")
	}
	if global.DB == nil {
		return nil, errors.New("数据库未初始化")
	}

	// 加载文章元信息候选
	var categoryOptions []Metainfos
	if err := global.DB.Model(&models.CategoryModel{}).Where("user_id = ?", uid).Order("id asc").
		Select("id", "title").Scan(&categoryOptions).Error; err != nil {
		global.Logger.Errorf("查询文章分类候选失败 user_id=%d err=%v", uid, err)
		return nil, fmt.Errorf("查询分类候选失败: %w", err)
	}

	// 加载文章标签候选
	var tagOptions []Metainfos
	if err := global.DB.Model(&models.TagModel{}).Where("is_enabled = ?", true).Order("sort desc, id asc").Select("id", "title").Scan(&tagOptions).Error; err != nil {
		global.Logger.Errorf("查询文章标签候选失败 err=%v", err)
		return nil, fmt.Errorf("查询标签候选失败: %w", err)
	}

	// 提取文章正文
	plainText := cleanArticleMetainfoContent(content)
	if plainText == "" {
		return nil, errors.New("文章正文提取结果为空")
	}

	// 生成文章元信息
	reply, err := requestArticleMetainfoFromAI(plainText, categoryOptions, tagOptions)
	if err != nil {
		return nil, err
	}

	return normalizeArticleMetainfoReply(reply, categoryOptions, tagOptions)
}

// 清洗文章正文，去除全部格式，保留最多 4096 个字符
func cleanArticleMetainfoContent(content string) string {
	text := markdown.MdToTextParagraph(content)

	maxChars := global.Config.AI.MaxInputChars
	if maxChars <= 0 {
		maxChars = 4096
	}

	return markdown.ExtractText(text, maxChars)
}

func requestArticleMetainfoFromAI(
	article string,
	categoryOptions []Metainfos,
	tagOptions []Metainfos,
) (string, error) {
	// 生成 prompt
	prompt := fmt.Sprintf(articleMetainfoPrompt, mustJSONString(categoryOptions), mustJSONString(tagOptions))

	// 发起请求
	msgList := []Message{
		{
			Role:    "system",
			Content: prompt,
		},
		{
			Role:    "user",
			Content: article,
		},
	}

	reply, err := Chat(msgList)
	if err != nil {
		return "", fmt.Errorf("文章元信息请求失败: %w", err)
	}
	return reply, nil
}

func normalizeArticleMetainfoReply(
	raw string,
	categoryOptions []Metainfos,
	tagOptions []Metainfos,
) (*MetainfoResponse, error) {
	// 从回答中提取 Json
	start := strings.Index(raw, "{")
	end := strings.LastIndex(raw, "}")
	if start >= 0 && end > start {
		raw = raw[start : end+1]
	}

	var payload MetainfoResponse
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		global.Logger.Errorf("解析文章元信息 JSON 失败 err=%v raw=%s", err, raw)
		return nil, fmt.Errorf("文章元信息结果不是有效 JSON: %w", err)
	}

	result := &MetainfoResponse{
		Title:    payload.Title,
		Abstract: payload.Abstract,
		Category: nil,
	}

	// 校验返回的 Tags ID 是否存在
	var validTags []Metainfos
	tagMap := make(map[uint]struct{}, len(tagOptions))
	for _, tag := range tagOptions {
		tagMap[tag.ID] = struct{}{}
	}
	for _, tag := range payload.Tags {
		if _, exists := tagMap[tag.ID]; exists {
			validTags = append(validTags, tag)
		}
	}
	result.Tags = validTags

	// 校验返回的 Category
	if payload.Category != nil {
		for _, category := range categoryOptions {
			if category.ID == payload.Category.ID {
				result.Category = &category
				break
			}
		}
	}

	return result, nil
}


