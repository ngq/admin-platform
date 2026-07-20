// Package iam — IAM 领域子聚合
package iam

import (
	"context"
	"fmt"
	"strconv"

	"github.com/ngq/admin-platform/internal/biz/system"
	"github.com/ngq/admin-platform/internal/errors"
	"golang.org/x/crypto/bcrypt"
)

// Biz IAM 领域子聚合，持有该领域所有仓储
type Biz struct {
	UserRepo       UserRepository
	DepartmentRepo DepartmentRepository
	RoleRepo       RoleRepository
	MenuRepo       MenuRepository
	APIRepo        APIRepository
	ButtonRepo     ButtonRepository
	SysParamRepo   system.SysParamRepository // 密码复杂度校验依赖
}

// NewBiz 创建 IAM 领域子聚合
func NewBiz(
	userRepo UserRepository,
	roleRepo RoleRepository,
	menuRepo MenuRepository,
	apiRepo APIRepository,
	buttonRepo ButtonRepository,
	sysParamRepo system.SysParamRepository,
	departmentRepos ...DepartmentRepository,
) *Biz {
	var departmentRepo DepartmentRepository
	if len(departmentRepos) > 0 {
		departmentRepo = departmentRepos[0]
	}
	return &Biz{
		UserRepo:       userRepo,
		DepartmentRepo: departmentRepo,
		RoleRepo:       roleRepo,
		MenuRepo:       menuRepo,
		APIRepo:        apiRepo,
		ButtonRepo:     buttonRepo,
		SysParamRepo:   sysParamRepo,
	}
}

// ── Department ────────────────────────────────────────────────

func (b *Biz) DepartmentCreate(ctx context.Context, dept *Department) error {
	return b.DepartmentRepo.Create(ctx, dept)
}

func (b *Biz) DepartmentUpdate(ctx context.Context, dept *Department) error {
	return b.DepartmentRepo.Update(ctx, dept)
}

func (b *Biz) DepartmentDelete(ctx context.Context, id int64) error {
	return b.DepartmentRepo.Delete(ctx, id)
}

func (b *Biz) DepartmentGetByID(ctx context.Context, id int64) (*Department, error) {
	return b.DepartmentRepo.GetByID(ctx, id)
}

func (b *Biz) DepartmentList(ctx context.Context, name, status string) ([]*Department, error) {
	depts, err := b.DepartmentRepo.List(ctx, name, status)
	if err != nil {
		return nil, err
	}
	return buildDepartmentTree(depts), nil
}

// ── Menu ──────────────────────────────────────────────────────

// MenuCreate 创建菜单
func (b *Biz) MenuCreate(ctx context.Context, menu *Menu) error {
	return b.MenuRepo.Create(ctx, menu)
}

// MenuUpdate 更新菜单信息
func (b *Biz) MenuUpdate(ctx context.Context, menu *Menu) error {
	return b.MenuRepo.Update(ctx, menu)
}

// MenuDelete 删除菜单
func (b *Biz) MenuDelete(ctx context.Context, id int64) error {
	return b.MenuRepo.Delete(ctx, id)
}

// MenuGetByID 根据 ID 获取菜单详情
func (b *Biz) MenuGetByID(ctx context.Context, id int64) (*Menu, error) {
	return b.MenuRepo.GetByID(ctx, id)
}

// MenuList 分页查询菜单列表
func (b *Biz) MenuList(ctx context.Context, page, pageSize int, name, status string) ([]*Menu, int64, error) {
	return b.MenuRepo.List(ctx, page, pageSize, name, status)
}

// MenuGetTree 获取菜单树形结构
func (b *Biz) MenuGetTree(ctx context.Context) ([]*Menu, error) {
	return b.MenuRepo.GetTree(ctx)
}

// MenuGetByRoleIDs 根据角色 ID 列表获取关联菜单
func (b *Biz) MenuGetByRoleIDs(ctx context.Context, roleIDs []int64) ([]*Menu, error) {
	return b.MenuRepo.GetByRoleIDs(ctx, roleIDs)
}

