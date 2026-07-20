package system

import (
	"context"
	"time"

	"github.com/ngq/admin-platform/internal/biz/system"
	"github.com/ngq/admin-platform/internal/datainfra"
)

// NewBiz 创建系统领域子聚合（DI 桥接）
func NewBiz(d *datainfra.Data, dashboardRepo system.DashboardRepo) *system.Biz {
	return system.NewBiz(NewParamRepo(d), dashboardRepo)
}

// dashboardRepo 仪表盘跨域查询实现
type dashboardRepo struct {
	d *datainfra.Data
}

// NewDashboardRepo 创建仪表盘跨域查询实现
func NewDashboardRepo(d *datainfra.Data) system.DashboardRepo {
	return &dashboardRepo{d: d}
}

// CountUsers 统计用户总数
func (r *dashboardRepo) CountUsers(ctx context.Context) (int64, error) {
	var count int64
	err := r.d.DB(ctx).Table("sys_users").Count(&count).Error
	return count, err
}

// CountRoles 统计角色总数
func (r *dashboardRepo) CountRoles(ctx context.Context) (int64, error) {
	var count int64
	err := r.d.DB(ctx).Table("sys_roles").Count(&count).Error
	return count, err
}

// CountMenus 统计菜单总数
func (r *dashboardRepo) CountMenus(ctx context.Context) (int64, error) {
	var count int64
	err := r.d.DB(ctx).Table("sys_menus").Count(&count).Error
	return count, err
}

// CountAPIs 统计API总数
func (r *dashboardRepo) CountAPIs(ctx context.Context) (int64, error) {
	var count int64
	err := r.d.DB(ctx).Table("sys_apis").Count(&count).Error
	return count, err
}

// CountTodayOperationLogs 统计今日操作日志数
func (r *dashboardRepo) CountTodayOperationLogs(ctx context.Context) (int64, error) {
	var count int64
	todayStart := time.Now().Truncate(24 * time.Hour)
	err := r.d.DB(ctx).Table("sys_operation_logs").Where("created_at >= ?", todayStart).Count(&count).Error
	return count, err
}

// CountOnlineUsers 统计在线用户数
func (r *dashboardRepo) CountOnlineUsers(ctx context.Context, since time.Time) (int64, error) {
	var count int64
	err := r.d.DB(ctx).Table("sys_operation_logs").
		Select("COUNT(DISTINCT username)").
		Where("created_at >= ?", since).
		Scan(&count).Error
	return count, err
}
