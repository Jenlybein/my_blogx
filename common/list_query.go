package common

import (
	"fmt"
	"myblogx/global"

	"gorm.io/gorm"
)

// TODO: 添加时间范围
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
	baseQuery := buildListQuery(model, option)

	var total int64
	if err = baseQuery.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return
	}
	count = int(total)

	listQuery := baseQuery.Session(&gorm.Session{})
	limit := option.PageInfo.GetLimit()
	offset := option.PageInfo.GetOffset(count)
	listQuery = listQuery.Limit(limit).Offset(offset)

	if option.PageInfo.Order != "" {
		if option.OrderMap == nil || !option.OrderMap[option.PageInfo.Order] {
			err = fmt.Errorf("排序字段错误")
			return
		}
		listQuery = listQuery.Order(option.PageInfo.Order)
	} else if option.DefaultOrder != "" {
		listQuery = listQuery.Order(option.DefaultOrder)
	}

	// Select 只影响列表查询，不影响 count。
	if len(option.Select) > 0 {
		listQuery = listQuery.Select(option.Select)
	}

	for _, preload := range option.Preloads {
		listQuery = listQuery.Preload(preload)
	}

	for preload, fields := range option.ExactPreloads {
		listQuery = listQuery.Preload(preload, func(db *gorm.DB) *gorm.DB {
			return db.Select(fields)
		})
	}

	err = listQuery.Find(&list).Error
	return
}

func buildListQuery[T any](model T, option Options) *gorm.DB {
	query := global.DB.Model(model).Where(model)

	if option.Debug {
		query = query.Debug()
	}

	if len(option.Likes) > 0 && option.PageInfo.Key != "" {
		query = query.Where(buildLikeCondition(option.Likes, option.PageInfo.Key))
	}

	// Where 用于追加额外过滤条件，不建议传入完整查询对象。
	if option.Where != nil {
		query = query.Where(option.Where)
	}

	return query
}

func buildLikeCondition(columns []string, key string) *gorm.DB {
	pattern := "%" + key + "%"
	likeQuery := global.DB.Where(fmt.Sprintf("%s LIKE ?", columns[0]), pattern)
	for _, column := range columns[1:] {
		likeQuery = likeQuery.Or(fmt.Sprintf("%s LIKE ?", column), pattern)
	}
	return likeQuery
}
