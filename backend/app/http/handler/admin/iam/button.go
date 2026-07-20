// Package handler - 按钮权限管理处理器
package iam

import (
	"strconv"

	bizIam "github.com/ngq/admin-platform/internal/biz/iam"
	"github.com/ngq/admin-platform/internal/errors"
	"github.com/ngq/gorp"
	"github.com/ngq/gorp/framework/httpx"
)

// createButtonRequest 创建按钮权限的请求参数
type createButtonRequest struct {
	MenuID      int64  `json:"menu_id"`     // 所属菜单 ID
	Name        string `json:"name"`        // 按钮名称
	Code        string `json:"code"`        // 按钮编码（权限标识）
	Description string `json:"description"` // 描述
}

// updateButtonRequest 更新按钮权限的请求参数（ID 从 URL 路径获取）
type updateButtonRequest struct {
	MenuID      int64  `json:"menu_id"`     // 所属菜单 ID
	Name        string `json:"name"`        // 按钮名称
	Code        string `json:"code"`        // 按钮编码（权限标识）
	Description string `json:"description"` // 描述
}

// ButtonHandler 按钮权限管理处理器
type ButtonHandler struct {
	biz *bizIam.Biz
}

// NewButtonHandler 创建按钮权限管理处理器
func NewButtonHandler(biz *bizIam.Biz) *ButtonHandler {
	return &ButtonHandler{biz: biz}
}

// GetByMenuID 根据菜单 ID 获取按钮权限列表
//
// GET /api/v1/menus/buttons?menuId=xxx
//
// 用于前端加载指定菜单下的按钮权限，控制页面按钮的显示与隐藏。
func (h *ButtonHandler) GetByMenuID(ctx gorp.Context) {
	menuID, err := strconv.ParseInt(ctx.Query("menuId"), 10, 64)
	if err != nil {
		httpx.Error(ctx, errors.BadRequest("无效的菜单 ID"))
		return
	}

	buttons, err := h.biz.ButtonGetByMenuID(ctx.Context(), menuID)
	if err != nil {
		httpx.Error(ctx, err)
		return
	}

	httpx.Success(ctx, buttons)
}

// Create 创建按钮权限
//
// POST /api/v1/buttons
//
// 请求体绑定到 createButtonRequest，转换为 iam.Button 后调用业务层创建。
func (h *ButtonHandler) Create(ctx gorp.Context) {
	var req createButtonRequest
	if err := ctx.Bind(&req); err != nil {
		httpx.Error(ctx, errors.BadRequest("参数绑定失败: "+err.Error()))
		return
	}

	// 将请求参数转换为业务实体
	button := &bizIam.Button{
		MenuID:      req.MenuID,
		Name:        req.Name,
		Code:        req.Code,
		Description: req.Description,
	}

	if err := h.biz.ButtonCreate(ctx.Context(), button); err != nil {
		httpx.Error(ctx, err)
		return
	}

	httpx.Success(ctx, nil)
}

// Update 更新按钮权限
//
// POST /api/v1/buttons/:id
//
// ID 从 URL 路径参数获取，请求体绑定到 updateButtonRequest。
func (h *ButtonHandler) Update(ctx gorp.Context) {
	id, err := ctx.Int64Param("id")
	if err != nil {
		httpx.Error(ctx, errors.BadRequest("无效的按钮 ID"))
		return
	}

	var req updateButtonRequest
	if err := ctx.Bind(&req); err != nil {
		httpx.Error(ctx, errors.BadRequest("参数绑定失败: "+err.Error()))
		return
	}

	// 将请求参数与路径 ID 组装为业务实体
	button := &bizIam.Button{
		ID:          id,
		MenuID:      req.MenuID,
		Name:        req.Name,
		Code:        req.Code,
		Description: req.Description,
	}

	if err := h.biz.ButtonUpdate(ctx.Context(), button); err != nil {
		httpx.Error(ctx, err)
		return
	}

	httpx.Success(ctx, nil)
}

// Delete 删除按钮权限
//
// POST /api/v1/buttons/delete/:id
func (h *ButtonHandler) Delete(ctx gorp.Context) {
	id, err := ctx.Int64Param("id")
	if err != nil {
		httpx.Error(ctx, errors.BadRequest("无效的按钮 ID"))
		return
	}

	if err := h.biz.ButtonDelete(ctx.Context(), id); err != nil {
		httpx.Error(ctx, err)
		return
	}

	httpx.Success(ctx, nil)
}
