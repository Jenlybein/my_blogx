package ai_metainfo

import (
	"errors"
	"fmt"
	"myblogx/global"
	"myblogx/models"
	"myblogx/models/ctype"
	"strings"
)

// GenerateArticleMetainfo 根据文章内容生成标题、摘要、分类和标签建议。
func GenerateArticleMetainfo(uid ctype.ID, content string) (*MetainfoResponse, error) {
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

	// 加载文章分类候选。
	var categoryOptions []Metainfos
	if err := global.DB.Model(&models.CategoryModel{}).
		Where("user_id = ?", uid).
		Order("id asc").
		Select("id", "title").
		Scan(&categoryOptions).Error; err != nil {
		global.Logger.Errorf("查询文章分类候选失败: 用户ID=%d 错误=%v", uid, err)
		return nil, fmt.Errorf("查询分类候选失败: %w", err)
	}

	// 加载文章标签候选。
	var tagOptions []Metainfos
	if err := global.DB.Model(&models.TagModel{}).
		Where("is_enabled = ?", true).
		Order("sort desc, id asc").
		Select("id", "title").
		Scan(&tagOptions).Error; err != nil {
		global.Logger.Errorf("查询文章标签候选失败: 错误=%v", err)
		return nil, fmt.Errorf("查询标签候选失败: %w", err)
	}

	plainText := cleanArticleMetainfoContent(content)
	if plainText == "" {
		return nil, errors.New("文章正文提取结果为空")
	}

	reply, err := requestArticleMetainfoFromAI(plainText, categoryOptions, tagOptions)
	if err != nil {
		return nil, err
	}

	return normalizeArticleMetainfoReply(reply, categoryOptions, tagOptions)
}
