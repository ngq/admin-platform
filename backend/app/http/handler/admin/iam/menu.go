// Package handler - 菜单管理处理器
// 使用 gorp 框架契约层 API（gorp.Context / gorp.Success / gorp.BadRequest / httpx.SuccessPaginated）。
package iam

import (
	"github.com/ngq/admin-platform/app/http/handler/common"
	bizIam "github.com/ngq/admin-platform/internal/biz/iam"
	"github.com/ngq/admin-platform/internal/errors"
	"github.com/ngq/gorp"
	"github.com/ngq/gorp/framework/httpx"
)

// createMenuRequest 创建菜单请求 DTO（handler 内部定义，不暴露到 request 包）
type createMenuRequest struct {
	ParentID  int64  `json:"parent_id"`                              // 父菜单 ID
	Name      string `json:"name" validate:"required,max=50"`        // 菜单名称
	Path      string `json:"path" validate:"max=200"`                // 路由路径
	Component string `json:"component" validate:"max=200"`           // 前端组件路径
	Redirect  string `json:"redirect" validate:"max=200"`            // 重定向地址
	Icon      string `json:"icon" validate:"max=50"`                 // 菜单图标
	Sort      int    `json:"sort"`                                   // 排序序号
	Type      int8   `json:"type" validate:"required,oneof=1 2 3"`   // 菜单类型：1-目录 2-菜单 3-按钮
	Visible   int8   `json:"visible" validate:"omitempty,oneof=0 1"` // 是否可见：0-隐藏 1-显示
	Status    int8   `json:"status" validate:"omitempty,oneof=1 2"`  // 状态：1-启用 2-禁用
}

// updateMenuRequest 更新菜单请求 DTO（handler 内部定义，不暴露到 request 包）
type updateMenuRequest struct {
	ParentID  int64  `json:"parent_id"`                              // 父菜单 ID
	Name      string `json:"name" validate:"max=50"`                 // 菜单名称
	Path      string `json:"path" validate:"max=200"`                // 路由路径
	Component string `json:"component" validate:"max=200"`           // 前端组件路径
	Redirect  string `json:"redirect" validate:"max=200"`            // 重定向地址
	Icon      string `json:"icon" validate:"max=50"`                 // 菜单图标
	Sort      int    `json:"sort"`                                   // 排序序号
	Type      int8   `json:"type" validate:"omitempty,oneof=1 2 3"`  // 菜单类型
	Visible   int8   `json:"visible" validate:"omitempty,oneof=0 1"` // 是否可见
	Status    int8   `json:"status" validate:"omitempty,oneof=1 2"`  // 状态
}

// MenuHandler 菜单管理处理器
// 封装菜单 CRUD 与树形查询的 HTTP 入口，委托 MenuService 执行业务逻辑。
type MenuHandler struct {
	biz *bizIam.Biz
}

// NewMenuHandler 创建菜单管理处理器
func NewMenuHandler(biz *bizIam.Biz) *MenuHandler {
	return &MenuHandler{biz: biz}
}

// List 分页查询菜单列表
//
// GET /api/v1/menus
//
// 从 query 参数读取 page / page_size / name / status，
// 委托 service 层执行分页查询，通过 httpx.SuccessPaginated 输出分页响应。
func (h *MenuHandler) List(ctx gorp.Context) {
	pg := common.ParsePagination(ctx, 20, 100)
	name := ctx.Query("name")
	status := ctx.Query("status")

	menus, total, err := h.biz.MenuList(ctx.Context(), pg.Page, pg.PageSize, name, status)
	if err != nil {
		httpx.Error(ctx, err)
		return
	}

	// 分页响应使用 httpx.SuccessPaginated
	httpx.SuccessPaginated(ctx, menus, total, pg.Page, pg.PageSize)
}

// GetTree 获取菜单树形结构
//
// GET /api/v1/menus/tree
//
// 返回完整的菜单树，不区分页。用于前端侧边栏渲染与角色菜单分配。
func (h *MenuHandler) GetTree(ctx gorp.Context) {
	menus, err := h.biz.MenuGetTree(ctx.Context())
	if err != nil {
		httpx.Error(ctx, err)
		return
	}

	httpx.Success(ctx, menus)
}

