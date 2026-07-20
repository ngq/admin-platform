// Package handler - 用户管理处理器
// 处理用户 CRUD、密码管理、角色分配等 HTTP 请求
package iam

import (
	"strconv"

	"github.com/ngq/admin-platform/app/http/handler/common"
	bizIam "github.com/ngq/admin-platform/internal/biz/iam"
	"github.com/ngq/admin-platform/internal/errors"
	"github.com/ngq/gorp"
	"github.com/ngq/gorp/framework/httpx"
)

// CreateUserRequest 创建用户请求
type CreateUserRequest struct {
	Username string  `json:"username" binding:"required"`
	Password string  `json:"password" binding:"required,min=6"`
	Nickname string  `json:"nickname"`
	DeptID   int64   `json:"dept_id"`
	Email    string  `json:"email"`
	Phone    string  `json:"phone"`
	Avatar   string  `json:"avatar"`
	Status   int8    `json:"status"`
	RoleIDs  []int64 `json:"roleIds"`
}

// UpdateUserRequest 更新用户请求
type UpdateUserRequest struct {
	Nickname string `json:"nickname"`
	DeptID   int64  `json:"dept_id"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Avatar   string `json:"avatar"`
	Status   int8   `json:"status"`
}

// ChangePasswordRequest 修改密码请求（用户本人）
type ChangePasswordRequest struct {
	OldPassword string `json:"oldPassword" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required,min=6"`
}

// ResetPasswordRequest 重置密码请求（管理员）
type ResetPasswordRequest struct {
	NewPassword string `json:"newPassword" binding:"required,min=6"`
}

// AssignRolesRequest 分配角色请求
type AssignRolesRequest struct {
	RoleIDs []int64 `json:"roleIds" binding:"required"`
}

// UserHandler 用户管理处理器
type UserHandler struct {
	biz *bizIam.Biz
}

// NewUserHandler 创建用户管理处理器
func NewUserHandler(biz *bizIam.Biz) *UserHandler {
	return &UserHandler{biz: biz}
}