// MenuGetUserMenus 获取用户菜单树
func (b *Biz) MenuGetUserMenus(ctx context.Context, userID int64) ([]*Menu, error) {
	roles, err := b.RoleRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if len(roles) == 0 {
		return []*Menu{}, nil
	}

	roleIDs := make([]int64, 0, len(roles))
	for _, r := range roles {
		roleIDs = append(roleIDs, r.ID)
	}

	menus, err := b.MenuRepo.GetByRoleIDs(ctx, roleIDs)
	if err != nil {
		return nil, err
	}
	if len(menus) == 0 {
		return []*Menu{}, nil
	}

	return buildUserMenuTree(menus), nil
}

// ── API ───────────────────────────────────────────────────────

// APICreate 创建 API 资源
func (b *Biz) APICreate(ctx context.Context, api *APIResource) error {
	return b.APIRepo.Create(ctx, api)
}

// APIUpdate 更新 API 资源信息
func (b *Biz) APIUpdate(ctx context.Context, api *APIResource) error {
	return b.APIRepo.Update(ctx, api)
}

// APIDelete 删除 API 资源
func (b *Biz) APIDelete(ctx context.Context, id int64) error {
	return b.APIRepo.Delete(ctx, id)
}

// APIGetByID 根据 ID 获取 API 资源详情
func (b *Biz) APIGetByID(ctx context.Context, id int64) (*APIResource, error) {
	return b.APIRepo.GetByID(ctx, id)
}

// APIList 分页查询 API 资源列表
func (b *Biz) APIList(ctx context.Context, page, pageSize int, group, path, method string) ([]*APIResource, int64, error) {
	return b.APIRepo.List(ctx, page, pageSize, group, path, method)
}

// APIGetAll 获取所有 API 资源
func (b *Biz) APIGetAll(ctx context.Context) ([]*APIResource, error) {
	return b.APIRepo.GetAll(ctx)
}

// APIBatchDelete 批量删除 API 资源
func (b *Biz) APIBatchDelete(ctx context.Context, ids []int64) error {
	return b.APIRepo.BatchDelete(ctx, ids)
}

// ── Button ────────────────────────────────────────────────────

// ButtonCreate 创建按钮权限
func (b *Biz) ButtonCreate(ctx context.Context, button *Button) error {
	return b.ButtonRepo.Create(ctx, button)
}

// ButtonUpdate 更新按钮权限信息
func (b *Biz) ButtonUpdate(ctx context.Context, button *Button) error {
	return b.ButtonRepo.Update(ctx, button)
}

// ButtonDelete 删除按钮权限
func (b *Biz) ButtonDelete(ctx context.Context, id int64) error {
	return b.ButtonRepo.Delete(ctx, id)
}

// ButtonGetByMenuID 根据菜单 ID 获取按钮权限列表
func (b *Biz) ButtonGetByMenuID(ctx context.Context, menuID int64) ([]*Button, error) {
	return b.ButtonRepo.GetByMenuID(ctx, menuID)
}

// ButtonGetByRoleIDs 根据角色 ID 列表获取关联的按钮权限
func (b *Biz) ButtonGetByRoleIDs(ctx context.Context, roleIDs []int64) ([]*Button, error) {
	return b.ButtonRepo.GetByRoleIDs(ctx, roleIDs)
}

// ── User ──────────────────────────────────────────────────────

// UserCreate 创建用户（自动哈希密码）
func (b *Biz) UserCreate(ctx context.Context, user *User) error {
	existing, err := b.UserRepo.GetByUsername(ctx, user.Username)
	if err == nil && existing != nil {
		return errors.UserAlreadyExists()
	}

	if err := validatePassword(ctx, b.SysParamRepo, user.Password); err != nil {
		return err
	}

	hashedPwd, err := hashPassword(user.Password)
	if err != nil {
		return errors.InternalError("密码哈希失败")
	}
	user.Password = hashedPwd

	return b.UserRepo.Create(ctx, user)
}

