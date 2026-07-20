// Package handler - 角色管理处理器
package iam

import (
	"github.com/ngq/admin-platform/app/http/handler/common"
	svciam "github.com/ngq/admin-platform/internal/service/iam"
	"strconv"

	"github.com/ngq/gorp"
	"github.com/ngq/gorp/framework/httpx"

	"github.com/ngq/admin-platform/internal/biz/iam"
	"github.com/ngq/admin-platform/internal/errors"
)

// ──────────────────────────────────────────────────────
// 角色相关请求结构体（内联定义，不依赖外部 request 包）
// ──────────────────────────────────────────────────────

// createRoleRequest 创建角色请求
type createRoleRequest struct {
	Name        string `json:"name" validate:"required,max=50"`               // 角色名称
	Code        string `json:"code" validate:"required,max=50"`               // 角色编码
	ParentID    int64  `json:"parent_id"`                                     // 父角色 ID
	Sort        int    `json:"sort"`                                          // 排序序号
	DataScope   int8   `json:"data_scope" validate:"omitempty,oneof=1 2 3 4"` // 数据权限范围
	Status      int8   `json:"status" validate:"omitempty,oneof=1 2"`         // 状态
	Description string `json:"description" validate:"max=200"`                // 角色描述
}

// updateRoleRequest 更新角色请求
type updateRoleRequest struct {
	Name        string `json:"name" validate:"max=50"`                        // 角色名称
	Code        string `json:"code" validate:"max=50"`                        // 角色编码
	ParentID    int64  `json:"parent_id"`                                     // 父角色 ID
	Sort        int    `json:"sort"`                                          // 排序序号
	DataScope   int8   `json:"data_scope" validate:"omitempty,oneof=1 2 3 4"` // 数据权限范围
	Status      int8   `json:"status" validate:"omitempty,oneof=1 2"`         // 状态
	Description string `json:"description" validate:"max=200"`                // 角色描述
}

// assignMenusRequest 分配菜单请求
type assignMenusRequest struct {
	MenuIDs []int64 `json:"menu_ids" validate:"required"` // 菜单 ID 列表
}

// assignAPIsRequest 分配 API 资源请求
type assignAPIsRequest struct {
	APIIDs []int64 `json:"api_ids" validate:"required"` // API 资源 ID 列表
}

// assignButtonsRequest 分配按钮权限请求
type assignButtonsRequest struct {
	ButtonIDs []int64 `json:"button_ids" validate:"required"` // 按钮 ID 列表
}

type assignPermissionConfigRequest struct {
	MenuIDs   []int64 `json:"menu_ids"`
	APIIDs    []int64 `json:"api_ids"`
	ButtonIDs []int64 `json:"button_ids"`
}

// ──────────────────────────────────────────────────────
// RoleHandler 角色管理处理器
// ──────────────────────────────────────────────────────

// RoleHandler 角色管理处理器，封装角色相关的 HTTP 接口
type RoleHandler struct {
	svc *svciam.RoleService
}

// NewRoleHandler 创建角色管理处理器
func NewRoleHandler(svc *svciam.RoleService) *RoleHandler {
	return &RoleHandler{svc: svc}
}

// List 分页查询角色列表
//
// 支持按名称、编码、状态筛选，返回分页结果。
//
// GET /api/v1/roles
func (h *RoleHandler) List(ctx gorp.Context) {
	// 解析分页参数，缺省 page=1, page_size=20，上限 100
	pg := common.ParsePagination(ctx, 20, 100)
	// 可选筛选条件
	name := ctx.Query("name")
	code := ctx.Query("code")
	status := ctx.Query("status")

	// 调用业务层获取分页数据
	roles, total, err := h.svc.List(ctx.Context(), pg.Page, pg.PageSize, name, code, status)
	if err != nil {
		httpx.Error(ctx, err)
		return
	}

	// 输出分页成功响应
	httpx.SuccessPaginated(ctx, roles, total, pg.Page, pg.PageSize)
}

// Get 根据 ID 获取角色详情
//
// GET /api/v1/roles/:id
func (h *RoleHandler) Get(ctx gorp.Context) {
	// 解析路径参数中的角色 ID
	id, err := ctx.Int64Param("id")
	if err != nil {
		httpx.Error(ctx, errors.BadRequest("无效的角色 ID"))
		return
	}

	// 查询角色详情
	role, err := h.svc.GetByID(ctx.Context(), id)
	if err != nil {
		httpx.Error(ctx, err)
		return
	}

	httpx.Success(ctx, role)
}

