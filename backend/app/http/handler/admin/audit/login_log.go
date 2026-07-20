// Package handler - 登录日志处理器
package audit

import (
	"github.com/ngq/admin-platform/app/http/handler/common"
	"strconv"
	"time"

	"github.com/ngq/admin-platform/internal/biz/audit"
	"github.com/ngq/admin-platform/internal/errors"
	"github.com/ngq/gorp"
	"github.com/ngq/gorp/framework/httpx"
)

// LoginLogHandler 登录日志处理器
type LoginLogHandler struct {
	biz *audit.Biz
}

// NewLoginLogHandler 创建登录日志处理器
func NewLoginLogHandler(biz *audit.Biz) *LoginLogHandler {
	return &LoginLogHandler{biz: biz}
}

// List 分页查询登录日志列表
//
// GET /api/v1/login-logs
func (h *LoginLogHandler) List(ctx gorp.Context) {
	pg := common.ParsePagination(ctx, 20, 100)
	username := ctx.Query("username")
	ip := ctx.Query("ip")
	statusStr := ctx.Query("status")
	startTimeStr := ctx.Query("start_time")
	endTimeStr := ctx.Query("end_time")

	var status int8
	if statusStr != "" {
		s, _ := strconv.ParseInt(statusStr, 10, 8)
		status = int8(s)
	}

	var st, et *time.Time
	if startTimeStr != "" {
		t, err := time.Parse(time.RFC3339, startTimeStr)
		if err == nil {
			st = &t
		}
	}
	if endTimeStr != "" {
		t, err := time.Parse(time.RFC3339, endTimeStr)
		if err == nil {
			et = &t
		}
	}

	logs, total, err := h.biz.LoginLogRepo.List(ctx.Context(), pg.Page, pg.PageSize, username, ip, status, st, et)
	if err != nil {
		httpx.Error(ctx, err)
		return
	}

	httpx.SuccessPaginated(ctx, logs, total, pg.Page, pg.PageSize)
}

// Delete 删除登录日志
//
// POST /api/v1/login-logs/delete/:id
func (h *LoginLogHandler) Delete(ctx gorp.Context) {
	id, err := ctx.Int64Param("id")
	if err != nil {
		httpx.Error(ctx, errors.BadRequest("无效的日志 ID"))
		return
	}

	if err := h.biz.LoginLogRepo.Delete(ctx.Context(), id); err != nil {
		httpx.Error(ctx, err)
		return
	}

	httpx.Success(ctx, nil)
}

// BatchDelete 批量删除登录日志
//
// POST /api/v1/login-logs/batch-delete
func (h *LoginLogHandler) BatchDelete(ctx gorp.Context) {
	var req struct {
		IDs []int64 `json:"ids" validate:"required,min=1"`
	}
	if err := ctx.Bind(&req); err != nil {
		httpx.Error(ctx, errors.BadRequest("参数绑定失败: "+err.Error()))
		return
	}

	if err := h.biz.LoginLogRepo.BatchDelete(ctx.Context(), req.IDs); err != nil {
		httpx.Error(ctx, err)
		return
	}

	httpx.Success(ctx, nil)
}
