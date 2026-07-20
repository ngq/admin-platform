// Package handler - 系统参数管理处理器
package system

import (
	"github.com/ngq/admin-platform/app/http/handler/common"
	"github.com/ngq/admin-platform/internal/biz/system"
	"github.com/ngq/admin-platform/internal/errors"
	"github.com/ngq/gorp"
	"github.com/ngq/gorp/framework/httpx"
)

// ParamHandler 系统参数管理处理器
type ParamHandler struct {
	biz *system.Biz
}

// NewParamHandler 创建系统参数管理处理器
func NewParamHandler(biz *system.Biz) *ParamHandler {
	return &ParamHandler{biz: biz}
}

// List 分页查询系统参数列表
//
// GET /api/v1/params
//
// 支持按分类、参数键筛选
func (h *ParamHandler) List(ctx gorp.Context) {
	pg := common.ParsePagination(ctx, 20, 100)
	category := ctx.Query("category")
	key := ctx.Query("key")

	params, total, err := h.biz.SysParamRepo.List(ctx.Context(), pg.Page, pg.PageSize, category, key)
	if err != nil {
		httpx.Error(ctx, err)
		return
	}

	httpx.SuccessPaginated(ctx, params, total, pg.Page, pg.PageSize)
}

// Get 根据 ID 获取系统参数详情
//
// GET /api/v1/params/:id
func (h *ParamHandler) Get(ctx gorp.Context) {
	id, err := ctx.Int64Param("id")
	if err != nil {
		httpx.Error(ctx, errors.BadRequest("无效的参数 ID"))
		return
	}

	p, err := h.biz.SysParamRepo.GetByID(ctx.Context(), id)
	if err != nil {
		httpx.Error(ctx, err)
		return
	}

	httpx.Success(ctx, p)
}

// Create 创建系统参数
//
// POST /api/v1/params
func (h *ParamHandler) Create(ctx gorp.Context) {
	var req struct {
		Key         string `json:"key" validate:"required,max=100"`
		Value       string `json:"value" validate:"required"`
		Category    string `json:"category" validate:"max=50"`
		Description string `json:"description" validate:"max=200"`
		Status      int8   `json:"status" validate:"omitempty,oneof=1 2"`
	}
	if err := ctx.Bind(&req); err != nil {
		httpx.Error(ctx, errors.BadRequest("参数绑定失败: "+err.Error()))
		return
	}

	p := &system.SysParam{
		Key:         req.Key,
		Value:       req.Value,
		Category:    req.Category,
		Description: req.Description,
		Status:      req.Status,
	}

	if err := h.biz.SysParamRepo.Create(ctx.Context(), p); err != nil {
		httpx.Error(ctx, err)
		return
	}

	httpx.Success(ctx, nil)
}

// Update 更新系统参数
//
// POST /api/v1/params/:id
func (h *ParamHandler) Update(ctx gorp.Context) {
	id, err := ctx.Int64Param("id")
	if err != nil {
		httpx.Error(ctx, errors.BadRequest("无效的参数 ID"))
		return
	}

	var req struct {
		Key         string `json:"key" validate:"max=100"`
		Value       string `json:"value"`
		Category    string `json:"category" validate:"max=50"`
		Description string `json:"description" validate:"max=200"`
		Status      int8   `json:"status" validate:"omitempty,oneof=1 2"`
	}
	if err := ctx.Bind(&req); err != nil {
		httpx.Error(ctx, errors.BadRequest("参数绑定失败: "+err.Error()))
		return
	}

	p := &system.SysParam{
		ID:          id,
		Key:         req.Key,
		Value:       req.Value,
		Category:    req.Category,
		Description: req.Description,
		Status:      req.Status,
	}

	if err := h.biz.SysParamRepo.Update(ctx.Context(), p); err != nil {
		httpx.Error(ctx, err)
		return
	}

	httpx.Success(ctx, nil)
}

// Delete 删除系统参数
//
// POST /api/v1/params/delete/:id
func (h *ParamHandler) Delete(ctx gorp.Context) {
	id, err := ctx.Int64Param("id")
	if err != nil {
		httpx.Error(ctx, errors.BadRequest("无效的参数 ID"))
		return
	}

	if err := h.biz.SysParamRepo.Delete(ctx.Context(), id); err != nil {
		httpx.Error(ctx, err)
		return
	}

	httpx.Success(ctx, nil)
}