// Create 创建角色
//
// POST /api/v1/roles
func (h *RoleHandler) Create(ctx gorp.Context) {
	// 绑定请求参数到内联结构体
	var req createRoleRequest
	if err := ctx.Bind(&req); err != nil {
		httpx.Error(ctx, errors.BadRequest("参数绑定失败: "+err.Error()))
		return
	}

	// 将请求 DTO 转换为业务实体
	role := &iam.Role{
		Name:        req.Name,
		Code:        req.Code,
		ParentID:    req.ParentID,
		Sort:        req.Sort,
		DataScope:   req.DataScope,
		Status:      req.Status,
		Description: req.Description,
	}

	// 调用业务层创建角色
	if err := h.svc.Create(ctx.Context(), role); err != nil {
		httpx.Error(ctx, err)
		return
	}

	httpx.Success(ctx, nil)
}

// Update 更新角色
//
// POST /api/v1/roles/:id
func (h *RoleHandler) Update(ctx gorp.Context) {
	// 解析路径参数中的角色 ID
	id, err := ctx.Int64Param("id")
	if err != nil {
		httpx.Error(ctx, errors.BadRequest("无效的角色 ID"))
		return
	}

	// 绑定请求参数到内联结构体
	var req updateRoleRequest
	if err := ctx.Bind(&req); err != nil {
		httpx.Error(ctx, errors.BadRequest("参数绑定失败: "+err.Error()))
		return
	}

	// 将请求 DTO 转换为业务实体，携带路径参数 ID
	role := &iam.Role{
		ID:          id,
		Name:        req.Name,
		Code:        req.Code,
		ParentID:    req.ParentID,
		Sort:        req.Sort,
		DataScope:   req.DataScope,
		Status:      req.Status,
		Description: req.Description,
	}

	// 调用业务层更新角色
	if err := h.svc.Update(ctx.Context(), role); err != nil {
		httpx.Error(ctx, err)
		return
	}

	httpx.Success(ctx, nil)
}

// Delete 删除角色
//
// POST /api/v1/roles/delete/:id
func (h *RoleHandler) Delete(ctx gorp.Context) {
	// 解析路径参数中的角色 ID
	id, err := ctx.Int64Param("id")
	if err != nil {
		httpx.Error(ctx, errors.BadRequest("无效的角色 ID"))
		return
	}

	// 调用业务层删除角色
	if err := h.svc.Delete(ctx.Context(), id); err != nil {
		httpx.Error(ctx, err)
		return
	}

	httpx.Success(ctx, nil)
}

// AssignMenus 分配菜单给角色
//
// 将一组菜单 ID 绑定到指定角色，用于角色-菜单权限管理。
//
// POST /api/v1/roles/menus?id=xxx
func (h *RoleHandler) AssignMenus(ctx gorp.Context) {
	// 从查询参数中解析角色 ID
	id, err := strconv.ParseInt(ctx.Query("id"), 10, 64)
	if err != nil {
		httpx.Error(ctx, errors.BadRequest("无效的角色 ID"))
		return
	}

	// 绑定请求参数到内联结构体
	var req assignMenusRequest
	if err := ctx.Bind(&req); err != nil {
		httpx.Error(ctx, errors.BadRequest("参数绑定失败: "+err.Error()))
		return
	}

	// 调用业务层分配菜单
	if err := h.svc.AssignMenus(ctx.Context(), id, req.MenuIDs); err != nil {
		httpx.Error(ctx, err)
		return
	}

	httpx.Success(ctx, nil)
}

// AssignAPIs 分配 API 资源给角色
//
// 将一组 API 资源 ID 绑定到指定角色，用于角色-API 权限管理。
//
// POST /api/v1/roles/apis?id=xxx
func (h *RoleHandler) AssignAPIs(ctx gorp.Context) {
	// 从查询参数中解析角色 ID
	id, err := strconv.ParseInt(ctx.Query("id"), 10, 64)
	if err != nil {
		httpx.Error(ctx, errors.BadRequest("无效的角色 ID"))
		return
	}

	// 绑定请求参数到内联结构体
	var req assignAPIsRequest
	if err := ctx.Bind(&req); err != nil {
		httpx.Error(ctx, errors.BadRequest("参数绑定失败: "+err.Error()))
		return
	}

	// 调用业务层分配 API 资源
	if err := h.svc.AssignAPIs(ctx.Context(), id, req.APIIDs); err != nil {
		httpx.Error(ctx, err)
		return
	}

	httpx.Success(ctx, nil)
}

