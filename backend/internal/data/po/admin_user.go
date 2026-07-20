// Domain: IAM
// 管理员用户持久化对象
// AdminUserPO 对应 sys_users 表，存储管理员基本信息与认证数据
package po

import (
	"time"

	"github.com/ngq/admin-platform/internal/biz/iam"
	"gorm.io/gorm"
)

// AdminUserPO 管理员用户表持久化对象
type AdminUserPO struct {
	ID          int64          `gorm:"primaryKey;autoIncrement" db:"id"`
	Username    string         `gorm:"type:varchar(64);uniqueIndex;not null" db:"username"`
	Password    string         `gorm:"type:varchar(128);not null" db:"password"`
	Nickname    string         `gorm:"type:varchar(64)" db:"nickname"`
	DeptID      int64          `gorm:"default:0;index" db:"dept_id"`
	DeptName    string         `gorm:"-" db:"-"`
	Email       string         `gorm:"type:varchar(128)" db:"email"`
	Phone       string         `gorm:"type:varchar(20)" db:"phone"`
	Avatar      string         `gorm:"type:varchar(256)" db:"avatar"`
	Status      int8           `gorm:"type:tinyint;default:1;comment:1-正常 2-禁用 3-锁定" db:"status"` // 1-正常 2-禁用 3-锁定
	LastLoginAt *time.Time     `gorm:"type:datetime" db:"last_login_at"`
	LastLoginIP string         `gorm:"type:varchar(45)" db:"last_login_ip"`
	CreatedAt   time.Time      `gorm:"autoCreateTime" db:"created_at"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime" db:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" db:"deleted_at"`
	Roles       []RolePO       `gorm:"many2many:sys_user_roles;foreignKey:ID;joinForeignKey:user_id;References:ID;joinReferences:role_id" db:"-"`
}

// TableName 指定表名
func (AdminUserPO) TableName() string { return "sys_users" }

// ToEntity 将 PO 转换为领域实体
func (u *AdminUserPO) ToEntity() *iam.User {
	user := &iam.User{
		ID:          u.ID,
		Username:    u.Username,
		Password:    u.Password,
		Nickname:    u.Nickname,
		DeptID:      u.DeptID,
		DeptName:    u.DeptName,
		Email:       u.Email,
		Phone:       u.Phone,
		Avatar:      u.Avatar,
		Status:      u.Status,
		LastLoginAt: u.LastLoginAt,
		LastLoginIP: u.LastLoginIP,
		CreatedAt:   u.CreatedAt,
		UpdatedAt:   u.UpdatedAt,
	}
	for _, r := range u.Roles {
		user.Roles = append(user.Roles, r.ToEntity())
	}
	return user
}

// AdminUserPOFromEntity 将领域实体转换为 PO
func AdminUserPOFromEntity(u *iam.User) *AdminUserPO {
	po := &AdminUserPO{
		ID:          u.ID,
		Username:    u.Username,
		Password:    u.Password,
		Nickname:    u.Nickname,
		DeptID:      u.DeptID,
		Email:       u.Email,
		Phone:       u.Phone,
		Avatar:      u.Avatar,
		Status:      u.Status,
		LastLoginAt: u.LastLoginAt,
		LastLoginIP: u.LastLoginIP,
	}
	for _, r := range u.Roles {
		po.Roles = append(po.Roles, *RolePOFromEntity(r))
	}
	return po
}

// UserRolePO 用户-角色关联表持久化对象
type UserRolePO struct {
	UserID int64 `gorm:"primaryKey" db:"user_id"`
	RoleID int64 `gorm:"primaryKey" db:"role_id"`
}

// TableName 指定表名
func (UserRolePO) TableName() string { return "sys_user_roles" }
