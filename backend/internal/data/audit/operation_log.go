// 操作日志仓储实现
// 基于 GORM 实现 audit.OperationLogRepo 接口
package audit

import (
	"context"
	"github.com/ngq/admin-platform/internal/datainfra"
	"time"

	"github.com/ngq/admin-platform/internal/biz/audit"
	"github.com/ngq/admin-platform/internal/data/po"
)

// operationLogRepo 操作日志仓储实现
type operationLogRepo struct {
	data *datainfra.Data
}

// NewOperationLogRepo 创建操作日志仓储实例
func NewOperationLogRepo(d *datainfra.Data) audit.OperationLogRepo {
	return &operationLogRepo{data: d}
}

// Create 创建操作日志
func (r *operationLogRepo) Create(ctx context.Context, log *audit.OperationLog) error {
	logPO := po.OperationLogPOFromEntity(log)
	return r.data.DB(ctx).Create(logPO).Error
}

// List 分页查询操作日志，支持按用户名/方法/路径过滤和时间范围
func (r *operationLogRepo) List(ctx context.Context, page, pageSize int, username, method, path string, startTime, endTime *time.Time) ([]*audit.OperationLog, int64, error) {
	var list []po.OperationLogPO
	var count int64

	db := r.data.DB(ctx).Model(&po.OperationLogPO{})

	// 按用户名模糊搜索
	if username != "" {
		db = db.Where("username LIKE ?", "%"+username+"%")
	}
	// 按请求方法过滤
	if method != "" {
		db = db.Where("method = ?", method)
	}
	// 按请求路径模糊搜索
	if path != "" {
		db = db.Where("path LIKE ?", "%"+path+"%")
	}
	// 时间范围过滤
	if startTime != nil {
		db = db.Where("created_at >= ?", *startTime)
	}
	if endTime != nil {
		db = db.Where("created_at <= ?", *endTime)
	}

	// 分页查询
	if err := datainfra.Paginate(db, page, pageSize, "id DESC", &list, &count); err != nil {
		return nil, 0, err
	}

	logs := make([]*audit.OperationLog, 0, len(list))
	for i := range list {
		logs = append(logs, list[i].ToEntity())
	}
	return logs, count, nil
}

// Delete 删除操作日志
func (r *operationLogRepo) Delete(ctx context.Context, id int64) error {
	return r.data.DB(ctx).Delete(&po.OperationLogPO{}, id).Error
}

// BatchDelete 批量删除操作日志
func (r *operationLogRepo) BatchDelete(ctx context.Context, ids []int64) error {
	return r.data.DB(ctx).Where("id IN ?", ids).Delete(&po.OperationLogPO{}).Error
}

// Count 查询操作日志总数（支持时间范围过滤，用于仪表盘统计）
func (r *operationLogRepo) Count(ctx context.Context, startTime, endTime *time.Time) (int64, error) {
	db := r.data.DB(ctx).Model(&po.OperationLogPO{})
	if startTime != nil {
		db = db.Where("created_at >= ?", *startTime)
	}
	if endTime != nil {
		db = db.Where("created_at <= ?", *endTime)
	}
	var count int64
	return count, db.Count(&count).Error
}

// CountDistinctUsers 查询指定时间段内不同用户数（在线用户统计）
// 使用 DB 层 COUNT(DISTINCT user_id)，避免内存过滤
func (r *operationLogRepo) CountDistinctUsers(ctx context.Context, since time.Time) (int64, error) {
	var count int64
	return count, r.data.DB(ctx).Model(&po.OperationLogPO{}).
		Where("created_at >= ? AND user_id > 0", since).
		Distinct("user_id").
		Count(&count).Error
}
