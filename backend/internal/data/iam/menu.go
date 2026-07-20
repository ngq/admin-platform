// 菜单仓储实现
// 基于 GORM 实现 iam.MenuRepository 接口
// 包含树形结构构建逻辑（buildMenuTree）
package iam

import (
	"context"
	"github.com/ngq/admin-platform/internal/datainfra"

	"github.com/ngq/admin-platform/internal/biz/iam"
	"github.com/ngq/admin-platform/internal/data/po"
	"gorm.io/gorm"
)

// menuRepo 菜单仓储实现
type menuRepo struct {
	data *datainfra.Data
}

// NewMenuRepo 创建菜单仓储实例
func NewMenuRepo(d *datainfra.Data) iam.MenuRepository {
	return &menuRepo{data: d}
}

// Create 创建菜单
func (r *menuRepo) Create(ctx context.Context, menu *iam.Menu) error {
	menuPO := po.MenuPOFromEntity(menu)
	if err := r.data.DB(ctx).Create(menuPO).Error; err != nil {
		return err
	}
	menu.ID = menuPO.ID
	return nil
}

// Update 更新菜单
func (r *menuRepo) Update(ctx context.Context, menu *iam.Menu) error {
	menuPO := po.MenuPOFromEntity(menu)
	return r.data.DB(ctx).Model(&po.MenuPO{}).
		Where("id = ?", menu.ID).
		Updates(map[string]interface{}{
			"parent_id": menuPO.ParentID,
			"name":      menuPO.Name,
			"path":      menuPO.Path,
			"component": menuPO.Component,
			"redirect":  menuPO.Redirect,
			"icon":      menuPO.Icon,
			"sort":      menuPO.Sort,
			"type":      menuPO.Type,
			"visible":   menuPO.Visible,
			"status":    menuPO.Status,
		}).Error
}

// Delete 删除菜单（级联删除子菜单）
func (r *menuRepo) Delete(ctx context.Context, id int64) error {
	return r.data.DB(ctx).Transaction(func(tx *gorm.DB) error {
		// 递归收集所有需要删除的菜单ID（包含自身和所有子菜单）
		ids := []int64{id}
		if err := collectMenuIDs(tx, id, &ids); err != nil {
			return err
		}
		// 删除角色-菜单关联
		if err := tx.Exec("DELETE FROM sys_role_menus WHERE menu_id IN ?", ids).Error; err != nil {
			return err
		}
		// 批量删除菜单
		return tx.Where("id IN ?", ids).Delete(&po.MenuPO{}).Error
	})
}

// collectMenuIDs 一次性查询收集所有子菜单ID（优化 N+1 问题）
func collectMenuIDs(tx *gorm.DB, parentID int64, ids *[]int64) error {
	// 一次性查询所有菜单的 parent_id 关系
	var allMenus []struct {
		ID       int64
		ParentID int64
	}
	if err := tx.Model(&po.MenuPO{}).Select("id, parent_id").Find(&allMenus).Error; err != nil {
		return err
	}

	// 构建 parent -> children 映射
	childrenMap := make(map[int64][]int64)
	for _, m := range allMenus {
		childrenMap[m.ParentID] = append(childrenMap[m.ParentID], m.ID)
	}

	// 使用 BFS 收集所有子菜单
	queue := []int64{parentID}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		children := childrenMap[current]
		for _, cid := range children {
			*ids = append(*ids, cid)
			queue = append(queue, cid)
		}
	}
	return nil
}

// GetByID 根据ID获取菜单
func (r *menuRepo) GetByID(ctx context.Context, id int64) (*iam.Menu, error) {
	var menuPO po.MenuPO
	if err := r.data.DB(ctx).First(&menuPO, id).Error; err != nil {
		return nil, err
	}
	return menuPO.ToEntity(), nil
}

// List 分页查询菜单列表，支持按名称模糊搜索和状态过滤
func (r *menuRepo) List(ctx context.Context, page, pageSize int, name, status string) ([]*iam.Menu, int64, error) {
	var list []po.MenuPO
	var count int64

	db := r.data.DB(ctx).Model(&po.MenuPO{})

	// 按名称模糊搜索
	if name != "" {
		db = db.Where("name LIKE ?", "%"+name+"%")
	}
	// 按状态过滤
	if status != "" {
		db = db.Where("status = ?", status)
	}

	// 分页查询
	if err := datainfra.Paginate(db, page, pageSize, "sort ASC, id ASC", &list, &count); err != nil {
		return nil, 0, err
	}

	// 转换为领域实体
	menus := make([]*iam.Menu, 0, len(list))
	for i := range list {
		menus = append(menus, list[i].ToEntity())
	}
	return menus, count, nil
}

// GetTree 获取菜单树形结构
// 先查询所有菜单，再在 Go 代码中根据 ParentID 构建树
func (r *menuRepo) GetTree(ctx context.Context) ([]*iam.Menu, error) {
	var list []po.MenuPO
	// 查询所有正常状态的菜单
	if err := r.data.DB(ctx).
		Where("status = 1").
		Order("sort ASC, id ASC").
		Find(&list).Error; err != nil {
		return nil, err
	}

	// 转换为领域实体
	menus := make([]*iam.Menu, 0, len(list))
	for i := range list {
		menus = append(menus, list[i].ToEntity())
	}
	return buildMenuTree(menus), nil
}

// GetByRoleIDs 根据角色ID列表获取关联的菜单
// 通过 sys_role_menus 关联表查询
func (r *menuRepo) GetByRoleIDs(ctx context.Context, roleIDs []int64) ([]*iam.Menu, error) {
	if len(roleIDs) == 0 {
		return nil, nil
	}

	// 先查询角色关联的菜单ID（去重）
	var menuIDs []int64
	if err := r.data.DB(ctx).
		Table("sys_role_menus").
		Where("role_id IN ?", roleIDs).
		Pluck("DISTINCT menu_id", &menuIDs).Error; err != nil {
		return nil, err
	}

	if len(menuIDs) == 0 {
		return nil, nil
	}

	// 查询正常状态的菜单详情
	var list []po.MenuPO
	if err := r.data.DB(ctx).
		Where("id IN ? AND status = 1", menuIDs).
		Order("sort ASC, id ASC").
		Find(&list).Error; err != nil {
		return nil, err
	}

	menus := make([]*iam.Menu, 0, len(list))
	for i := range list {
		menus = append(menus, list[i].ToEntity())
	}
	return menus, nil
}

// buildMenuTree 根据父ID构建菜单树形结构
// parentId = 0 表示顶级菜单
func buildMenuTree(menus []*iam.Menu) []*iam.Menu {
	if len(menus) == 0 {
		return menus
	}

	// 建立 ID -> 菜单 的映射，方便快速查找
	menuMap := make(map[int64]*iam.Menu, len(menus))
	for _, m := range menus {
		menuMap[m.ID] = m
		// 初始化 Children 切片，避免前端收到 null
		m.Children = make([]*iam.Menu, 0)
	}

	// 构建树：根节点收集到 roots 中
	roots := make([]*iam.Menu, 0)
	for _, m := range menus {
		if m.ParentID == 0 {
			// 顶级菜单
			roots = append(roots, m)
		} else {
			// 子菜单：挂到父节点的 Children 上
			if parent, ok := menuMap[m.ParentID]; ok {
				parent.Children = append(parent.Children, m)
			}
		}
	}
	return roots
}
