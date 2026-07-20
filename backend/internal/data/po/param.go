// Domain: System
// 系统参数持久化对象
// SysParamPO 对应 sys_params 表，存储运行时可修改的系统配置参数
package po

import (
	"time"

	"github.com/ngq/admin-platform/internal/biz/system"
	"gorm.io/gorm"
)

// SysParamPO 系统参数表持久化对象
type SysParamPO struct {
	ID          int64          `gorm:"primaryKey;autoIncrement" db:"id"`
	Key         string         `gorm:"type:varchar(128);uniqueIndex;not null" db:"key"`
	Value       string         `gorm:"type:text" db:"value"`
	Category    string         `gorm:"type:varchar(64);index" db:"category"`
	Description string         `gorm:"type:varchar(256)" db:"description"`
	Status      int8           `gorm:"type:tinyint;default:1;comment:1-正常 2-禁用" db:"status"`
	CreatedAt   time.Time      `gorm:"autoCreateTime" db:"created_at"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime" db:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" db:"deleted_at"`
}

// TableName 指定表名
func (SysParamPO) TableName() string { return "sys_params" }

// ToEntity 将 PO 转换为领域实体
func (p *SysParamPO) ToEntity() *system.SysParam {
	return &system.SysParam{
		ID:          p.ID,
		Key:         p.Key,
		Value:       p.Value,
		Category:    p.Category,
		Description: p.Description,
		Status:      p.Status,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}

// SysParamPOFromEntity 将领域实体转换为 PO
func SysParamPOFromEntity(p *system.SysParam) *SysParamPO {
	return &SysParamPO{
		ID:          p.ID,
		Key:         p.Key,
		Value:       p.Value,
		Category:    p.Category,
		Description: p.Description,
		Status:      p.Status,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}
