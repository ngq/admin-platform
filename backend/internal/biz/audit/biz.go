// Package audit — 审计领域子聚合
package audit

import (
	"context"
	"time"
)

// Biz 审计领域子聚合
type Biz struct {
	OperationLogRepo OperationLogRepo
	LoginLogRepo     LoginLogRepo
	ErrorLogRepo     ErrorLogRepo
}

// NewBiz 创建审计领域子聚合
func NewBiz(operationLogRepo OperationLogRepo, loginLogRepo LoginLogRepo, errorLogRepo ErrorLogRepo) *Biz {
	return &Biz{
		OperationLogRepo: operationLogRepo,
		LoginLogRepo:     loginLogRepo,
		ErrorLogRepo:     errorLogRepo,
	}
}

// ========== 操作日志业务方法 ==========

// CreateOperationLog 创建操作日志记录
func (b *Biz) CreateOperationLog(ctx context.Context, log *OperationLog) error {
	return b.OperationLogRepo.Create(ctx, log)
}

// ListOperationLogs 分页查询操作日志
func (b *Biz) ListOperationLogs(ctx context.Context, page, pageSize int, username, method, path string, startTime, endTime *time.Time) ([]*OperationLog, int64, error) {
	return b.OperationLogRepo.List(ctx, page, pageSize, username, method, path, startTime, endTime)
}

// DeleteOperationLog 删除操作日志
func (b *Biz) DeleteOperationLog(ctx context.Context, id int64) error {
	return b.OperationLogRepo.Delete(ctx, id)
}

// BatchDeleteOperationLogs 批量删除操作日志
func (b *Biz) BatchDeleteOperationLogs(ctx context.Context, ids []int64) error {
	return b.OperationLogRepo.BatchDelete(ctx, ids)
}

// ========== 登录日志业务方法 ==========

// ListLoginLogs 分页查询登录日志
func (b *Biz) ListLoginLogs(ctx context.Context, page, pageSize int, username, ip string, status int8, startTime, endTime *time.Time) ([]*LoginLog, int64, error) {
	return b.LoginLogRepo.List(ctx, page, pageSize, username, ip, status, startTime, endTime)
}

// DeleteLoginLog 删除登录日志
func (b *Biz) DeleteLoginLog(ctx context.Context, id int64) error {
	return b.LoginLogRepo.Delete(ctx, id)
}

// BatchDeleteLoginLogs 批量删除登录日志
func (b *Biz) BatchDeleteLoginLogs(ctx context.Context, ids []int64) error {
	return b.LoginLogRepo.BatchDelete(ctx, ids)
}

// ========== 错误日志业务方法 ==========

// ListErrorLogs 分页查询错误日志
func (b *Biz) ListErrorLogs(ctx context.Context, page, pageSize int, username, path string, status int8) ([]*ErrorLog, int64, error) {
	return b.ErrorLogRepo.List(ctx, page, pageSize, username, path, status)
}

// UpdateErrorLog 更新错误日志
func (b *Biz) UpdateErrorLog(ctx context.Context, log *ErrorLog) error {
	return b.ErrorLogRepo.Update(ctx, log)
}

// DeleteErrorLog 删除错误日志
func (b *Biz) DeleteErrorLog(ctx context.Context, id int64) error {
	return b.ErrorLogRepo.Delete(ctx, id)
}

// BatchDeleteErrorLogs 批量删除错误日志
func (b *Biz) BatchDeleteErrorLogs(ctx context.Context, ids []int64) error {
	return b.ErrorLogRepo.BatchDelete(ctx, ids)
}
