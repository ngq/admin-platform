// 角色仓储实现
// 基于 GORM 实现 iam.RoleRepository 接口
package iam

import (
	"context"
	"github.com/ngq/admin-platform/internal/datainfra"

	"github.com/ngq/admin-platform/internal/biz/iam"
	"github.com/ngq/admin-platform/internal/data/po"
	"gorm.io/gorm"
)

// roleRepo 角色仓储实现
type roleRepo struct {
	data *datainfra.Data
}

// NewRoleRepo 创建角色仓储实例
func NewRoleRepo(d *datainfra.Data) iam.RoleRepository {
	return &roleRepo{data: d}
}

// Create 创建角色
func (r *roleRepo) Create(ctx context.Context, role *iam.Role) error {
	rolePO := po.RolePOFromEntity(role)
	return r.data.DB(ctx).Create(rolePO).Error
}

// Update 更新角色基本信息
func (r *roleRepo) Update(ctx context.Context, role *iam.Role) error {
	rolePO := po.RolePOFromEntity(role)
	return r.data.DB(ctx).Model(&po.RolePO{}).
		Where("id = ?", role.ID).
		Updates(map[string]interface{}{
			"name":        rolePO.Name,
			"code":        rolePO.Code,
			"parent_id":   rolePO.ParentID,
			"sort":        rolePO.Sort,
			"data_scope":  rolePO.DataScope,
			"status":      rolePO.Status,
			"description": rolePO.Description,
		}).Error
}

// Delete 删除角色，同时清理关联表数据
func (r *roleRepo) Delete(ctx context.Context, id int64) error {
	return r.data.DB(ctx).Transaction(func(tx *gorm.DB) error {
		// 删除角色-菜单关联
		if err := tx.Table("sys_role_menus").Where("role_id = ?", id).Delete(nil).Error; err != nil {
			return err
		}
		// 删除角色-API关联
		if err := tx.Table("sys_role_apis").Where("role_id = ?", id).Delete(nil).Error; err != nil {
			return err
		}
		// 删除角色-按钮关联
		if err := tx.Table("sys_role_buttons").Where("role_id = ?", id).Delete(nil).Error; err != nil {
			return err
		}
		// 删除角色记录
		return tx.Delete(&po.RolePO{}, id).Error
	})
}

// GetByID 根据ID获取角色
func (r *roleRepo) GetByID(ctx context.Context, id int64) (*iam.Role, error) {
	var rolePO po.RolePO
	if err := r.data.DB(ctx).First(&rolePO, id).Error; err != nil {
		return nil, err
	}
	return rolePO.ToEntity(), nil
}

// GetByCode 根据编码获取角色
func (r *roleRepo) GetByCode(ctx context.Context, code string) (*iam.Role, error) {
	var rolePO po.RolePO
	if err := r.data.DB(ctx).Where("code = ?", code).First(&rolePO).Error; err != nil {
		return nil, err
	}
	return rolePO.ToEntity(), nil
}

// List 分页查询角色列表，支持按名称/编码模糊搜索和状态过滤
func (r *roleRepo) List(ctx context.Context, page, pageSize int, name, code, status string) ([]*iam.Role, int64, error) {
	var list []po.RolePO
	var count int64

	db := r.data.DB(ctx).Model(&po.RolePO{})

	// 按名称模糊搜索
	if name != "" {
		db = db.Where("name LIKE ?", "%"+name+"%")
	}
	// 按编码模糊搜索
	if code != "" {
		db = db.Where("code LIKE ?", "%"+code+"%")
	}
	// 按状态过滤
	if status != "" {
		db = db.Where("status = ?", status)
	}

	// 分页查询
	if err := datainfra.Paginate(db, page, pageSize, "sort ASC, id DESC", &list, &count); err != nil {
		return nil, 0, err
	}

	// 转换为领域实体
	roles := make([]*iam.Role, 0, len(list))
	for i := range list {
		roles = append(roles, list[i].ToEntity())
	}
	return roles, count, nil
}

// GetByUserID 根据用户ID获取其关联角色列表
// 通过 sys_user_roles 关联表查询
func (r *roleRepo) GetByUserID(ctx context.Context, userID int64) ([]*iam.Role, error) {
	// 先查询用户关联的角色ID
	var roleIDs []int64
	if err := r.data.DB(ctx).
		Table("sys_user_roles").
		Where("user_id = ?", userID).
		Pluck("role_id", &roleIDs).Error; err != nil {
		return nil, err
	}

	if len(roleIDs) == 0 {
		return nil, nil
	}

	// 查询角色详情（只查询正常状态的角色）
	var list []po.RolePO
	if err := r.data.DB(ctx).
		Where("id IN ? AND status = 1", roleIDs).
		Find(&list).Error; err != nil {
		return nil, err
	}

	roles := make([]*iam.Role, 0, len(list))
	for i := range list {
		roles = append(roles, list[i].ToEntity())
	}
	return roles, nil
}

