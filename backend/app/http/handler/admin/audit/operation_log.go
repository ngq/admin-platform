// Package handler - 操作日志处理器
package audit

import (
	"github.com/ngq/admin-platform/app/http/handler/common"
	"time"

	"github.com/ngq/admin-platform/internal/biz/audit"
	"github.com/ngq/admin-platform/internal/errors"
	"github.com/ngq/gorp"
	"github.com/ngq/gorp/framework/httpx"
)

// OperationLogHandler 操作日志处理器
type OperationLogHandler struct {
	biz *audit.Biz
}

// NewOperationLogHandler 创建操作日志处理器
func NewOperationLogHandler(biz *audit.Biz) *OperationLogHandler {
	return &OperationLogHandler{biz: biz}
}

// List 分页查询操作日志列表
//
// GET /api/v1/operation-logs
func (h *OperationLogHandler) List(ctx gorp.Context) {
	pg := common.ParsePagination(ctx, 20, 100)
	username := ctx.Query("username")
	method := ctx.Query("method")
	path := ctx.Query("path")
	startTimeStr := ctx.Query("start_time")
	endTimeStr := ctx.Query("end_time")

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

	logs, total, err := h.biz.OperationLogRepo.List(ctx.Context(), pg.Page, pg.PageSize, username, method, path, st, et)
	if err != nil {
		httpx.Error(ctx, err)
		return
	}

	httpx.SuccessPaginated(ctx, logs, total, pg.Page, pg.PageSize)
}

// Delete 删除操作日志
//
// POST /api/v1/operation-logs/delete/:id
func (h *OperationLogHandler) Delete(ctx gorp.Context) {
	id, err := ctx.Int64Param("id")
	if err != nil {
		httpx.Error(ctx, errors.BadRequest("无效的日志 ID"))
		return
	}

	if err := h.biz.OperationLogRepo.Delete(ctx.Context(), id); err != nil {
		httpx.Error(ctx, err)
		return
	}

	httpx.Success(ctx, nil)
}

// BatchDelete 批量删除操作日志
//
// POST /api/v1/operation-logs/batch-delete
func (h *OperationLogHandler) BatchDelete(ctx gorp.Context) {
	var req struct {
		IDs []int64 `json:"ids" validate:"required,min=1"`
	}
	if err := ctx.Bind(&req); err != nil {
		httpx.Error(ctx, errors.BadRequest("参数绑定失败: "+err.Error()))
		return
	}

	if err := h.biz.OperationLogRepo.BatchDelete(ctx.Context(), req.IDs); err != nil {
		httpx.Error(ctx, err)
		return
	}

	httpx.Success(ctx, nil)
}
