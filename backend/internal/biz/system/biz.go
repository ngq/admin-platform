// Package system — 系统领域子聚合
package system

import (
	"context"
	"fmt"
	"runtime"
	"time"
)

// appStartTime 记录应用启动时间
var appStartTime time.Time

// DashboardRepo 仪表盘跨域查询接口（避免 biz/system → biz 循环引用）
type DashboardRepo interface {
	CountUsers(ctx context.Context) (int64, error)
	CountRoles(ctx context.Context) (int64, error)
	CountMenus(ctx context.Context) (int64, error)
	CountAPIs(ctx context.Context) (int64, error)
	CountTodayOperationLogs(ctx context.Context) (int64, error)
	CountOnlineUsers(ctx context.Context, since time.Time) (int64, error)
}

// Biz 系统领域子聚合
type Biz struct {
	SysParamRepo SysParamRepository
	Dashboard    DashboardRepo // 仪表盘跨域查询
}

// NewBiz 创建系统领域子聚合
func NewBiz(sysParamRepo SysParamRepository, dashboardRepo DashboardRepo) *Biz {
	if appStartTime.IsZero() {
		appStartTime = time.Now()
	}
	return &Biz{SysParamRepo: sysParamRepo, Dashboard: dashboardRepo}
}

// DashboardData 仪表盘数据
type DashboardData struct {
	UserCount   int64 `json:"user_count"`
	RoleCount   int64 `json:"role_count"`
	MenuCount   int64 `json:"menu_count"`
	APICount    int64 `json:"api_count"`
	OnlineCount int64 `json:"online_count"`
	LogCount    int64 `json:"log_count"`
}

// SystemInfoData 系统运行信息
type SystemInfoData struct {
	GoVersion    string `json:"go_version"`
	NumGoroutine int    `json:"num_goroutine"`
	NumCPU       int    `json:"num_cpu"`
	StartTime    string `json:"start_time"`
	Uptime       string `json:"uptime"`
}

// GetDashboard 获取仪表盘数据
func (b *Biz) GetDashboard(ctx context.Context) (*DashboardData, error) {
	data := &DashboardData{}
	userCount, err := b.Dashboard.CountUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("获取用户统计失败: %w", err)
	}
	data.UserCount = userCount

	roleCount, err := b.Dashboard.CountRoles(ctx)
	if err != nil {
		return nil, fmt.Errorf("获取角色统计失败: %w", err)
	}
	data.RoleCount = roleCount

	menuCount, err := b.Dashboard.CountMenus(ctx)
	if err != nil {
		return nil, fmt.Errorf("获取菜单统计失败: %w", err)
	}
	data.MenuCount = menuCount

	apiCount, err := b.Dashboard.CountAPIs(ctx)
	if err != nil {
		return nil, fmt.Errorf("获取API统计失败: %w", err)
	}
	data.APICount = apiCount

	logCount, err := b.Dashboard.CountTodayOperationLogs(ctx)
	if err != nil {
		return nil, fmt.Errorf("获取日志统计失败: %w", err)
	}
	data.LogCount = logCount

	thirtyMinAgo := time.Now().Add(-30 * time.Minute)
	onlineCount, err := b.Dashboard.CountOnlineUsers(ctx, thirtyMinAgo)
	if err != nil {
		return nil, fmt.Errorf("获取在线用户统计失败: %w", err)
	}
	data.OnlineCount = onlineCount
	return data, nil
}

// GetSystemInfo 获取系统运行信息
func (b *Biz) GetSystemInfo(ctx context.Context) (*SystemInfoData, error) {
	uptime := time.Since(appStartTime)
	return &SystemInfoData{
		GoVersion:    runtime.Version(),
		NumGoroutine: runtime.NumGoroutine(),
		NumCPU:       runtime.NumCPU(),
		StartTime:    appStartTime.Format("2006-01-02 15:04:05"),
		Uptime:       fmt.Sprintf("%.0f小时%.0f分钟", uptime.Hours(), uptime.Minutes()-uptime.Hours()*60),
	}, nil
}

// ========== 系统参数业务方法 ==========

// CreateParam 创建系统参数
func (b *Biz) CreateParam(ctx context.Context, p *SysParam) error {
	return b.SysParamRepo.Create(ctx, p)
}

// UpdateParam 更新系统参数
func (b *Biz) UpdateParam(ctx context.Context, p *SysParam) error {
	return b.SysParamRepo.Update(ctx, p)
}

// DeleteParam 删除系统参数
func (b *Biz) DeleteParam(ctx context.Context, id int64) error {
	return b.SysParamRepo.Delete(ctx, id)
}

// GetParamByID 根据ID获取系统参数
func (b *Biz) GetParamByID(ctx context.Context, id int64) (*SysParam, error) {
	return b.SysParamRepo.GetByID(ctx, id)
}

// GetParamByKey 根据键名获取系统参数
func (b *Biz) GetParamByKey(ctx context.Context, key string) (*SysParam, error) {
	return b.SysParamRepo.GetByKey(ctx, key)
}

// ListParams 分页查询系统参数列表
func (b *Biz) ListParams(ctx context.Context, page, pageSize int, category, key string) ([]*SysParam, int64, error) {
	return b.SysParamRepo.List(ctx, page, pageSize, category, key)
}
