package common

import (
	"fmt"
	"myblogx/global"

	"gorm.io/gorm"
)

// TODO: 加时间范围
type PageInfo struct {
	Limit int    `form:"limit"`
	Page  int    `form:"page"`
	Key   string `form:"key"`
	Order string `form:"order"` // 前端可覆盖
}

func (p PageInfo) GetPage(count ...int) int {
	page := p.Page
	if page <= 0 {
		page = 1
	}

	// 兼容旧行为：未传总数时按历史规则限制最大页为20。
	if len(count) == 0 {
		if page > 20 {
			return 1
		}
		return page
	}

	total := count[0]
	if total <= 0 {
		return 1
	}

	limit := p.GetLimit()
	max := (total + limit - 1) / limit
	if page > max {
		return max
	}
	return page
}

func (p PageInfo) GetLimit() int {
	if p.Limit <= 0 || p.Limit > 100 {
		p.Limit = 10
	}
	return p.Limit
}

func (p PageInfo) GetOffset(count ...int) int {
	return (p.GetPage(count...) - 1) * p.GetLimit()
}

type Options struct {
	PageInfo
	Likes        []string
	Preloads     []string
	Where        *gorm.DB
	Debug        bool
	OrderMap     map[string]bool
	DefaultOrder string
}

func ListQuery[T any](model T, option Options) (list []T, count int, err error) {
	// 查询基础
	query := global.DB.Model(model).Where(model)

	// 日志显示
	if option.Debug {
		query = query.Debug()
	}

	// 模糊匹配
	if len(option.Likes) > 0 && option.PageInfo.Key != "" {
		likes := query.Where("")
		for _, column := range option.Likes {
			likes = likes.Or(
				fmt.Sprintf("%s LIKE ?", column),
				"%"+option.PageInfo.Key+"%")
		}
		query = query.Where(likes)
	}

	// 定制化查询
	if option.Where != nil {
		query.Where(option.Where)
	}

	// 查总数
	var _c int64
	query.Count(&_c)
	count = int(_c)

	// 分页
	limit := option.PageInfo.GetLimit()
	offset := option.PageInfo.GetOffset(count)
	query.Limit(limit).Offset(offset)

	// 排序
	if option.PageInfo.Order != "" {
		// 前端已配置排序
		if option.OrderMap != nil && option.OrderMap[option.PageInfo.Order] {
			query = query.Order(option.PageInfo.Order)
		} else {
			err = fmt.Errorf("排序字段错误")
			return
		}
	} else if option.DefaultOrder != "" {
		// 前端未配置排序，使用默认排序
		query = query.Order(option.DefaultOrder)
	}

	// 预加载
	for _, preload := range option.Preloads {
		query = query.Preload(preload)
	}

	err = query.Find(&list).Error
	return
}
