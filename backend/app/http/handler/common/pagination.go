package common

import (
	"github.com/ngq/gorp"
)

const (
	defaultPage     = 1
	defaultPageSize = 20
	defaultMaxSize  = 100
)

type Pagination struct {
	Page     int `form:"page" json:"page"`
	PageSize int `form:"page_size" json:"page_size"`
}

func (p Pagination) Offset() int {
	return (p.Page - 1) * p.PageSize
}

// ParsePagination 从请求上下文中解析分页参数 page 和 page_size，
// 强制约束：page >= 1, 1 <= pageSize <= maxSize（默认上限 100）。
// defSize: 默认每页大小（传 0 使用默认值 20）
// maxSize: 最大每页大小（传 0 使用默认上限 100）
func ParsePagination(c gorp.Context, defSize, maxSize int) Pagination {
	if defSize <= 0 {
		defSize = defaultPageSize
	}
	if maxSize <= 0 {
		maxSize = defaultMaxSize
	}

	page := c.DefaultIntQuery("page", defaultPage)
	if page < 1 {
		page = 1
	}

	pageSize := c.DefaultIntQuery("page_size", defSize)
	if pageSize < 1 {
		pageSize = defSize
	}
	if pageSize > maxSize {
		pageSize = maxSize
	}

	return Pagination{Page: page, PageSize: pageSize}
}