// UpdateMenus 更新角色关联的菜单ID列表
// 先删除旧关联，再批量插入新关联
func (r *roleRepo) UpdateMenus(ctx context.Context, roleID int64, menuIDs []int64) error {
	return r.data.DB(ctx).Transaction(func(tx *gorm.DB) error {
		// 删除旧的角色-菜单关联
		if err := tx.Table("sys_role_menus").Where("role_id = ?", roleID).Delete(nil).Error; err != nil {
			return err
		}
		// 批量插入新关联
		if len(menuIDs) > 0 {
			records := make([]map[string]interface{}, 0, len(menuIDs))
			for _, menuID := range menuIDs {
				records = append(records, map[string]interface{}{
					"role_id": roleID,
					"menu_id": menuID,
				})
			}
			return tx.Table("sys_role_menus").Create(records).Error
		}
		return nil
	})
}

// UpdateAPIs 更新角色关联的API ID列表
func (r *roleRepo) UpdateAPIs(ctx context.Context, roleID int64, apiIDs []int64) error {
	return r.data.DB(ctx).Transaction(func(tx *gorm.DB) error {
		// 删除旧的角色-API关联
		if err := tx.Table("sys_role_apis").Where("role_id = ?", roleID).Delete(nil).Error; err != nil {
			return err
		}
		// 批量插入新关联
		if len(apiIDs) > 0 {
			records := make([]map[string]interface{}, 0, len(apiIDs))
			for _, apiID := range apiIDs {
				records = append(records, map[string]interface{}{
					"role_id": roleID,
					"api_id":  apiID,
				})
			}
			return tx.Table("sys_role_apis").Create(records).Error
		}
		return nil
	})
}

// UpdateButtons 更新角色关联的按钮ID列表
func (r *roleRepo) UpdateButtons(ctx context.Context, roleID int64, buttonIDs []int64) error {
	return r.data.DB(ctx).Transaction(func(tx *gorm.DB) error {
		// 删除旧的角色-按钮关联
		if err := tx.Table("sys_role_buttons").Where("role_id = ?", roleID).Delete(nil).Error; err != nil {
			return err
		}
		// 批量插入新关联
		if len(buttonIDs) > 0 {
			records := make([]map[string]interface{}, 0, len(buttonIDs))
			for _, buttonID := range buttonIDs {
				records = append(records, map[string]interface{}{
					"role_id":   roleID,
					"button_id": buttonID,
				})
			}
			return tx.Table("sys_role_buttons").Create(records).Error
		}
		return nil
	})
}

// UpdatePermissions 在一个事务内更新角色的菜单、API 和按钮权限。
func (r *roleRepo) UpdatePermissions(ctx context.Context, roleID int64, menuIDs []int64, apiIDs []int64, buttonIDs []int64) error {
	return r.data.DB(ctx).Transaction(func(tx *gorm.DB) error {
		if err := replaceRoleRelations(tx, "sys_role_menus", "menu_id", roleID, menuIDs); err != nil {
			return err
		}
		if err := replaceRoleRelations(tx, "sys_role_apis", "api_id", roleID, apiIDs); err != nil {
			return err
		}
		return replaceRoleRelations(tx, "sys_role_buttons", "button_id", roleID, buttonIDs)
	})
}

func replaceRoleRelations(tx *gorm.DB, table string, targetColumn string, roleID int64, targetIDs []int64) error {
	if err := tx.Table(table).Where("role_id = ?", roleID).Delete(nil).Error; err != nil {
		return err
	}
	if len(targetIDs) == 0 {
		return nil
	}

	records := make([]map[string]interface{}, 0, len(targetIDs))
	for _, targetID := range targetIDs {
		records = append(records, map[string]interface{}{
			"role_id":    roleID,
			targetColumn: targetID,
		})
	}
	return tx.Table(table).Create(records).Error
}

// GetMenuIDs 获取角色关联的菜单ID列表
func (r *roleRepo) GetMenuIDs(ctx context.Context, roleID int64) ([]int64, error) {
	var ids []int64
	err := r.data.DB(ctx).
		Table("sys_role_menus").
		Where("role_id = ?", roleID).
		Pluck("menu_id", &ids).Error
	return ids, err
}

// GetAPIIDs 获取角色关联的API ID列表
func (r *roleRepo) GetAPIIDs(ctx context.Context, roleID int64) ([]int64, error) {
	var ids []int64
	err := r.data.DB(ctx).
		Table("sys_role_apis").
		Where("role_id = ?", roleID).
		Pluck("api_id", &ids).Error
	return ids, err
}

// GetButtonIDs 获取角色关联的按钮ID列表
func (r *roleRepo) GetButtonIDs(ctx context.Context, roleID int64) ([]int64, error) {
	var ids []int64
	err := r.data.DB(ctx).
		Table("sys_role_buttons").
		Where("role_id = ?", roleID).
		Pluck("button_id", &ids).Error
	return ids, err
}

// Count 查询角色总数（用于仪表盘统计，避免 List 的额外开销）
func (r *roleRepo) Count(ctx context.Context) (int64, error) {
	var count int64
	return count, r.data.DB(ctx).Model(&po.RolePO{}).Count(&count).Error
}