// UserUpdate 更新用户信息（不含角色和密码）
func (b *Biz) UserUpdate(ctx context.Context, user *User) error {
	return b.UserRepo.Update(ctx, user)
}

// UserDelete 删除用户
func (b *Biz) UserDelete(ctx context.Context, id int64) error {
	return b.UserRepo.Delete(ctx, id)
}

// UserGetByID 根据 ID 获取用户详情
func (b *Biz) UserGetByID(ctx context.Context, id int64) (*User, error) {
	user, err := b.UserRepo.GetByID(ctx, id)
	if err != nil {
		return nil, errors.UserNotFound()
	}
	return user, nil
}

// UserList 分页查询用户列表
func (b *Biz) UserList(ctx context.Context, page, pageSize int, username, status string) ([]*User, int64, error) {
	return b.UserRepo.List(ctx, page, pageSize, username, status)
}

// UserListByCurrentUserScope 根据当前登录人的角色 data_scope 过滤用户列表。
func (b *Biz) UserListByCurrentUserScope(ctx context.Context, currentUserID int64, page, pageSize int, username, status string, deptID int64) ([]*User, int64, error) {
	currentUser, err := b.UserRepo.GetByID(ctx, currentUserID)
	if err != nil {
		return nil, 0, errors.UserNotFound()
	}
	roles, err := b.RoleRepo.GetByUserID(ctx, currentUserID)
	if err != nil {
		return nil, 0, err
	}
	dataScope := widestDataScope(roles)
	scope := DataPermissionScope{
		DataScope: dataScope,
		UserID:    currentUserID,
		DeptID:    currentUser.DeptID,
	}
	if dataScope == DataScopeDeptSub && currentUser.DeptID > 0 && b.DepartmentRepo != nil {
		deptIDs, err := b.DepartmentRepo.GetDescendantIDs(ctx, currentUser.DeptID)
		if err != nil {
			return nil, 0, err
		}
		scope.DeptIDs = deptIDs
	}
	if dataScope != DataScopeAll && b.DepartmentRepo != nil {
		managedDeptIDs, err := b.DepartmentRepo.GetManagedDeptIDs(ctx, currentUserID)
		if err != nil {
			return nil, 0, err
		}
		if len(managedDeptIDs) > 0 {
			scope.DataScope = DataScopeDeptSub
			scope.DeptIDs = mergeDeptIDs(scope.DeptIDs, managedDeptIDs)
			for _, deptID := range managedDeptIDs {
				descendantIDs, err := b.DepartmentRepo.GetDescendantIDs(ctx, deptID)
				if err != nil {
					return nil, 0, err
				}
				scope.DeptIDs = mergeDeptIDs(scope.DeptIDs, descendantIDs)
			}
		}
	}
	dataScopedRepo, ok := b.UserRepo.(DataScopedUserRepository)
	if !ok {
		return b.UserRepo.List(ctx, page, pageSize, username, status)
	}
	return dataScopedRepo.ListByDataScope(ctx, page, pageSize, username, status, deptID, scope)
}

// UserUpdateRoles 更新用户角色关联
func (b *Biz) UserUpdateRoles(ctx context.Context, userID int64, roleIDs []int64) error {
	return b.UserRepo.UpdateRoles(ctx, userID, roleIDs)
}

// UserChangePassword 修改密码（校验旧密码后设置新密码）
func (b *Biz) UserChangePassword(ctx context.Context, userID int64, oldPassword, newPassword string) error {
	user, err := b.UserRepo.GetByID(ctx, userID)
	if err != nil {
		return errors.UserNotFound()
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword)); err != nil {
		return errors.UserPasswordWrong("旧密码错误")
	}

	if err := validatePassword(ctx, b.SysParamRepo, newPassword); err != nil {
		return err
	}

	hashedPwd, err := hashPassword(newPassword)
	if err != nil {
		return errors.InternalError("密码哈希失败")
	}

	return b.UserRepo.SetPassword(ctx, userID, hashedPwd)
}

