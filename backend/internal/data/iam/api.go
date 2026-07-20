// API 资源仓储实现
// 基于 GORM 实现 iam.APIRepository 接口
package iam

import (
	"context"
	"github.com/ngq/admin-platform/internal/datainfra"

	"github.com/ngq/admin-platform/internal/biz/iam"
	"github.com/ngq/admin-platform/internal/data/po"
)

// apiRepo API 资源仓储实现
type apiRepo struct {
	data *datainfra.Data
}

// NewAPIRepo 创建 API 资源仓储实例
func NewAPIRepo(d *datainfra.Data) iam.APIRepository {
	return &apiRepo{data: d}
}

// Create 创建 API 资源
func (r *apiRepo) Create(ctx context.Context, api *iam.APIResource) error {
	apiPO := po.APIPOFromEntity(api)
	return r.data.DB(ctx).Create(apiPO).Error
}

// Update 更新 API 资源
func (r *apiRepo) Update(ctx context.Context, api *iam.APIResource) error {
	apiPO := po.APIPOFromEntity(api)
	return r.data.DB(ctx).Model(&po.APIPO{}).
		Where("id = ?", api.ID).
		Updates(map[string]interface{}{
			"group_name":      apiPO.Group,
			"path":            apiPO.Path,
			"method":          apiPO.Method,
			"description":     apiPO.Description,
			"menu_id":         apiPO.MenuID,
			"action_type":     apiPO.ActionType,
			"permission_mode": apiPO.PermissionMode,
			"ignored":         apiPO.Ignored,
		}).Error
}

// Delete 删除 API 资源
func (r *apiRepo) Delete(ctx context.Context, id int64) error {
	return r.data.DB(ctx).Delete(&po.APIPO{}, id).Error
}

// GetByID 根据 ID 获取 API 资源
func (r *apiRepo) GetByID(ctx context.Context, id int64) (*iam.APIResource, error) {
	var apiPO po.APIPO
	if err := r.data.DB(ctx).First(&apiPO, id).Error; err != nil {
		return nil, err
	}
	return apiPO.ToEntity(), nil
}

// GetByIDs 根据 ID 列表获取 API 资源（用于权限分配时获取 API 详情）
func (r *apiRepo) GetByIDs(ctx context.Context, ids []int64) ([]*iam.APIResource, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var list []po.APIPO
	if err := r.data.DB(ctx).Where("id IN ?", ids).Find(&list).Error; err != nil {
		return nil, err
	}
	apis := make([]*iam.APIResource, 0, len(list))
	for i := range list {
		apis = append(apis, list[i].ToEntity())
	}
	return apis, nil
}

// List 分页查询 API 资源列表，支持按分组、路径和方法过滤
func (r *apiRepo) List(ctx context.Context, page, pageSize int, group, path, method string) ([]*iam.APIResource, int64, error) {
	var list []po.APIPO
	var count int64

	db := r.data.DB(ctx).Model(&po.APIPO{})

	// 按分组过滤
	if group != "" {
		db = db.Where("group_name = ?", group)
	}
	// 按路径模糊搜索
	if path != "" {
		db = db.Where("path LIKE ?", "%"+path+"%")
	}
	// 按请求方法过滤
	if method != "" {
		db = db.Where("method = ?", method)
	}

	// 分页查询
	if err := datainfra.Paginate(db, page, pageSize, "id DESC", &list, &count); err != nil {
		return nil, 0, err
	}

	// 转换为领域实体
	apis := make([]*iam.APIResource, 0, len(list))
	for i := range list {
		apis = append(apis, list[i].ToEntity())
	}
	return apis, count, nil
}

// GetAll 获取所有 API 资源（用于 Casbin 策略同步等场景）
func (r *apiRepo) GetAll(ctx context.Context) ([]*iam.APIResource, error) {
	var list []po.APIPO
	// 只查询非忽略权限校验的 API
	if err := r.data.DB(ctx).
		Where("ignored = 0").
		Where("permission_mode = ? OR permission_mode = '' OR permission_mode IS NULL", "casbin").
		Find(&list).Error; err != nil {
		return nil, err
	}

	apis := make([]*iam.APIResource, 0, len(list))
	for i := range list {
		apis = append(apis, list[i].ToEntity())
	}
	return apis, nil
}

// GetByRoleIDs 根据角色 ID 列表获取关联的 API 资源
func (r *apiRepo) GetByRoleIDs(ctx context.Context, roleIDs []int64) ([]*iam.APIResource, error) {
	if len(roleIDs) == 0 {
		return nil, nil
	}

	// 先查询角色关联的 API ID（去重）
	var apiIDs []int64
	if err := r.data.DB(ctx).
		Table("sys_role_apis").
		Where("role_id IN ?", roleIDs).
		Pluck("DISTINCT api_id", &apiIDs).Error; err != nil {
		return nil, err
	}

	if len(apiIDs) == 0 {
		return nil, nil
	}

	// 查询 API 详情
	var list []po.APIPO
	if err := r.data.DB(ctx).
		Where("id IN ?", apiIDs).
		Find(&list).Error; err != nil {
		return nil, err
	}

	apis := make([]*iam.APIResource, 0, len(list))
	for i := range list {
		apis = append(apis, list[i].ToEntity())
	}
	return apis, nil
}

// BatchDelete 批量删除 API 资源
func (r *apiRepo) BatchDelete(ctx context.Context, ids []int64) error {
	return r.data.DB(ctx).Where("id IN ?", ids).Delete(&po.APIPO{}).Error
}
