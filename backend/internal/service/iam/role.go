// Package service - 角色管理服务
// 封装角色 CRUD、菜单/API/按钮权限分配等业务逻辑
package iam

import (
	"context"

	"github.com/casbin/casbin/v2"
	bizIam "github.com/ngq/admin-platform/internal/biz/iam"
	"github.com/ngq/admin-platform/internal/errors"
)

// RoleService 角色管理服务
type RoleService struct {
	iamBiz *bizIam.Biz
	casbin *casbin.SyncedEnforcer
}

type RolePermissionConfig struct {
	Menus             []*bizIam.Menu        `json:"menus"`
	APIs              []*bizIam.APIResource `json:"apis"`
	Buttons           []*bizIam.Button      `json:"buttons"`
	SelectedMenuIDs   []int64               `json:"selected_menu_ids"`
	SelectedAPIIDs    []int64               `json:"selected_api_ids"`
	SelectedButtonIDs []int64               `json:"selected_button_ids"`
}

// NewRoleService 创建角色管理服务
func NewRoleService(iamBiz *bizIam.Biz, casbin *casbin.SyncedEnforcer) *RoleService {
	return &RoleService{iamBiz: iamBiz, casbin: casbin}
}

// Create 创建角色
func (s *RoleService) Create(ctx context.Context, role *bizIam.Role) error {
	// 检查角色编码是否已存在
	existing, err := s.iamBiz.RoleRepo.GetByCode(ctx, role.Code)
	if err == nil && existing != nil {
		return errors.Conflict("角色编码已存在")
	}
	return s.iamBiz.RoleRepo.Create(ctx, role)
}

// Update 更新角色信息
func (s *RoleService) Update(ctx context.Context, role *bizIam.Role) error {
	return s.iamBiz.RoleRepo.Update(ctx, role)
}

// Delete 删除角色
// 同时删除数据库记录、关联表数据和 Casbin 策略
func (s *RoleService) Delete(ctx context.Context, id int64) error {
	// 1. 获取角色信息（用于清理 Casbin 策略）
	role, err := s.iamBiz.RoleRepo.GetByID(ctx, id)
	if err != nil {
		return errors.NotFound("角色不存在")
	}

	// 2. 删除数据库记录和关联表
	if err := s.iamBiz.RoleRepo.Delete(ctx, id); err != nil {
		return err
	}

	// 3. 清理 Casbin 策略
	// 如果 Casbin 清理失败，记录错误但不回滚数据库删除（Casbin 不支持事务回滚）
	if s.casbin != nil {
		s.casbin.RemoveFilteredPolicy(0, role.Code)
		if err := s.casbin.SavePolicy(); err != nil {
			// 策略持久化失败不影响删除结果，但可能导致权限不一致
			// 后续 SyncAllCasbinPolicies 可以修复不一致
			_ = err
		}
	}

	return nil
}

// GetByID 根据 ID 获取角色详情
func (s *RoleService) GetByID(ctx context.Context, id int64) (*bizIam.Role, error) {
	role, err := s.iamBiz.RoleRepo.GetByID(ctx, id)
	if err != nil {
		return nil, errors.NotFound("角色不存在")
	}
	return role, nil
}

// List 分页查询角色列表
func (s *RoleService) List(ctx context.Context, page, pageSize int, name, code, status string) ([]*bizIam.Role, int64, error) {
	return s.iamBiz.RoleRepo.List(ctx, page, pageSize, name, code, status)
}

// GetByUserID 根据用户 ID 获取其关联角色列表
func (s *RoleService) GetByUserID(ctx context.Context, userID int64) ([]*bizIam.Role, error) {
	return s.iamBiz.RoleRepo.GetByUserID(ctx, userID)
}

// AssignMenus 分配角色菜单权限
func (s *RoleService) AssignMenus(ctx context.Context, roleID int64, menuIDs []int64) error {
	return s.iamBiz.RoleRepo.UpdateMenus(ctx, roleID, menuIDs)
}

// AssignAPIs 分配角色 API 权限
// 同时更新 sys_role_apis 表和 Casbin 策略，确保权限校验生效
func (s *RoleService) AssignAPIs(ctx context.Context, roleID int64, apiIDs []int64) error {
	// 1. 获取角色信息（用于 Casbin 策略的 sub 字段）
	role, err := s.iamBiz.RoleRepo.GetByID(ctx, roleID)
	if err != nil {
		return errors.NotFound("角色不存在")
	}

	// 2. 更新 sys_role_apis 关联表
	if err := s.iamBiz.RoleRepo.UpdateAPIs(ctx, roleID, apiIDs); err != nil {
		return err
	}

	// 3. 同步 Casbin 策略
	if s.casbin != nil {
		// 清除该角色的旧策略
		s.casbin.RemoveFilteredPolicy(0, role.Code)

		// 添加新策略
		if len(apiIDs) > 0 {
			// 根据 ID 列表获取 API 详情
			apis, err := s.iamBiz.APIRepo.GetByIDs(ctx, apiIDs)
			if err == nil && len(apis) > 0 {
				for _, api := range apis {
					if api.Ignored == 1 || (api.PermissionMode != "" && api.PermissionMode != "casbin") {
						continue
					}
					// Casbin 策略格式: (role_code, path, method)
					s.casbin.AddPolicy(role.Code, api.Path, api.Method)
				}
			}
		}

		// 持久化策略到数据库
		s.casbin.SavePolicy()
	}

	return nil
}

// AssignButtons 分配角色按钮权限
func (s *RoleService) AssignButtons(ctx context.Context, roleID int64, buttonIDs []int64) error {
	return s.iamBiz.RoleRepo.UpdateButtons(ctx, roleID, buttonIDs)
}

