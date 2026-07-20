// 按钮权限仓储实现
// 基于 GORM 实现 iam.ButtonRepository 接口
package iam

import (
	"context"
	"github.com/ngq/admin-platform/internal/datainfra"

	"github.com/ngq/admin-platform/internal/biz/iam"
	"github.com/ngq/admin-platform/internal/data/po"
)

// buttonRepo 按钮权限仓储实现
type buttonRepo struct {
	data *datainfra.Data
}

// NewButtonRepo 创建按钮权限仓储实例
func NewButtonRepo(d *datainfra.Data) iam.ButtonRepository {
	return &buttonRepo{data: d}
}

// Create 创建按钮权限
func (r *buttonRepo) Create(ctx context.Context, button *iam.Button) error {
	buttonPO := po.ButtonPOFromEntity(button)
	return r.data.DB(ctx).Create(buttonPO).Error
}

// Update 更新按钮权限
func (r *buttonRepo) Update(ctx context.Context, button *iam.Button) error {
	buttonPO := po.ButtonPOFromEntity(button)
	return r.data.DB(ctx).Model(&po.ButtonPO{}).
		Where("id = ?", button.ID).
		Updates(map[string]interface{}{
			"menu_id":     buttonPO.MenuID,
			"name":        buttonPO.Name,
			"code":        buttonPO.Code,
			"description": buttonPO.Description,
		}).Error
}

// Delete 删除按钮权限
func (r *buttonRepo) Delete(ctx context.Context, id int64) error {
	return r.data.DB(ctx).Delete(&po.ButtonPO{}, id).Error
}

// GetByMenuID 根据菜单 ID 获取按钮权限列表
func (r *buttonRepo) GetByMenuID(ctx context.Context, menuID int64) ([]*iam.Button, error) {
	var list []po.ButtonPO
	if err := r.data.DB(ctx).
		Where("menu_id = ?", menuID).
		Order("id ASC").
		Find(&list).Error; err != nil {
		return nil, err
	}

	buttons := make([]*iam.Button, 0, len(list))
	for i := range list {
		buttons = append(buttons, list[i].ToEntity())
	}
	return buttons, nil
}

// GetByRoleIDs 根据角色 ID 列表获取关联的按钮权限
func (r *buttonRepo) GetByRoleIDs(ctx context.Context, roleIDs []int64) ([]*iam.Button, error) {
	if len(roleIDs) == 0 {
		return nil, nil
	}

	// 先查询角色关联的按钮 ID（去重）
	var buttonIDs []int64
	if err := r.data.DB(ctx).
		Table("sys_role_buttons").
		Where("role_id IN ?", roleIDs).
		Pluck("DISTINCT button_id", &buttonIDs).Error; err != nil {
		return nil, err
	}

	if len(buttonIDs) == 0 {
		return nil, nil
	}

	// 查询按钮详情
	var list []po.ButtonPO
	if err := r.data.DB(ctx).
		Where("id IN ?", buttonIDs).
		Find(&list).Error; err != nil {
		return nil, err
	}

	buttons := make([]*iam.Button, 0, len(list))
	for i := range list {
		buttons = append(buttons, list[i].ToEntity())
	}
	return buttons, nil
}