// List 分页查询用户列表
// GET /api/v1/users
func (h *UserHandler) List(ctx gorp.Context) {
	pg := common.ParsePagination(ctx, 20, 100)
	username := ctx.Query("username")
	status := ctx.Query("status")
	deptID, _ := strconv.ParseInt(firstNonEmpty(ctx.Query("dept_id"), ctx.Query("deptId")), 10, 64)

	claims, ok := gorp.ClaimsFromContext(ctx)
	if !ok || claims == nil {
		httpx.Error(ctx, errors.Unauthorized("未认证"))
		return
	}

	users, total, err := h.biz.UserListByCurrentUserScope(ctx.Context(), claims.SubjectID, pg.Page, pg.PageSize, username, status, deptID)
	if err != nil {
		httpx.Error(ctx, err)
		return
	}

	httpx.SuccessPaginated(ctx, users, total, pg.Page, pg.PageSize)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

// Get 获取用户详情
// GET /api/v1/users/:id
func (h *UserHandler) Get(ctx gorp.Context) {
	id, err := ctx.Int64Param("id")
	if err != nil {
		httpx.Error(ctx, errors.BadRequest("无效的用户 ID"))
		return
	}

	user, err := h.biz.UserGetByID(ctx.Context(), id)
	if err != nil {
		httpx.Error(ctx, err)
		return
	}

	httpx.Success(ctx, user)
}

// Create 创建用户
// POST /api/v1/users
func (h *UserHandler) Create(ctx gorp.Context) {
	var req CreateUserRequest
	if err := ctx.Bind(&req); err != nil {
		httpx.Error(ctx, errors.BadRequest("请求参数错误"))
		return
	}

	user := &bizIam.User{
		Username: req.Username,
		Password: req.Password, // 明文密码，由 service.Create 自动哈希化
		Nickname: req.Nickname,
		DeptID:   req.DeptID,
		Email:    req.Email,
		Phone:    req.Phone,
		Avatar:   req.Avatar,
		Status:   req.Status,
	}
	if user.Status == 0 {
		user.Status = bizIam.UserStatusNormal
	}

	if err := h.biz.UserCreate(ctx.Context(), user); err != nil {
		httpx.Error(ctx, err)
		return
	}
	if len(req.RoleIDs) > 0 {
		if err := h.biz.UserUpdateRoles(ctx.Context(), user.ID, req.RoleIDs); err != nil {
			httpx.Error(ctx, err)
			return
		}
	}

	httpx.Success(ctx, user)
}

// Update 更新用户
// POST /api/v1/users/:id
func (h *UserHandler) Update(ctx gorp.Context) {
	id, err := ctx.Int64Param("id")
	if err != nil {
		httpx.Error(ctx, errors.BadRequest("无效的用户 ID"))
		return
	}

	var req UpdateUserRequest
	if err := ctx.Bind(&req); err != nil {
		httpx.Error(ctx, errors.BadRequest("请求参数错误"))
		return
	}

	user := &bizIam.User{
		ID:       id,
		Nickname: req.Nickname,
		DeptID:   req.DeptID,
		Email:    req.Email,
		Phone:    req.Phone,
		Avatar:   req.Avatar,
		Status:   req.Status,
	}

	if err := h.biz.UserUpdate(ctx.Context(), user); err != nil {
		httpx.Error(ctx, err)
		return
	}

	httpx.Success(ctx, nil)
}

// Delete 删除用户
// POST /api/v1/users/delete/:id
func (h *UserHandler) Delete(ctx gorp.Context) {
	id, err := ctx.Int64Param("id")
	if err != nil {
		httpx.Error(ctx, errors.BadRequest("无效的用户 ID"))
		return
	}

	if err := h.biz.UserDelete(ctx.Context(), id); err != nil {
		httpx.Error(ctx, err)
		return
	}

	httpx.Success(ctx, nil)
}

// UpdatePassword 重置用户密码
// POST /api/v1/users/password?id=xxx
// 仅限管理员使用，用户修改自己密码请用 POST /api/v1/auth/password
func (h *UserHandler) UpdatePassword(ctx gorp.Context) {
	id, err := strconv.ParseInt(ctx.Query("id"), 10, 64)
	if err != nil {
		httpx.Error(ctx, errors.BadRequest("无效的用户 ID"))
		return
	}

	var req ResetPasswordRequest
	if err := ctx.Bind(&req); err != nil {
		httpx.Error(ctx, errors.BadRequest("请求参数错误"))
		return
	}

	// 校验管理员权限（Casbin 已在路由层校验，此处为双重保障）
	claims, ok := gorp.ClaimsFromContext(ctx)
	if !ok || claims == nil {
		httpx.Error(ctx, errors.Unauthorized("未认证"))
		return
	}
	isAdmin := false
	for _, role := range claims.Roles {
		if role == "super-admin" || role == "admin" {
			isAdmin = true
			break
		}
	}
	if !isAdmin {
		httpx.Error(ctx, errors.Forbidden("无权限重置他人密码"))
		return
	}

	// 管理员重置密码，无需校验旧密码
	if err := h.biz.UserResetPassword(ctx.Context(), id, req.NewPassword); err != nil {
		httpx.Error(ctx, err)
		return
	}
	httpx.Success(ctx, nil)
}

// AssignRoles 分配角色
// POST /api/v1/users/roles?id=xxx
func (h *UserHandler) AssignRoles(ctx gorp.Context) {
	id, err := strconv.ParseInt(ctx.Query("id"), 10, 64)
	if err != nil {
		httpx.Error(ctx, errors.BadRequest("无效的用户 ID"))
		return
	}

	var req AssignRolesRequest
	if err := ctx.Bind(&req); err != nil {
		httpx.Error(ctx, errors.BadRequest("请求参数错误"))
		return
	}

	if err := h.biz.UserUpdateRoles(ctx.Context(), id, req.RoleIDs); err != nil {
		httpx.Error(ctx, err)
		return
	}

	httpx.Success(ctx, nil)
}