// GetAll 获取当前用户菜单树
// GET /api/v1/menu/all
// 返回当前用户可访问的菜单树，用于前端路由生成
func (h *MenuHandler) GetAll(ctx gorp.Context) {
	// 从 JWT claims 中获取用户 ID
	claims, ok := gorp.ClaimsFromContext(ctx)
	if !ok || claims == nil {
		httpx.Error(ctx, errors.Unauthorized("未认证"))
		return
	}

	menus, err := h.biz.MenuGetUserMenus(ctx.Context(), claims.SubjectID)
	if err != nil {
		httpx.Error(ctx, err)
		return
	}

	httpx.Success(ctx, menus)
}

// Get 根据 ID 获取菜单详情
//
// GET /api/v1/menus/:id
//
// 路径参数 id 解析失败时返回 bad request，查询成功返回菜单实体。
func (h *MenuHandler) Get(ctx gorp.Context) {
	id, err := ctx.Int64Param("id")
	if err != nil {
		httpx.Error(ctx, errors.BadRequest("无效的菜单 ID"))
		return
	}

	menu, err := h.biz.MenuGetByID(ctx.Context(), id)
	if err != nil {
		httpx.Error(ctx, err)
		return
	}

	httpx.Success(ctx, menu)
}

// Create 创建菜单
//
// POST /api/v1/menus
//
// 使用 ctx.Bind 绑定请求体（自动根据 Content-Type 选择绑定方式），
// 绑定失败返回 bad request，创建成功返回空 data。
func (h *MenuHandler) Create(ctx gorp.Context) {
	var req createMenuRequest
	if err := ctx.Bind(&req); err != nil {
		httpx.Error(ctx, errors.BadRequest("参数绑定失败: "+err.Error()))
		return
	}

	// 将请求 DTO 映射为领域实体，委托 service 执行创建
	menu := &bizIam.Menu{
		ParentID:  req.ParentID,
		Name:      req.Name,
		Path:      req.Path,
		Component: req.Component,
		Redirect:  req.Redirect,
		Icon:      req.Icon,
		Sort:      req.Sort,
		Type:      req.Type,
		Visible:   req.Visible,
		Status:    req.Status,
	}

	if err := h.biz.MenuCreate(ctx.Context(), menu); err != nil {
		httpx.Error(ctx, err)
		return
	}

	httpx.Success(ctx, nil)
}

// Update 更新菜单
//
// POST /api/v1/menus/:id
//
// 路径参数 id 指定更新目标，请求体通过 ctx.Bind 绑定，
// 绑定失败或 ID 无效时返回 bad request。
func (h *MenuHandler) Update(ctx gorp.Context) {
	id, err := ctx.Int64Param("id")
	if err != nil {
		httpx.Error(ctx, errors.BadRequest("无效的菜单 ID"))
		return
	}

	var req updateMenuRequest
	if err := ctx.Bind(&req); err != nil {
		httpx.Error(ctx, errors.BadRequest("参数绑定失败: "+err.Error()))
		return
	}

	// 将路径 ID 与请求体字段合并为领域实体，委托 service 执行更新
	menu := &bizIam.Menu{
		ID:        id,
		ParentID:  req.ParentID,
		Name:      req.Name,
		Path:      req.Path,
		Component: req.Component,
		Redirect:  req.Redirect,
		Icon:      req.Icon,
		Sort:      req.Sort,
		Type:      req.Type,
		Visible:   req.Visible,
		Status:    req.Status,
	}

	if err := h.biz.MenuUpdate(ctx.Context(), menu); err != nil {
		httpx.Error(ctx, err)
		return
	}

	httpx.Success(ctx, nil)
}

// Delete 删除菜单
//
// POST /api/v1/menus/delete/:id
//
// 路径参数 id 解析失败时返回 bad request，删除成功返回空 data。
func (h *MenuHandler) Delete(ctx gorp.Context) {
	id, err := ctx.Int64Param("id")
	if err != nil {
		httpx.Error(ctx, errors.BadRequest("无效的菜单 ID"))
		return
	}

	if err := h.biz.MenuDelete(ctx.Context(), id); err != nil {
		httpx.Error(ctx, err)
		return
	}

	httpx.Success(ctx, nil)
}
