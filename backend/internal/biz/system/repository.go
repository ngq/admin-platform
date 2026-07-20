package system

import "context"

// SysParamRepository 系统参数仓储接口
type SysParamRepository interface {
	// Create 创建系统参数
	Create(ctx context.Context, p *SysParam) error
	// Update 更新系统参数
	Update(ctx context.Context, p *SysParam) error
	// Delete 删除系统参数
	Delete(ctx context.Context, id int64) error
	// GetByID 根据ID获取系统参数
	GetByID(ctx context.Context, id int64) (*SysParam, error)
	// GetByKey 根据键名获取系统参数
	GetByKey(ctx context.Context, key string) (*SysParam, error)
	// List 分页查询系统参数列表
	List(ctx context.Context, page, pageSize int, category, key string) ([]*SysParam, int64, error)
}
