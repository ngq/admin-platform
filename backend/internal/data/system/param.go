// 系统参数仓储实现
// 基于 GORM 实现 system.SysParamRepository 接口
package system

import (
	"context"
	"github.com/ngq/admin-platform/internal/datainfra"

	"github.com/ngq/admin-platform/internal/biz/system"
	"github.com/ngq/admin-platform/internal/data/po"
)

// paramRepo 系统参数仓储实现
type paramRepo struct {
	data *datainfra.Data
}

// NewParamRepo 创建系统参数仓储实例
func NewParamRepo(d *datainfra.Data) system.SysParamRepository {
	return &paramRepo{data: d}
}

// Create 创建系统参数
func (r *paramRepo) Create(ctx context.Context, p *system.SysParam) error {
	paramPO := po.SysParamPOFromEntity(p)
	if err := r.data.DB(ctx).Create(paramPO).Error; err != nil {
		return err
	}
	p.ID = paramPO.ID
	return nil
}

// Update 更新系统参数
func (r *paramRepo) Update(ctx context.Context, p *system.SysParam) error {
	paramPO := po.SysParamPOFromEntity(p)
	return r.data.DB(ctx).Model(&po.SysParamPO{}).
		Where("id = ?", p.ID).
		Updates(map[string]interface{}{
			"key":         paramPO.Key,
			"value":       paramPO.Value,
			"category":    paramPO.Category,
			"description": paramPO.Description,
			"status":      paramPO.Status,
		}).Error
}

// Delete 删除系统参数（软删除）
func (r *paramRepo) Delete(ctx context.Context, id int64) error {
	return r.data.DB(ctx).Delete(&po.SysParamPO{}, id).Error
}

// GetByID 根据ID获取系统参数
func (r *paramRepo) GetByID(ctx context.Context, id int64) (*system.SysParam, error) {
	var paramPO po.SysParamPO
	if err := r.data.DB(ctx).First(&paramPO, id).Error; err != nil {
		return nil, err
	}
	return paramPO.ToEntity(), nil
}

// GetByKey 根据参数Key获取系统参数
func (r *paramRepo) GetByKey(ctx context.Context, key string) (*system.SysParam, error) {
	var paramPO po.SysParamPO
	if err := r.data.DB(ctx).Where("`key` = ?", key).First(&paramPO).Error; err != nil {
		return nil, err
	}
	return paramPO.ToEntity(), nil
}

// List 分页查询系统参数，支持按分类和Key过滤
func (r *paramRepo) List(ctx context.Context, page, pageSize int, category, key string) ([]*system.SysParam, int64, error) {
	var list []po.SysParamPO
	var count int64

	db := r.data.DB(ctx).Model(&po.SysParamPO{})

	// 按分类过滤
	if category != "" {
		db = db.Where("category = ?", category)
	}
	// 按 Key 模糊搜索
	if key != "" {
		db = db.Where("`key` LIKE ?", "%"+key+"%")
	}

	// 分页查询
	if err := datainfra.Paginate(db, page, pageSize, "id DESC", &list, &count); err != nil {
		return nil, 0, err
	}

	params := make([]*system.SysParam, 0, len(list))
	for i := range list {
		params = append(params, list[i].ToEntity())
	}
	return params, count, nil
}
