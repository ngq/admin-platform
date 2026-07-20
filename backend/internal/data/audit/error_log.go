// 错误日志仓储实现
// 基于 GORM 实现 audit.ErrorLogRepo 接口
package audit

import (
	"context"
	"github.com/ngq/admin-platform/internal/datainfra"

	"github.com/ngq/admin-platform/internal/biz/audit"
	"github.com/ngq/admin-platform/internal/data/po"
)

// errorLogRepo 错误日志仓储实现
type errorLogRepo struct {
	data *datainfra.Data
}

// NewErrorLogRepo 创建错误日志仓储实例
func NewErrorLogRepo(d *datainfra.Data) audit.ErrorLogRepo {
	return &errorLogRepo{data: d}
}

// Create 创建错误日志
func (r *errorLogRepo) Create(ctx context.Context, log *audit.ErrorLog) error {
	logPO := po.ErrorLogPOFromEntity(log)
	return r.data.DB(ctx).Create(logPO).Error
}

// List 分页查询错误日志，支持按用户名、请求路径和处理状态过滤
func (r *errorLogRepo) List(ctx context.Context, page, pageSize int, username, path string, status int8) ([]*audit.ErrorLog, int64, error) {
	var list []po.ErrorLogPO
	var count int64

	db := r.data.DB(ctx).Model(&po.ErrorLogPO{})

	// 按用户名模糊搜索
	if username != "" {
		db = db.Where("username LIKE ?", "%"+username+"%")
	}
	// 按请求路径模糊搜索
	if path != "" {
		db = db.Where("path LIKE ?", "%"+path+"%")
	}
	// 按处理状态过滤（0 表示不过滤）
	if status != 0 {
		db = db.Where("status = ?", status)
	}

	// 分页查询
	if err := datainfra.Paginate(db, page, pageSize, "id DESC", &list, &count); err != nil {
		return nil, 0, err
	}

	logs := make([]*audit.ErrorLog, 0, len(list))
	for i := range list {
		logs = append(logs, list[i].ToEntity())
	}
	return logs, count, nil
}

// Update 更新错误日志（如标记为已处理）
func (r *errorLogRepo) Update(ctx context.Context, log *audit.ErrorLog) error {
	logPO := po.ErrorLogPOFromEntity(log)
	return r.data.DB(ctx).Model(&po.ErrorLogPO{}).
		Where("id = ?", log.ID).
		Updates(map[string]interface{}{
			"status": logPO.Status,
		}).Error
}

// Delete 删除错误日志
func (r *errorLogRepo) Delete(ctx context.Context, id int64) error {
	return r.data.DB(ctx).Delete(&po.ErrorLogPO{}, id).Error
}

// BatchDelete 批量删除错误日志
func (r *errorLogRepo) BatchDelete(ctx context.Context, ids []int64) error {
	return r.data.DB(ctx).Where("id IN ?", ids).Delete(&po.ErrorLogPO{}).Error
}
