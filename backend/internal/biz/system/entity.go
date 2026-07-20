package system

import "time"

// SysParam 系统参数领域实体
type SysParam struct {
	ID          int64     // 参数唯一标识
	Key         string    // 参数键（唯一标识符）
	Value       string    // 参数值
	Category    string    // 参数分类
	Description string    // 参数描述
	Status      int8      // 状态：1-正常 2-禁用
	CreatedAt   time.Time // 创建时间
	UpdatedAt   time.Time // 更新时间
}

// SysParamStatus 系统参数状态常量
const (
	SysParamStatusNormal   int8 = 1 // 正常
	SysParamStatusDisabled int8 = 2 // 禁用
)
