// Package handler - 错误日志处理器
package audit

import (
	"github.com/ngq/admin-platform/app/http/handler/common"
	"strconv"

	"github.com/ngq/admin-platform/internal/biz/audit"
	"github.com/ngq/admin-platform/internal/errors"
	"github.com/ngq/gorp"
	"github.com/ngq/gorp/framework/httpx"
)

// ErrorLogHandler 错误日志处理器
type ErrorLogHandler struct {
	biz *audit.Biz
}

// NewErrorLogHandler 创建错误日志处理器
func NewErrorLogHandler(biz *audit.Biz) *ErrorLogHandler {
	return &ErrorLogHandler{biz: biz}
}

// List 分页查询错误日志列表
//
// GET /api/v1/error-logs
func (h *ErrorLogHandler) List(ctx gorp.Context) {
	pg := common.ParsePagination(ctx, 20, 100)
	username := ctx.Query("username")
	path := ctx.Query("path")
	statusStr := ctx.Query("status")

	var status int8
	if statusStr != "" {
		s, _ := strconv.ParseInt(statusStr, 10, 8)
		status = int8(s)
	}

	logs, total, err := h.biz.ErrorLogRepo.List(ctx.Context(), pg.Page, pg.PageSize, username, path, status)
	if err != nil {
		httpx.Error(ctx, err)
		return
	}

	httpx.SuccessPaginated(ctx, logs, total, pg.Page, pg.PageSize)
}

// Update 更新错误日志（如标记为已处理）
//
// POST /api/v1/error-logs/:id
func (h *ErrorLogHandler) Update(ctx gorp.Context) {
	id, err := ctx.Int64Param("id")
	if err != nil {
		httpx.Error(ctx, errors.BadRequest("无效的日志 ID"))
		return
	}

	var req struct {
		Status int8 `json:"status" validate:"required"`
	}
	if err := ctx.Bind(&req); err != nil {
		httpx.Error(ctx, errors.BadRequest("参数绑定失败: "+err.Error()))
		return
	}

	log := &audit.ErrorLog{
		ID:     id,
		Status: req.Status,
	}

	if err := h.biz.ErrorLogRepo.Update(ctx.Context(), log); err != nil {
		httpx.Error(ctx, err)
		return
	}

	httpx.Success(ctx, nil)
}

// Delete 删除错误日志
//
// POST /api/v1/error-logs/delete/:id
func (h *ErrorLogHandler) Delete(ctx gorp.Context) {
	id, err := ctx.Int64Param("id")
	if err != nil {
		httpx.Error(ctx, errors.BadRequest("无效的日志 ID"))
		return
	}

	if err := h.biz.ErrorLogRepo.Delete(ctx.Context(), id); err != nil {
		httpx.Error(ctx, err)
		return
	}

	httpx.Success(ctx, nil)
}

// BatchDelete 批量删除错误日志
//
// POST /api/v1/error-logs/batch-delete
func (h *ErrorLogHandler) BatchDelete(ctx gorp.Context) {
	var req struct {
		IDs []int64 `json:"ids" validate:"required,min=1"`
	}
	if err := ctx.Bind(&req); err != nil {
		httpx.Error(ctx, errors.BadRequest("参数绑定失败: "+err.Error()))
		return
	}

	if err := h.biz.ErrorLogRepo.BatchDelete(ctx.Context(), req.IDs); err != nil {
		httpx.Error(ctx, err)
		return
	}

	httpx.Success(ctx, nil)
}
