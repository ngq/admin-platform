// Domain: IAM
// 角色持久化对象
// RolePO 对应 sys_roles 表，存储角色信息与数据权限范围
package po

import (
	"time"

	"github.com/ngq/admin-platform/internal/biz/iam"
	"gorm.io/gorm"
)

// RolePO 角色表持久化对象
type RolePO struct {
	ID          int64          `gorm:"primaryKey;autoIncrement" db:"id"`
	Name        string         `gorm:"type:varchar(64);not null" db:"name"`
	Code        string         `gorm:"type:varchar(64);uniqueIndex;not null" db:"code"`
	ParentID    int64          `gorm:"default:0" db:"parent_id"`
	Sort        int            `gorm:"default:0" db:"sort"`
	DataScope   int8           `gorm:"type:tinyint;default:1;comment:1-全部 2-本部门 3-本部门及下级 4-仅本人" db:"data_scope"`
	Status      int8           `gorm:"type:tinyint;default:1;comment:1-正常 2-禁用" db:"status"`
	Description string         `gorm:"type:varchar(256)" db:"description"`
	CreatedAt   time.Time      `gorm:"autoCreateTime" db:"created_at"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime" db:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" db:"deleted_at"`
}

// TableName 指定表名
func (RolePO) TableName() string { return "sys_roles" }

// ToEntity 将 PO 转换为领域实体
func (r *RolePO) ToEntity() *iam.Role {
	return &iam.Role{
		ID:          r.ID,
		Name:        r.Name,
		Code:        r.Code,
		ParentID:    r.ParentID,
		Sort:        r.Sort,
		DataScope:   r.DataScope,
		Status:      r.Status,
		Description: r.Description,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}
}

// RolePOFromEntity 将领域实体转换为 PO
func RolePOFromEntity(r *iam.Role) *RolePO {
	return &RolePO{
		ID:          r.ID,
		Name:        r.Name,
		Code:        r.Code,
		ParentID:    r.ParentID,
		Sort:        r.Sort,
		DataScope:   r.DataScope,
		Status:      r.Status,
		Description: r.Description,
	}
}

// RoleMenuPO 角色-菜单关联表持久化对象
type RoleMenuPO struct {
	RoleID int64 `gorm:"primaryKey" db:"role_id"`
	MenuID int64 `gorm:"primaryKey" db:"menu_id"`
}

// TableName 指定表名
func (RoleMenuPO) TableName() string { return "sys_role_menus" }

// RoleAPIPO 角色-API关联表持久化对象
type RoleAPIPO struct {
	RoleID int64 `gorm:"primaryKey" db:"role_id"`
	ApiID  int64 `gorm:"primaryKey" db:"api_id"`
}

// TableName 指定表名
func (RoleAPIPO) TableName() string { return "sys_role_apis" }

// RoleButtonPO 角色-按钮关联表持久化对象
type RoleButtonPO struct {
	RoleID   int64 `gorm:"primaryKey" db:"role_id"`
	ButtonID int64 `gorm:"primaryKey" db:"button_id"`
}

// TableName 指定表名
func (RoleButtonPO) TableName() string { return "sys_role_buttons" }
