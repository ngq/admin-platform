// Package handler - 系统信息处理器（仪表盘、健康检查等）
package system

import (
	"github.com/ngq/admin-platform/internal/biz/system"
	"github.com/ngq/gorp"
	"github.com/ngq/gorp/framework/httpx"
)

// SystemHandler 系统信息处理器
type SystemHandler struct {
	biz *system.Biz
}

// NewSystemHandler 创建系统信息处理器
func NewSystemHandler(biz *system.Biz) *SystemHandler {
	return &SystemHandler{biz: biz}
}

// Dashboard 获取仪表盘数据
//
// GET /api/v1/system/dashboard
func (h *SystemHandler) Dashboard(ctx gorp.Context) {
	data, err := h.biz.GetDashboard(ctx.Context())
	if err != nil {
		httpx.Error(ctx, err)
		return
	}

	httpx.Success(ctx, data)
}

// SystemInfo 获取系统信息
//
// GET /api/v1/system/info
func (h *SystemHandler) SystemInfo(ctx gorp.Context) {
	data, err := h.biz.GetSystemInfo(ctx.Context())
	if err != nil {
		httpx.Error(ctx, err)
		return
	}

	httpx.Success(ctx, data)
}

// HealthCheck 健康检查
//
// GET /api/v1/system/health
func (h *SystemHandler) HealthCheck(ctx gorp.Context) {
	httpx.Success(ctx, map[string]string{"status": "ok"})
}
