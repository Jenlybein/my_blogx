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

	// 兼容历史行为：不传总数时最大页限制为 20。
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
	Select        []string
	Likes         []string
	Preloads      []string
	ExactPreloads map[string][]string
	Where         *gorm.DB
	Debug         bool
	OrderMap      map[string]bool
	DefaultOrder  string
}

func ListQuery[T any](model T, option Options) (list []T, count int, err error) {
	query := global.DB.Model(model).Where(model)

	if option.Debug {
		query = query.Debug()
	}

	if len(option.Likes) > 0 && option.PageInfo.Key != "" {
		likes := query.Where("")
		for _, column := range option.Likes {
			likes = likes.Or(
				fmt.Sprintf("%s LIKE ?", column),
				"%"+option.PageInfo.Key+"%",
			)
		}
		query = query.Where(likes)
	}

	if option.Where != nil {
		query = query.Where(option.Where)
	}

	var _c int64
	query.Count(&_c)
	count = int(_c)

	limit := option.PageInfo.GetLimit()
	offset := option.PageInfo.GetOffset(count)
	query = query.Limit(limit).Offset(offset)

	if option.PageInfo.Order != "" {
		if option.OrderMap != nil && option.OrderMap[option.PageInfo.Order] {
			query = query.Order(option.PageInfo.Order)
		} else {
			err = fmt.Errorf("排序字段错误")
			return
		}
	} else if option.DefaultOrder != "" {
		query = query.Order(option.DefaultOrder)
	}

	// Select 只影响列表查询，不影响 count。
	if len(option.Select) > 0 {
		query = query.Select(option.Select)
	}

	for _, preload := range option.Preloads {
		query = query.Preload(preload)
	}

	for preload, fields := range option.ExactPreloads {
		query = query.Preload(preload, func(db *gorm.DB) *gorm.DB {
			return db.Select(fields)
		})
	}

	err = query.Find(&list).Error
	return
}