// AssignButtons 分配按钮权限给角色
//
// 将一组按钮 ID 绑定到指定角色，用于角色-按钮权限管理。
//
// POST /api/v1/roles/buttons?id=xxx
func (h *RoleHandler) AssignButtons(ctx gorp.Context) {
	// 从查询参数中解析角色 ID
	id, err := strconv.ParseInt(ctx.Query("id"), 10, 64)
	if err != nil {
		httpx.Error(ctx, errors.BadRequest("无效的角色 ID"))
		return
	}

	// 绑定请求参数到内联结构体
	var req assignButtonsRequest
	if err := ctx.Bind(&req); err != nil {
		httpx.Error(ctx, errors.BadRequest("参数绑定失败: "+err.Error()))
		return
	}

	// 调用业务层分配按钮权限
	if err := h.svc.AssignButtons(ctx.Context(), id, req.ButtonIDs); err != nil {
		httpx.Error(ctx, err)
		return
	}

	httpx.Success(ctx, nil)
}

// GetMenuIDs 获取角色关联的菜单 ID 列表
//
// GET /api/v1/roles/menus?id=xxx
func (h *RoleHandler) GetMenuIDs(ctx gorp.Context) {
	// 从查询参数中解析角色 ID
	id, err := strconv.ParseInt(ctx.Query("id"), 10, 64)
	if err != nil {
		httpx.Error(ctx, errors.BadRequest("无效的角色 ID"))
		return
	}

	// 查询角色关联的菜单 ID
	menuIDs, err := h.svc.GetMenuIDs(ctx.Context(), id)
	if err != nil {
		httpx.Error(ctx, err)
		return
	}

	httpx.Success(ctx, map[string]any{"menu_ids": menuIDs})
}

// GetAPIIDs 获取角色关联的 API 资源 ID 列表
//
// GET /api/v1/roles/apis?id=xxx
func (h *RoleHandler) GetAPIIDs(ctx gorp.Context) {
	// 从查询参数中解析角色 ID
	id, err := strconv.ParseInt(ctx.Query("id"), 10, 64)
	if err != nil {
		httpx.Error(ctx, errors.BadRequest("无效的角色 ID"))
		return
	}

	// 查询角色关联的 API 资源 ID
	apiIDs, err := h.svc.GetAPIIDs(ctx.Context(), id)
	if err != nil {
		httpx.Error(ctx, err)
		return
	}

	httpx.Success(ctx, map[string]any{"api_ids": apiIDs})
}

// GetButtonIDs 获取角色关联的按钮权限 ID 列表
//
// GET /api/v1/roles/buttons?id=xxx
func (h *RoleHandler) GetButtonIDs(ctx gorp.Context) {
	// 从查询参数中解析角色 ID
	id, err := strconv.ParseInt(ctx.Query("id"), 10, 64)
	if err != nil {
		httpx.Error(ctx, errors.BadRequest("无效的角色 ID"))
		return
	}

	// 查询角色关联的按钮权限 ID
	buttonIDs, err := h.svc.GetButtonIDs(ctx.Context(), id)
	if err != nil {
		httpx.Error(ctx, err)
		return
	}

	httpx.Success(ctx, map[string]any{"button_ids": buttonIDs})
}

// GetPermissionConfig 获取角色权限配置聚合模型
//
// GET /api/v1/roles/permission-config?id=xxx
func (h *RoleHandler) GetPermissionConfig(ctx gorp.Context) {
	id, err := strconv.ParseInt(ctx.Query("id"), 10, 64)
	if err != nil {
		httpx.Error(ctx, errors.BadRequest("无效的角色 ID"))
		return
	}

	config, err := h.svc.GetPermissionConfig(ctx.Context(), id)
	if err != nil {
		httpx.Error(ctx, err)
		return
	}

	httpx.Success(ctx, config)
}

// AssignPermissionConfig 保存角色权限配置聚合模型
//
// POST /api/v1/roles/permission-config?id=xxx
func (h *RoleHandler) AssignPermissionConfig(ctx gorp.Context) {
	id, err := strconv.ParseInt(ctx.Query("id"), 10, 64)
	if err != nil {
		httpx.Error(ctx, errors.BadRequest("无效的角色 ID"))
		return
	}

	var req assignPermissionConfigRequest
	if err := ctx.Bind(&req); err != nil {
		httpx.Error(ctx, errors.BadRequest("参数绑定失败: "+err.Error()))
		return
	}

	if err := h.svc.AssignPermissions(ctx.Context(), id, req.MenuIDs, req.APIIDs, req.ButtonIDs); err != nil {
		httpx.Error(ctx, err)
		return
	}

	httpx.Success(ctx, nil)
}
