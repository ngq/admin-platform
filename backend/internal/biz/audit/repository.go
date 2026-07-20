// Package audit 定义审计日志领域的仓储接口。
//
// 中文说明：
// - 操作日志、登录日志、错误日志各自有独立的仓储接口；
// - 仓储接口由 data 层实现，biz 层通过接口解耦。
package audit

import (
	"context"
	"time"
)

// OperationLogRepo 操作日志仓储接口
type OperationLogRepo interface {
	// Create 创建操作日志
	Create(ctx context.Context, log *OperationLog) error
	// List 分页查询操作日志列表
	List(ctx context.Context, page, pageSize int, username, method, path string, startTime, endTime *time.Time) ([]*OperationLog, int64, error)
	// Count 查询操作日志总数（用于仪表盘统计，避免 List 的额外开销）
	Count(ctx context.Context, startTime, endTime *time.Time) (int64, error)
	// CountDistinctUsers 查询指定时间段内不同用户数（在线用户统计）
	CountDistinctUsers(ctx context.Context, since time.Time) (int64, error)
	// Delete 删除操作日志
	Delete(ctx context.Context, id int64) error
	// BatchDelete 批量删除操作日志
	BatchDelete(ctx context.Context, ids []int64) error
}

// LoginLogRepo 登录日志仓储接口
type LoginLogRepo interface {
	// Create 创建登录日志
	Create(ctx context.Context, log *LoginLog) error
	// List 分页查询登录日志列表
	List(ctx context.Context, page, pageSize int, username, ip string, status int8, startTime, endTime *time.Time) ([]*LoginLog, int64, error)
	// CountDistinctUsers 查询指定时间段内不同用户数（在线用户统计）
	CountDistinctUsers(ctx context.Context, since time.Time) (int64, error)
	// Delete 删除登录日志
	Delete(ctx context.Context, id int64) error
	// BatchDelete 批量删除登录日志
	BatchDelete(ctx context.Context, ids []int64) error
}

// ErrorLogRepo 错误日志仓储接口
type ErrorLogRepo interface {
	// Create 创建错误日志
	Create(ctx context.Context, log *ErrorLog) error
	// List 分页查询错误日志列表
	List(ctx context.Context, page, pageSize int, username, path string, status int8) ([]*ErrorLog, int64, error)
	// Update 更新错误日志
	Update(ctx context.Context, log *ErrorLog) error
	// Delete 删除错误日志
	Delete(ctx context.Context, id int64) error
	// BatchDelete 批量删除错误日志
	BatchDelete(ctx context.Context, ids []int64) error
}