func (s *RoleService) GetPermissionConfig(ctx context.Context, roleID int64) (*RolePermissionConfig, error) {
	if _, err := s.iamBiz.RoleRepo.GetByID(ctx, roleID); err != nil {
		return nil, errors.NotFound("角色不存在")
	}

	menus, err := s.iamBiz.MenuRepo.GetTree(ctx)
	if err != nil {
		return nil, err
	}
	apis, err := s.iamBiz.APIRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	buttons, err := s.getButtonsForMenus(ctx, menus)
	if err != nil {
		return nil, err
	}
	menuIDs, err := s.iamBiz.RoleRepo.GetMenuIDs(ctx, roleID)
	if err != nil {
		return nil, err
	}
	apiIDs, err := s.iamBiz.RoleRepo.GetAPIIDs(ctx, roleID)
	if err != nil {
		return nil, err
	}
	buttonIDs, err := s.iamBiz.RoleRepo.GetButtonIDs(ctx, roleID)
	if err != nil {
		return nil, err
	}

	return &RolePermissionConfig{
		Menus:             menus,
		APIs:              apis,
		Buttons:           buttons,
		SelectedMenuIDs:   menuIDs,
		SelectedAPIIDs:    apiIDs,
		SelectedButtonIDs: buttonIDs,
	}, nil
}

func (s *RoleService) AssignPermissions(ctx context.Context, roleID int64, menuIDs []int64, apiIDs []int64, buttonIDs []int64) error {
	role, err := s.iamBiz.RoleRepo.GetByID(ctx, roleID)
	if err != nil {
		return errors.NotFound("角色不存在")
	}
	if err := s.iamBiz.RoleRepo.UpdatePermissions(ctx, roleID, menuIDs, apiIDs, buttonIDs); err != nil {
		return err
	}
	return s.syncRoleAPIPolicies(ctx, role, apiIDs)
}

func (s *RoleService) getButtonsForMenus(ctx context.Context, menus []*bizIam.Menu) ([]*bizIam.Button, error) {
	buttons := make([]*bizIam.Button, 0)
	var walk func([]*bizIam.Menu) error
	walk = func(items []*bizIam.Menu) error {
		for _, menu := range items {
			list, err := s.iamBiz.ButtonRepo.GetByMenuID(ctx, menu.ID)
			if err != nil {
				return err
			}
			buttons = append(buttons, list...)
			if len(menu.Children) > 0 {
				if err := walk(menu.Children); err != nil {
					return err
				}
			}
		}
		return nil
	}
	if err := walk(menus); err != nil {
		return nil, err
	}
	return buttons, nil
}

// GetMenuIDs 获取角色关联的菜单 ID 列表
func (s *RoleService) GetMenuIDs(ctx context.Context, roleID int64) ([]int64, error) {
	return s.iamBiz.RoleRepo.GetMenuIDs(ctx, roleID)
}

// GetAPIIDs 获取角色关联的 API 资源 ID 列表
func (s *RoleService) GetAPIIDs(ctx context.Context, roleID int64) ([]int64, error) {
	return s.iamBiz.RoleRepo.GetAPIIDs(ctx, roleID)
}

// GetButtonIDs 获取角色关联的按钮权限 ID 列表
func (s *RoleService) GetButtonIDs(ctx context.Context, roleID int64) ([]int64, error) {
	return s.iamBiz.RoleRepo.GetButtonIDs(ctx, roleID)
}

func (s *RoleService) syncRoleAPIPolicies(ctx context.Context, role *bizIam.Role, apiIDs []int64) error {
	if s.casbin == nil {
		return nil
	}

	s.casbin.RemoveFilteredPolicy(0, role.Code)
	if len(apiIDs) > 0 {
		apis, err := s.iamBiz.APIRepo.GetByIDs(ctx, apiIDs)
		if err == nil && len(apis) > 0 {
			for _, api := range apis {
				if api.Ignored == 1 || (api.PermissionMode != "" && api.PermissionMode != "casbin") {
					continue
				}
				s.casbin.AddPolicy(role.Code, api.Path, api.Method)
			}
		}
	}
	s.casbin.SavePolicy()
	return nil
}

// SyncAllCasbinPolicies 启动时同步所有 Casbin 策略
// 从 sys_role_apis + sys_apis 加载策略到 Casbin，确保权限校验生效
// 使用 AddPolicies 批量添加策略，替代逐条 AddPolicy，提升性能
func (s *RoleService) SyncAllCasbinPolicies(ctx context.Context) error {
	if s.casbin == nil {
		return nil
	}

	// 1. 获取所有角色
	roles, _, err := s.iamBiz.RoleRepo.List(ctx, 1, 1000, "", "", "")
	if err != nil {
		return err
	}

	// 2. 清除所有旧策略
	s.casbin.ClearPolicy()

	// 3. 批量收集所有策略，然后一次性写入
	var allPolicies [][]string
	for _, role := range roles {
		if role.Status != 1 {
			continue
		}

		apiIDs, err := s.iamBiz.RoleRepo.GetAPIIDs(ctx, role.ID)
		if err != nil || len(apiIDs) == 0 {
			continue
		}

		apis, err := s.iamBiz.APIRepo.GetByIDs(ctx, apiIDs)
		if err != nil || len(apis) == 0 {
			continue
		}

		for _, api := range apis {
			if api.Ignored == 1 || (api.PermissionMode != "" && api.PermissionMode != "casbin") {
				continue
			}
			allPolicies = append(allPolicies, []string{role.Code, api.Path, api.Method})
		}
	}

	// 批量添加策略（比逐条 AddPolicy 高效得多）
	if len(allPolicies) > 0 {
		_, _ = s.casbin.AddPolicies(allPolicies)
	}

	// 4. 持久化到数据库
	return s.casbin.SavePolicy()
}