// UserResetPassword 重置密码（管理员操作，校验复杂度但不校验旧密码）
func (b *Biz) UserResetPassword(ctx context.Context, userID int64, newPassword string) error {
	if err := validatePassword(ctx, b.SysParamRepo, newPassword); err != nil {
		return err
	}
	hashedPwd, err := hashPassword(newPassword)
	if err != nil {
		return errors.InternalError("密码哈希失败")
	}
	return b.UserRepo.SetPassword(ctx, userID, hashedPwd)
}

// UserSetPassword 设置用户密码（哈希值）
func (b *Biz) UserSetPassword(ctx context.Context, id int64, hashedPassword string) error {
	return b.UserRepo.SetPassword(ctx, id, hashedPassword)
}

// UserUpdateLastLogin 更新最后登录信息
func (b *Biz) UserUpdateLastLogin(ctx context.Context, id int64, ip string) error {
	return b.UserRepo.UpdateLastLogin(ctx, id, ip)
}

// ── 密码辅助函数 ──────────────────────────────────────────────

// hashPassword 使用 bcrypt 对明文密码进行哈希
func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// validatePassword 校验密码复杂度
func validatePassword(ctx context.Context, sysParamRepo system.SysParamRepository, password string) error {
	minLength := 6
	if sysParamRepo != nil {
		if param, err := sysParamRepo.GetByKey(ctx, "password_min_length"); err == nil {
			if length, err := strconv.Atoi(param.Value); err == nil && length > 0 {
				minLength = length
			}
		}
	}

	if len(password) < minLength {
		return errors.BadRequest(fmt.Sprintf("密码长度不能少于 %d 位", minLength))
	}

	hasLetter := false
	hasDigit := false
	for _, c := range password {
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') {
			hasLetter = true
		}
		if c >= '0' && c <= '9' {
			hasDigit = true
		}
	}
	if !hasLetter || !hasDigit {
		return errors.BadRequest("密码必须同时包含字母和数字")
	}

	return nil
}

// ── 菜单树构建 ────────────────────────────────────────────────

// buildUserMenuTree 根据父 ID 构建菜单树形结构
func buildUserMenuTree(menus []*Menu) []*Menu {
	if len(menus) == 0 {
		return menus
	}

	menuMap := make(map[int64]*Menu, len(menus))
	for _, m := range menus {
		menuMap[m.ID] = m
		m.Children = make([]*Menu, 0)
	}

	roots := make([]*Menu, 0)
	for _, m := range menus {
		if m.ParentID == 0 {
			roots = append(roots, m)
		} else {
			if parent, ok := menuMap[m.ParentID]; ok {
				parent.Children = append(parent.Children, m)
			}
		}
	}

	if len(roots) == 0 {
		return menus
	}

	return roots
}

func buildDepartmentTree(depts []*Department) []*Department {
	if len(depts) == 0 {
		return depts
	}
	deptMap := make(map[int64]*Department, len(depts))
	for _, dept := range depts {
		dept.Children = make([]*Department, 0)
		deptMap[dept.ID] = dept
	}

	roots := make([]*Department, 0)
	for _, dept := range depts {
		if dept.ParentID == 0 {
			roots = append(roots, dept)
			continue
		}
		if parent, ok := deptMap[dept.ParentID]; ok {
			parent.Children = append(parent.Children, dept)
		} else {
			roots = append(roots, dept)
		}
	}
	return roots
}

func mergeDeptIDs(base []int64, extra []int64) []int64 {
	seen := make(map[int64]struct{}, len(base)+len(extra))
	result := make([]int64, 0, len(base)+len(extra))
	for _, id := range append(base, extra...) {
		if id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	return result
}
