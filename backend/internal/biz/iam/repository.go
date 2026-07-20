package iam

import "context"

// UserRepository 用户仓储接口
type UserRepository interface {
	// Create 创建用户
	Create(ctx context.Context, user *User) error
	// Update 更新用户信息
	Update(ctx context.Context, user *User) error
	// Delete 删除用户
	Delete(ctx context.Context, id int64) error
	// GetByID 根据ID获取用户
	GetByID(ctx context.Context, id int64) (*User, error)
	// GetByUsername 根据用户名获取用户
	GetByUsername(ctx context.Context, username string) (*User, error)
	// List 分页查询用户列表
	List(ctx context.Context, page, pageSize int, username, status string) ([]*User, int64, error)
	// Count 查询用户总数（用于仪表盘等场景，避免 List 的额外开销）
	Count(ctx context.Context) (int64, error)
	// UpdateRoles 更新用户角色关联
	UpdateRoles(ctx context.Context, userID int64, roleIDs []int64) error
	// SetPassword 设置用户密码（哈希值）
	SetPassword(ctx context.Context, id int64, hashedPassword string) error
	// UpdateLastLogin 更新最后登录信息
	UpdateLastLogin(ctx context.Context, id int64, ip string) error
}

// DataScopedUserRepository 是支持数据范围过滤的用户仓储扩展。
type DataScopedUserRepository interface {
	ListByDataScope(ctx context.Context, page, pageSize int, username, status string, deptID int64, scope DataPermissionScope) ([]*User, int64, error)
}

// DepartmentRepository 部门仓储接口
type DepartmentRepository interface {
	Create(ctx context.Context, dept *Department) error
	Update(ctx context.Context, dept *Department) error
	Delete(ctx context.Context, id int64) error
	GetByID(ctx context.Context, id int64) (*Department, error)
	List(ctx context.Context, name, status string) ([]*Department, error)
	GetDescendantIDs(ctx context.Context, id int64) ([]int64, error)
	GetManagedDeptIDs(ctx context.Context, leaderUserID int64) ([]int64, error)
}

// RoleRepository 角色仓储接口
type RoleRepository interface {
	// Create 创建角色
	Create(ctx context.Context, role *Role) error
	// Update 更新角色信息
	Update(ctx context.Context, role *Role) error
	// Delete 删除角色
	Delete(ctx context.Context, id int64) error
	// GetByID 根据ID获取角色
	GetByID(ctx context.Context, id int64) (*Role, error)
	// GetByCode 根据编码获取角色
	GetByCode(ctx context.Context, code string) (*Role, error)
	// List 分页查询角色列表
	List(ctx context.Context, page, pageSize int, name, code, status string) ([]*Role, int64, error)
	// Count 查询角色总数（用于仪表盘等场景，避免 List 的额外开销）
	Count(ctx context.Context) (int64, error)
	// GetByUserID 根据用户ID获取其关联角色列表
	GetByUserID(ctx context.Context, userID int64) ([]*Role, error)
	// UpdateMenus 更新角色关联的菜单
	UpdateMenus(ctx context.Context, roleID int64, menuIDs []int64) error
	// UpdateAPIs 更新角色关联的API资源
	UpdateAPIs(ctx context.Context, roleID int64, apiIDs []int64) error
	// UpdateButtons 更新角色关联的按钮权限
	UpdateButtons(ctx context.Context, roleID int64, buttonIDs []int64) error
	// UpdatePermissions 更新角色关联的菜单、API 和按钮权限
	UpdatePermissions(ctx context.Context, roleID int64, menuIDs []int64, apiIDs []int64, buttonIDs []int64) error
	// GetMenuIDs 获取角色关联的菜单ID列表
	GetMenuIDs(ctx context.Context, roleID int64) ([]int64, error)
	// GetAPIIDs 获取角色关联的API资源ID列表
	GetAPIIDs(ctx context.Context, roleID int64) ([]int64, error)
	// GetButtonIDs 获取角色关联的按钮权限ID列表
	GetButtonIDs(ctx context.Context, roleID int64) ([]int64, error)
}

// MenuRepository 菜单仓储接口
type MenuRepository interface {
	// Create 创建菜单
	Create(ctx context.Context, menu *Menu) error
	// Update 更新菜单信息
	Update(ctx context.Context, menu *Menu) error
	// Delete 删除菜单
	Delete(ctx context.Context, id int64) error
	// GetByID 根据ID获取菜单
	GetByID(ctx context.Context, id int64) (*Menu, error)
	// List 分页查询菜单列表
	List(ctx context.Context, page, pageSize int, name, status string) ([]*Menu, int64, error)
	// GetTree 获取菜单树形结构
	GetTree(ctx context.Context) ([]*Menu, error)
	// GetByRoleIDs 根据角色ID列表获取关联菜单
	GetByRoleIDs(ctx context.Context, roleIDs []int64) ([]*Menu, error)
}

// APIRepository API资源仓储接口
type APIRepository interface {
	// Create 创建API资源
	Create(ctx context.Context, api *APIResource) error
	// Update 更新API资源信息
	Update(ctx context.Context, api *APIResource) error
	// Delete 删除API资源
	Delete(ctx context.Context, id int64) error
	// GetByID 根据ID获取API资源
	GetByID(ctx context.Context, id int64) (*APIResource, error)
	// GetByIDs 根据ID列表获取API资源（用于权限分配时获取 API 详情）
	GetByIDs(ctx context.Context, ids []int64) ([]*APIResource, error)
	// List 分页查询API资源列表
	List(ctx context.Context, page, pageSize int, group, path, method string) ([]*APIResource, int64, error)
	// GetAll 获取全部API资源
	GetAll(ctx context.Context) ([]*APIResource, error)
	// GetByRoleIDs 根据角色ID列表获取关联API资源
	GetByRoleIDs(ctx context.Context, roleIDs []int64) ([]*APIResource, error)
	// BatchDelete 批量删除API资源
	BatchDelete(ctx context.Context, ids []int64) error
}

// ButtonRepository 按钮权限仓储接口
type ButtonRepository interface {
	// Create 创建按钮权限
	Create(ctx context.Context, button *Button) error
	// Update 更新按钮权限信息
	Update(ctx context.Context, button *Button) error
	// Delete 删除按钮权限
	Delete(ctx context.Context, id int64) error
	// GetByMenuID 根据菜单ID获取按钮权限列表
	GetByMenuID(ctx context.Context, menuID int64) ([]*Button, error)
	// GetByRoleIDs 根据角色ID列表获取关联按钮权限
	GetByRoleIDs(ctx context.Context, roleIDs []int64) ([]*Button, error)
}
