// 登录日志仓储实现
// 基于 GORM 实现 audit.LoginLogRepo 接口
package audit

import (
	"context"
	"github.com/ngq/admin-platform/internal/datainfra"
	"time"

	"github.com/ngq/admin-platform/internal/biz/audit"
	"github.com/ngq/admin-platform/internal/data/po"
)

// loginLogRepo 登录日志仓储实现
type loginLogRepo struct {
	data *datainfra.Data
}

// NewLoginLogRepo 创建登录日志仓储实例
func NewLoginLogRepo(d *datainfra.Data) audit.LoginLogRepo {
	return &loginLogRepo{data: d}
}

// Create 创建登录日志
func (r *loginLogRepo) Create(ctx context.Context, log *audit.LoginLog) error {
	logPO := po.LoginLogPOFromEntity(log)
	return r.data.DB(ctx).Create(logPO).Error
}

// List 分页查询登录日志，支持按用户名/IP/状态过滤和时间范围
func (r *loginLogRepo) List(ctx context.Context, page, pageSize int, username, ip string, status int8, startTime, endTime *time.Time) ([]*audit.LoginLog, int64, error) {
	var list []po.LoginLogPO
	var count int64

	db := r.data.DB(ctx).Model(&po.LoginLogPO{})

	// 按用户名模糊搜索
	if username != "" {
		db = db.Where("username LIKE ?", "%"+username+"%")
	}
	// 按IP过滤
	if ip != "" {
		db = db.Where("ip = ?", ip)
	}
	// 按状态过滤（0表示不过滤）
	if status != 0 {
		db = db.Where("status = ?", status)
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

	logs := make([]*audit.LoginLog, 0, len(list))
	for i := range list {
		logs = append(logs, list[i].ToEntity())
	}
	return logs, count, nil
}

// Delete 删除登录日志
func (r *loginLogRepo) Delete(ctx context.Context, id int64) error {
	return r.data.DB(ctx).Delete(&po.LoginLogPO{}, id).Error
}

// BatchDelete 批量删除登录日志
func (r *loginLogRepo) BatchDelete(ctx context.Context, ids []int64) error {
	return r.data.DB(ctx).Where("id IN ?", ids).Delete(&po.LoginLogPO{}).Error
}

// CountDistinctUsers 查询指定时间段内不同用户数（在线用户统计）
// 使用 COUNT(DISTINCT user_id) 对登录记录去重，避免 List+内存去重的不准确问题
func (r *loginLogRepo) CountDistinctUsers(ctx context.Context, since time.Time) (int64, error) {
	var count int64
	return count, r.data.DB(ctx).Model(&po.LoginLogPO{}).
		Where("created_at >= ?", since).
		Distinct("user_id").
		Count(&count).Error
}
