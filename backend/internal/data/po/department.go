package po

import (
	"time"

	"github.com/ngq/admin-platform/internal/biz/iam"
	"gorm.io/gorm"
)

// DepartmentPO 对应 sys_departments 表，存储后台组织架构。
type DepartmentPO struct {
	ID           int64          `gorm:"primaryKey;autoIncrement" db:"id"`
	ParentID     int64          `gorm:"default:0;index" db:"parent_id"`
	Name         string         `gorm:"type:varchar(64);not null" db:"name"`
	LeaderUserID int64          `gorm:"default:0;index" db:"leader_user_id"`
	Sort         int            `gorm:"default:0" db:"sort"`
	Status       int8           `gorm:"type:tinyint;default:1;comment:1-正常 2-禁用" db:"status"`
	CreatedAt    time.Time      `gorm:"autoCreateTime" db:"created_at"`
	UpdatedAt    time.Time      `gorm:"autoUpdateTime" db:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" db:"deleted_at"`
}

func (DepartmentPO) TableName() string { return "sys_departments" }

func (d *DepartmentPO) ToEntity() *iam.Department {
	return &iam.Department{
		ID:           d.ID,
		ParentID:     d.ParentID,
		Name:         d.Name,
		LeaderUserID: d.LeaderUserID,
		Sort:         d.Sort,
		Status:       d.Status,
		CreatedAt:    d.CreatedAt,
		UpdatedAt:    d.UpdatedAt,
	}
}

func DepartmentPOFromEntity(d *iam.Department) *DepartmentPO {
	return &DepartmentPO{
		ID:           d.ID,
		ParentID:     d.ParentID,
		Name:         d.Name,
		LeaderUserID: d.LeaderUserID,
		Sort:         d.Sort,
		Status:       d.Status,
	}
}
