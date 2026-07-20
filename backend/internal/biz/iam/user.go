package iam

import "time"

// User 用户领域实体
type User struct {
	ID          int64      `json:"id"`                      // 用户唯一标识
	Username    string     `json:"username"`                // 用户名
	Password    string     `json:"-"`                       // 密码（哈希值），不序列化到 JSON
	Nickname    string     `json:"nickname"`                // 昵称
	DeptID      int64      `json:"dept_id"`                 // 所属部门 ID
	DeptName    string     `json:"dept_name,omitempty"`     // 所属部门名称
	Email       string     `json:"email"`                   // 邮箱
	Phone       string     `json:"phone"`                   // 手机号
	Avatar      string     `json:"avatar"`                  // 头像URL
	Status      int8       `json:"status"`                  // 状态：1-正常 2-禁用 3-锁定
	LastLoginAt *time.Time `json:"last_login_at,omitempty"` // 最后登录时间
	LastLoginIP string     `json:"last_login_ip"`           // 最后登录IP
	Roles       []*Role    `json:"roles"`                   // 关联角色列表
	CreatedAt   time.Time  `json:"created_at"`              // 创建时间
	UpdatedAt   time.Time  `json:"updated_at"`              // 更新时间
}

// UserStatus 用户状态常量
const (
	UserStatusNormal   int8 = 1 // 正常
	UserStatusDisabled int8 = 2 // 禁用
	UserStatusLocked   int8 = 3 // 锁定
)

// IsNormal 判断用户是否处于正常状态
func (u *User) IsNormal() bool {
	return u.Status == UserStatusNormal
}
