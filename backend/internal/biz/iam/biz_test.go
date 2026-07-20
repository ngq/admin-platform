package iam

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ngq/admin-platform/internal/biz/system"
)

// ── Mock Repositories ──────────────────────────────────────────────────────

// MockUserRepository 用户仓储 Mock
type MockUserRepository struct {
	users     map[int64]*User
	usernames map[string]int64
	nextID    int64
	roleRepo  *MockRoleRepository
}

func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		users:     make(map[int64]*User),
		usernames: make(map[string]int64),
		nextID:    1,
	}
}

func (m *MockUserRepository) SetRoleRepo(roleRepo *MockRoleRepository) {
	m.roleRepo = roleRepo
}

func (m *MockUserRepository) Create(ctx context.Context, user *User) error {
	user.ID = m.nextID
	m.nextID++
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	m.users[user.ID] = user
	m.usernames[user.Username] = user.ID
	return nil
}

func (m *MockUserRepository) Update(ctx context.Context, user *User) error {
	if _, ok := m.users[user.ID]; !ok {
		return errors.New("user not found")
	}
	user.UpdatedAt = time.Now()
	m.users[user.ID] = user
	return nil
}

func (m *MockUserRepository) Delete(ctx context.Context, id int64) error {
	if user, ok := m.users[id]; ok {
		delete(m.usernames, user.Username)
		delete(m.users, id)
	}
	return nil
}

func (m *MockUserRepository) GetByID(ctx context.Context, id int64) (*User, error) {
	user, ok := m.users[id]
	if !ok {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func (m *MockUserRepository) GetByUsername(ctx context.Context, username string) (*User, error) {
	id, ok := m.usernames[username]
	if !ok {
		return nil, errors.New("user not found")
	}
	return m.users[id], nil
}

func (m *MockUserRepository) List(ctx context.Context, page, pageSize int, username, status string) ([]*User, int64, error) {
	var result []*User
	for _, u := range m.users {
		result = append(result, u)
	}
	return result, int64(len(result)), nil
}

func (m *MockUserRepository) Count(ctx context.Context) (int64, error) {
	return int64(len(m.users)), nil
}

func (m *MockUserRepository) UpdateRoles(ctx context.Context, userID int64, roleIDs []int64) error {
	if _, ok := m.users[userID]; !ok {
		return errors.New("user not found")
	}
	// 同步更新 MockRoleRepository 中的 userRoles 映射
	if m.roleRepo != nil {
		m.roleRepo.userRoles[userID] = roleIDs
	}
	return nil
}

func (m *MockUserRepository) SetPassword(ctx context.Context, id int64, hashedPassword string) error {
	if user, ok := m.users[id]; ok {
		user.Password = hashedPassword
		return nil
	}
	return errors.New("user not found")
}

func (m *MockUserRepository) UpdateLastLogin(ctx context.Context, id int64, ip string) error {
	if user, ok := m.users[id]; ok {
		now := time.Now()
		user.LastLoginAt = &now
		user.LastLoginIP = ip
		return nil
	}
	return errors.New("user not found")
}

// MockRoleRepository 角色仓储 Mock
type MockRoleRepository struct {
	roles       map[int64]*Role
	codes       map[string]int64
	nextID      int64
	userRoles   map[int64][]int64
	roleMenus   map[int64][]int64
	roleAPIs    map[int64][]int64
	roleButtons map[int64][]int64
}

func NewMockRoleRepository() *MockRoleRepository {
	return &MockRoleRepository{
		roles:       make(map[int64]*Role),
		codes:       make(map[string]int64),
		nextID:      1,
		userRoles:   make(map[int64][]int64),
		roleMenus:   make(map[int64][]int64),
		roleAPIs:    make(map[int64][]int64),
		roleButtons: make(map[int64][]int64),
	}
}

func (m *MockRoleRepository) Create(ctx context.Context, role *Role) error {
	role.ID = m.nextID
	m.nextID++
	role.CreatedAt = time.Now()
	role.UpdatedAt = time.Now()
	m.roles[role.ID] = role
	m.codes[role.Code] = role.ID
	return nil
}

func (m *MockRoleRepository) Update(ctx context.Context, role *Role) error {
	if _, ok := m.roles[role.ID]; !ok {
		return errors.New("role not found")
	}
	role.UpdatedAt = time.Now()
	m.roles[role.ID] = role
	return nil
}

func (m *MockRoleRepository) Delete(ctx context.Context, id int64) error {
	if role, ok := m.roles[id]; ok {
		delete(m.codes, role.Code)
		delete(m.roles, id)
	}
	return nil
}

func (m *MockRoleRepository) GetByID(ctx context.Context, id int64) (*Role, error) {
	role, ok := m.roles[id]
	if !ok {
		return nil, errors.New("role not found")
	}
	return role, nil
}

func (m *MockRoleRepository) GetByCode(ctx context.Context, code string) (*Role, error) {
	id, ok := m.codes[code]
	if !ok {
		return nil, errors.New("role not found")
	}
	return m.roles[id], nil
}

func (m *MockRoleRepository) List(ctx context.Context, page, pageSize int, name, code, status string) ([]*Role, int64, error) {
	var result []*Role
	for _, r := range m.roles {
		result = append(result, r)
	}
	return result, int64(len(result)), nil
}

func (m *MockRoleRepository) Count(ctx context.Context) (int64, error) {
	return int64(len(m.roles)), nil
}

func (m *MockRoleRepository) GetByUserID(ctx context.Context, userID int64) ([]*Role, error) {
	roleIDs, ok := m.userRoles[userID]
	if !ok {
		return []*Role{}, nil
	}
	var roles []*Role
	for _, id := range roleIDs {
		if r, ok := m.roles[id]; ok {
			roles = append(roles, r)
		}
	}
	return roles, nil
}

func (m *MockRoleRepository) UpdateMenus(ctx context.Context, roleID int64, menuIDs []int64) error {
	m.roleMenus[roleID] = menuIDs
	return nil
}

func (m *MockRoleRepository) UpdateAPIs(ctx context.Context, roleID int64, apiIDs []int64) error {
	m.roleAPIs[roleID] = apiIDs
	return nil
}

func (m *MockRoleRepository) UpdateButtons(ctx context.Context, roleID int64, buttonIDs []int64) error {
	m.roleButtons[roleID] = buttonIDs
	return nil
}

func (m *MockRoleRepository) UpdatePermissions(ctx context.Context, roleID int64, menuIDs []int64, apiIDs []int64, buttonIDs []int64) error {
	m.roleMenus[roleID] = menuIDs
	m.roleAPIs[roleID] = apiIDs
	m.roleButtons[roleID] = buttonIDs
	return nil
}

func (m *MockRoleRepository) GetMenuIDs(ctx context.Context, roleID int64) ([]int64, error) {
	return m.roleMenus[roleID], nil
}

func (m *MockRoleRepository) GetAPIIDs(ctx context.Context, roleID int64) ([]int64, error) {
	return m.roleAPIs[roleID], nil
}

func (m *MockRoleRepository) GetButtonIDs(ctx context.Context, roleID int64) ([]int64, error) {
	return m.roleButtons[roleID], nil
}

// MockMenuRepository 菜单仓储 Mock
type MockMenuRepository struct {
	menus  map[int64]*Menu
	nextID int64
}

func NewMockMenuRepository() *MockMenuRepository {
	return &MockMenuRepository{
		menus:  make(map[int64]*Menu),
		nextID: 1,
	}
}

func (m *MockMenuRepository) Create(ctx context.Context, menu *Menu) error {
	menu.ID = m.nextID
	m.nextID++
	menu.CreatedAt = time.Now()
	menu.UpdatedAt = time.Now()
	m.menus[menu.ID] = menu
	return nil
}

func (m *MockMenuRepository) Update(ctx context.Context, menu *Menu) error {
	if _, ok := m.menus[menu.ID]; !ok {
		return errors.New("menu not found")
	}
	menu.UpdatedAt = time.Now()
	m.menus[menu.ID] = menu
	return nil
}

func (m *MockMenuRepository) Delete(ctx context.Context, id int64) error {
	delete(m.menus, id)
	return nil
}

func (m *MockMenuRepository) GetByID(ctx context.Context, id int64) (*Menu, error) {
	menu, ok := m.menus[id]
	if !ok {
		return nil, errors.New("menu not found")
	}
	return menu, nil
}

func (m *MockMenuRepository) List(ctx context.Context, page, pageSize int, name, status string) ([]*Menu, int64, error) {
	var result []*Menu
	for _, menu := range m.menus {
		result = append(result, menu)
	}
	return result, int64(len(result)), nil
}

func (m *MockMenuRepository) GetTree(ctx context.Context) ([]*Menu, error) {
	return buildUserMenuTree(m.getAllMenus()), nil
}

func (m *MockMenuRepository) GetByRoleIDs(ctx context.Context, roleIDs []int64) ([]*Menu, error) {
	return m.getAllMenus(), nil
}

func (m *MockMenuRepository) getAllMenus() []*Menu {
	var result []*Menu
	for _, menu := range m.menus {
		result = append(result, menu)
	}
	return result
}

// MockAPIRepository API资源仓储 Mock
type MockAPIRepository struct {
	apis   map[int64]*APIResource
	nextID int64
}

func NewMockAPIRepository() *MockAPIRepository {
	return &MockAPIRepository{
		apis:   make(map[int64]*APIResource),
		nextID: 1,
	}
}

func (m *MockAPIRepository) Create(ctx context.Context, api *APIResource) error {
	api.ID = m.nextID
	m.nextID++
	api.CreatedAt = time.Now()
	api.UpdatedAt = time.Now()
	m.apis[api.ID] = api
	return nil
}

func (m *MockAPIRepository) Update(ctx context.Context, api *APIResource) error {
	if _, ok := m.apis[api.ID]; !ok {
		return errors.New("api not found")
	}
	api.UpdatedAt = time.Now()
	m.apis[api.ID] = api
	return nil
}

func (m *MockAPIRepository) Delete(ctx context.Context, id int64) error {
	delete(m.apis, id)
	return nil
}

func (m *MockAPIRepository) GetByID(ctx context.Context, id int64) (*APIResource, error) {
	api, ok := m.apis[id]
	if !ok {
		return nil, errors.New("api not found")
	}
	return api, nil
}

func (m *MockAPIRepository) GetByIDs(ctx context.Context, ids []int64) ([]*APIResource, error) {
	var result []*APIResource
	for _, id := range ids {
		if api, ok := m.apis[id]; ok {
			result = append(result, api)
		}
	}
	return result, nil
}

func (m *MockAPIRepository) List(ctx context.Context, page, pageSize int, group, path, method string) ([]*APIResource, int64, error) {
	var result []*APIResource
	for _, api := range m.apis {
		result = append(result, api)
	}
	return result, int64(len(result)), nil
}

func (m *MockAPIRepository) GetAll(ctx context.Context) ([]*APIResource, error) {
	var result []*APIResource
	for _, api := range m.apis {
		result = append(result, api)
	}
	return result, nil
}

func (m *MockAPIRepository) GetByRoleIDs(ctx context.Context, roleIDs []int64) ([]*APIResource, error) {
	return m.GetAll(ctx)
}

func (m *MockAPIRepository) BatchDelete(ctx context.Context, ids []int64) error {
	for _, id := range ids {
		delete(m.apis, id)
	}
	return nil
}

// MockButtonRepository 按钮权限仓储 Mock
type MockButtonRepository struct {
	buttons map[int64]*Button
	nextID  int64
}

func NewMockButtonRepository() *MockButtonRepository {
	return &MockButtonRepository{
		buttons: make(map[int64]*Button),
		nextID:  1,
	}
}

func (m *MockButtonRepository) Create(ctx context.Context, button *Button) error {
	button.ID = m.nextID
	m.nextID++
	button.CreatedAt = time.Now()
	button.UpdatedAt = time.Now()
	m.buttons[button.ID] = button
	return nil
}

func (m *MockButtonRepository) Update(ctx context.Context, button *Button) error {
	if _, ok := m.buttons[button.ID]; !ok {
		return errors.New("button not found")
	}
	button.UpdatedAt = time.Now()
	m.buttons[button.ID] = button
	return nil
}

func (m *MockButtonRepository) Delete(ctx context.Context, id int64) error {
	delete(m.buttons, id)
	return nil
}

func (m *MockButtonRepository) GetByMenuID(ctx context.Context, menuID int64) ([]*Button, error) {
	var result []*Button
	for _, btn := range m.buttons {
		if btn.MenuID == menuID {
			result = append(result, btn)
		}
	}
	return result, nil
}

func (m *MockButtonRepository) GetByRoleIDs(ctx context.Context, roleIDs []int64) ([]*Button, error) {
	var result []*Button
	for _, btn := range m.buttons {
		result = append(result, btn)
	}
	return result, nil
}

// MockSysParamRepository 系统参数仓储 Mock
type MockSysParamRepository struct {
	params map[string]*system.SysParam
}

func NewMockSysParamRepository() *MockSysParamRepository {
	return &MockSysParamRepository{
		params: make(map[string]*system.SysParam),
	}
}

func (m *MockSysParamRepository) Create(ctx context.Context, p *system.SysParam) error {
	m.params[p.Key] = p
	return nil
}

func (m *MockSysParamRepository) Update(ctx context.Context, p *system.SysParam) error {
	m.params[p.Key] = p
	return nil
}

func (m *MockSysParamRepository) Delete(ctx context.Context, id int64) error {
	return nil
}

func (m *MockSysParamRepository) GetByID(ctx context.Context, id int64) (*system.SysParam, error) {
	return nil, errors.New("not found")
}

func (m *MockSysParamRepository) GetByKey(ctx context.Context, key string) (*system.SysParam, error) {
	p, ok := m.params[key]
	if !ok {
		return nil, errors.New("not found")
	}
	return p, nil
}

func (m *MockSysParamRepository) List(ctx context.Context, page, pageSize int, category, status string) ([]*system.SysParam, int64, error) {
	var result []*system.SysParam
	for _, p := range m.params {
		result = append(result, p)
	}
	return result, int64(len(result)), nil
}

// ── Test Helper ────────────────────────────────────────────────────────────

func newTestBiz() *Biz {
	userRepo := NewMockUserRepository()
	roleRepo := NewMockRoleRepository()
	menuRepo := NewMockMenuRepository()
	apiRepo := NewMockAPIRepository()
	buttonRepo := NewMockButtonRepository()
	sysParamRepo := NewMockSysParamRepository()

	// 设置角色仓储引用，用于 UpdateRoles 同步
	userRepo.SetRoleRepo(roleRepo)

	return NewBiz(
		userRepo,
		roleRepo,
		menuRepo,
		apiRepo,
		buttonRepo,
		sysParamRepo,
	)
}

// ── User Tests ──────────────────────────────────────────────────────────────

// TestUserCreate 测试用户创建
func TestUserCreate(t *testing.T) {
	biz := newTestBiz()
	ctx := context.Background()

	tests := []struct {
		name    string
		user    *User
		wantErr bool
	}{
		{
			name: "创建正常用户",
			user: &User{
				Username: "testuser",
				Password: "abc123",
				Nickname: "测试用户",
				Status:   UserStatusNormal,
			},
			wantErr: false,
		},
		{
			name: "用户名已存在",
			user: &User{
				Username: "testuser",
				Password: "xyz789",
				Nickname: "重复用户",
				Status:   UserStatusNormal,
			},
			wantErr: true,
		},
		{
			name: "密码过短",
			user: &User{
				Username: "shortpwd",
				Password: "ab1",
				Nickname: "短密码用户",
				Status:   UserStatusNormal,
			},
			wantErr: true,
		},
		{
			name: "密码无数字",
			user: &User{
				Username: "nonumb",
				Password: "abcdefgh",
				Nickname: "无数字密码",
				Status:   UserStatusNormal,
			},
			wantErr: true,
		},
		{
			name: "密码无字母",
			user: &User{
				Username: "noletter",
				Password: "12345678",
				Nickname: "无字母密码",
				Status:   UserStatusNormal,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := biz.UserCreate(ctx, tt.user)
			if (err != nil) != tt.wantErr {
				t.Errorf("UserCreate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestUserUpdate 测试用户更新
func TestUserUpdate(t *testing.T) {
	biz := newTestBiz()
	ctx := context.Background()

	// 先创建用户
	user := &User{
		Username: "updateuser",
		Password: "abc123",
		Nickname: "更新前",
		Status:   UserStatusNormal,
	}
	_ = biz.UserCreate(ctx, user)

	// 更新用户
	user.Nickname = "更新后"
	err := biz.UserUpdate(ctx, user)
	if err != nil {
		t.Fatalf("UserUpdate failed: %v", err)
	}

	// 验证更新结果
	got, err := biz.UserRepo.GetByID(ctx, user.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if got.Nickname != "更新后" {
		t.Errorf("expected nickname '更新后', got '%s'", got.Nickname)
	}
}

// TestUserDelete 测试用户删除
func TestUserDelete(t *testing.T) {
	biz := newTestBiz()
	ctx := context.Background()

	// 先创建用户
	user := &User{
		Username: "deleteuser",
		Password: "abc123",
		Nickname: "待删除用户",
		Status:   UserStatusNormal,
	}
	_ = biz.UserCreate(ctx, user)

	// 删除用户
	err := biz.UserDelete(ctx, user.ID)
	if err != nil {
		t.Fatalf("UserDelete failed: %v", err)
	}

	// 验证删除结果
	_, err = biz.UserRepo.GetByID(ctx, user.ID)
	if err == nil {
		t.Error("expected user to be deleted, but still exists")
	}
}

// TestUserList 测试用户列表查询
func TestUserList(t *testing.T) {
	biz := newTestBiz()
	ctx := context.Background()

	// 创建多个用户
	for i := 1; i <= 3; i++ {
		_ = biz.UserCreate(ctx, &User{
			Username: "listuser" + string(rune('0'+i)),
			Password: "abc123",
			Nickname: "列表用户" + string(rune('0'+i)),
			Status:   UserStatusNormal,
		})
	}

	// 查询列表
	users, total, err := biz.UserList(ctx, 1, 10, "", "")
	if err != nil {
		t.Fatalf("UserList failed: %v", err)
	}

	if total < 3 {
		t.Errorf("expected at least 3 users, got %d", total)
	}
	if len(users) < 3 {
		t.Errorf("expected at least 3 users in list, got %d", len(users))
	}
}

// TestUserGetByID 测试根据ID获取用户
func TestUserGetByID(t *testing.T) {
	biz := newTestBiz()
	ctx := context.Background()

	// 先创建用户
	user := &User{
		Username: "getuser",
		Password: "abc123",
		Nickname: "获取测试",
		Status:   UserStatusNormal,
	}
	_ = biz.UserCreate(ctx, user)

	// 正常获取
	got, err := biz.UserGetByID(ctx, user.ID)
	if err != nil {
		t.Fatalf("UserGetByID failed: %v", err)
	}
	if got.Username != "getuser" {
		t.Errorf("expected username 'getuser', got '%s'", got.Username)
	}

	// 获取不存在的用户
	_, err = biz.UserGetByID(ctx, 9999)
	if err == nil {
		t.Error("expected error for non-existent user, got nil")
	}
}

// TestUserChangePassword 测试修改密码
func TestUserChangePassword(t *testing.T) {
	biz := newTestBiz()
	ctx := context.Background()

	// 先创建用户
	user := &User{
		Username: "changepwd",
		Password: "abc123",
		Nickname: "改密码用户",
		Status:   UserStatusNormal,
	}
	_ = biz.UserCreate(ctx, user)

	tests := []struct {
		name        string
		oldPassword string
		newPassword string
		wantErr     bool
	}{
		{
			name:        "正常修改密码",
			oldPassword: "abc123",
			newPassword: "xyz789",
			wantErr:     false,
		},
		{
			name:        "旧密码错误",
			oldPassword: "wrongpwd",
			newPassword: "newpass1",
			wantErr:     true,
		},
		{
			name:        "新密码不符合规则",
			oldPassword: "xyz789",
			newPassword: "123",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := biz.UserChangePassword(ctx, user.ID, tt.oldPassword, tt.newPassword)
			if (err != nil) != tt.wantErr {
				t.Errorf("UserChangePassword() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestUserResetPassword 测试重置密码
func TestUserResetPassword(t *testing.T) {
	biz := newTestBiz()
	ctx := context.Background()

	// 先创建用户
	user := &User{
		Username: "resetpwd",
		Password: "abc123",
		Nickname: "重置密码用户",
		Status:   UserStatusNormal,
	}
	_ = biz.UserCreate(ctx, user)

	tests := []struct {
		name        string
		newPassword string
		wantErr     bool
	}{
		{
			name:        "正常重置密码",
			newPassword: "newpass1",
			wantErr:     false,
		},
		{
			name:        "密码过短",
			newPassword: "abc",
			wantErr:     true,
		},
		{
			name:        "密码无字母",
			newPassword: "12345678",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := biz.UserResetPassword(ctx, user.ID, tt.newPassword)
			if (err != nil) != tt.wantErr {
				t.Errorf("UserResetPassword() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestUserUpdateRoles 测试更新用户角色
func TestUserUpdateRoles(t *testing.T) {
	biz := newTestBiz()
	ctx := context.Background()

	// 创建用户
	user := &User{
		Username: "roleuser",
		Password: "abc123",
		Nickname: "角色用户",
		Status:   UserStatusNormal,
	}
	_ = biz.UserCreate(ctx, user)

	// 创建角色
	role1 := &Role{Name: "角色1", Code: "role1", Status: 1}
	role2 := &Role{Name: "角色2", Code: "role2", Status: 1}
	_ = biz.RoleRepo.Create(ctx, role1)
	_ = biz.RoleRepo.Create(ctx, role2)

	// 更新用户角色
	err := biz.UserUpdateRoles(ctx, user.ID, []int64{role1.ID, role2.ID})
	if err != nil {
		t.Fatalf("UserUpdateRoles failed: %v", err)
	}
}

// ── Role Tests ──────────────────────────────────────────────────────────────

// TestRoleCreate 测试角色创建
func TestRoleCreate(t *testing.T) {
	biz := newTestBiz()
	ctx := context.Background()

	role := &Role{
		Name:      "测试角色",
		Code:      "test_role",
		ParentID:  0,
		Sort:      1,
		DataScope: DataScopeAll,
		Status:    1,
	}

	err := biz.RoleRepo.Create(ctx, role)
	if err != nil {
		t.Fatalf("RoleCreate failed: %v", err)
	}
	if role.ID == 0 {
		t.Error("expected role ID to be set after creation")
	}
}

// TestRoleUpdate 测试角色更新
func TestRoleUpdate(t *testing.T) {
	biz := newTestBiz()
	ctx := context.Background()

	// 先创建角色
	role := &Role{
		Name:      "更新前角色",
		Code:      "before_update",
		ParentID:  0,
		Sort:      1,
		DataScope: DataScopeAll,
		Status:    1,
	}
	_ = biz.RoleRepo.Create(ctx, role)

	// 更新角色
	role.Name = "更新后角色"
	err := biz.RoleRepo.Update(ctx, role)
	if err != nil {
		t.Fatalf("RoleUpdate failed: %v", err)
	}

	// 验证
	got, _ := biz.RoleRepo.GetByID(ctx, role.ID)
	if got.Name != "更新后角色" {
		t.Errorf("expected name '更新后角色', got '%s'", got.Name)
	}
}

// TestRoleDelete 测试角色删除
func TestRoleDelete(t *testing.T) {
	biz := newTestBiz()
	ctx := context.Background()

	// 先创建角色
	role := &Role{
		Name:      "待删除角色",
		Code:      "to_delete",
		ParentID:  0,
		Sort:      1,
		DataScope: DataScopeAll,
		Status:    1,
	}
	_ = biz.RoleRepo.Create(ctx, role)

	// 删除角色
	err := biz.RoleRepo.Delete(ctx, role.ID)
	if err != nil {
		t.Fatalf("RoleDelete failed: %v", err)
	}

	// 验证
	_, err = biz.RoleRepo.GetByID(ctx, role.ID)
	if err == nil {
		t.Error("expected role to be deleted")
	}
}

// TestRoleAssignMenus 测试角色分配菜单
func TestRoleAssignMenus(t *testing.T) {
	biz := newTestBiz()
	ctx := context.Background()

	// 创建角色和菜单
	role := &Role{Name: "菜单角色", Code: "menu_role", Status: 1}
	_ = biz.RoleRepo.Create(ctx, role)

	menu1 := &Menu{Name: "菜单1", Path: "/menu1", Type: MenuTypeMenu}
	menu2 := &Menu{Name: "菜单2", Path: "/menu2", Type: MenuTypeMenu}
	_ = biz.MenuRepo.Create(ctx, menu1)
	_ = biz.MenuRepo.Create(ctx, menu2)

	// 分配菜单
	err := biz.RoleRepo.UpdateMenus(ctx, role.ID, []int64{menu1.ID, menu2.ID})
	if err != nil {
		t.Fatalf("UpdateMenus failed: %v", err)
	}

	// 验证
	menuIDs, _ := biz.RoleRepo.GetMenuIDs(ctx, role.ID)
	if len(menuIDs) != 2 {
		t.Errorf("expected 2 menu IDs, got %d", len(menuIDs))
	}
}

// TestRoleAssignAPIs 测试角色分配API资源
func TestRoleAssignAPIs(t *testing.T) {
	biz := newTestBiz()
	ctx := context.Background()

	// 创建角色和API
	role := &Role{Name: "API角色", Code: "api_role", Status: 1}
	_ = biz.RoleRepo.Create(ctx, role)

	api1 := &APIResource{Group: "user", Path: "/api/user", Method: "GET"}
	api2 := &APIResource{Group: "user", Path: "/api/user", Method: "POST"}
	_ = biz.APIRepo.Create(ctx, api1)
	_ = biz.APIRepo.Create(ctx, api2)

	// 分配API
	err := biz.RoleRepo.UpdateAPIs(ctx, role.ID, []int64{api1.ID, api2.ID})
	if err != nil {
		t.Fatalf("UpdateAPIs failed: %v", err)
	}

	// 验证
	apiIDs, _ := biz.RoleRepo.GetAPIIDs(ctx, role.ID)
	if len(apiIDs) != 2 {
		t.Errorf("expected 2 API IDs, got %d", len(apiIDs))
	}
}

// TestRoleAssignButtons 测试角色分配按钮权限
func TestRoleAssignButtons(t *testing.T) {
	biz := newTestBiz()
	ctx := context.Background()

	// 创建角色和按钮
	role := &Role{Name: "按钮角色", Code: "btn_role", Status: 1}
	_ = biz.RoleRepo.Create(ctx, role)

	btn1 := &Button{MenuID: 1, Name: "新增", Code: "add"}
	btn2 := &Button{MenuID: 1, Name: "删除", Code: "delete"}
	_ = biz.ButtonRepo.Create(ctx, btn1)
	_ = biz.ButtonRepo.Create(ctx, btn2)

	// 分配按钮
	err := biz.RoleRepo.UpdateButtons(ctx, role.ID, []int64{btn1.ID, btn2.ID})
	if err != nil {
		t.Fatalf("UpdateButtons failed: %v", err)
	}

	// 验证
	btnIDs, _ := biz.RoleRepo.GetButtonIDs(ctx, role.ID)
	if len(btnIDs) != 2 {
		t.Errorf("expected 2 button IDs, got %d", len(btnIDs))
	}
}

// ── Menu Tests ───────────────────────────────────────────────────────────────

// TestMenuCreate 测试菜单创建
func TestMenuCreate(t *testing.T) {
	biz := newTestBiz()
	ctx := context.Background()

	menu := &Menu{
		Name:      "系统管理",
		Path:      "/system",
		Component: "Layout",
		Type:      MenuTypeDirectory,
		Sort:      1,
		Status:    1,
	}

	err := biz.MenuCreate(ctx, menu)
	if err != nil {
		t.Fatalf("MenuCreate failed: %v", err)
	}
	if menu.ID == 0 {
		t.Error("expected menu ID to be set")
	}
}

// TestMenuUpdate 测试菜单更新
func TestMenuUpdate(t *testing.T) {
	biz := newTestBiz()
	ctx := context.Background()

	// 创建菜单
	menu := &Menu{
		Name:      "更新前菜单",
		Path:      "/before",
		Component: "Layout",
		Type:      MenuTypeMenu,
		Sort:      1,
		Status:    1,
	}
	_ = biz.MenuCreate(ctx, menu)

	// 更新
	menu.Name = "更新后菜单"
	err := biz.MenuUpdate(ctx, menu)
	if err != nil {
		t.Fatalf("MenuUpdate failed: %v", err)
	}

	// 验证
	got, _ := biz.MenuRepo.GetByID(ctx, menu.ID)
	if got.Name != "更新后菜单" {
		t.Errorf("expected '更新后菜单', got '%s'", got.Name)
	}
}

// TestMenuDelete 测试菜单删除
func TestMenuDelete(t *testing.T) {
	biz := newTestBiz()
	ctx := context.Background()

	// 创建菜单
	menu := &Menu{
		Name:   "待删除菜单",
		Path:   "/delete",
		Type:   MenuTypeMenu,
		Sort:   1,
		Status: 1,
	}
	_ = biz.MenuCreate(ctx, menu)

	// 删除
	err := biz.MenuDelete(ctx, menu.ID)
	if err != nil {
		t.Fatalf("MenuDelete failed: %v", err)
	}

	// 验证
	_, err = biz.MenuRepo.GetByID(ctx, menu.ID)
	if err == nil {
		t.Error("expected menu to be deleted")
	}
}

// TestMenuGetTree 测试获取菜单树
func TestMenuGetTree(t *testing.T) {
	biz := newTestBiz()
	ctx := context.Background()

	// 创建菜单树
	parent := &Menu{Name: "父菜单", Path: "/parent", Type: MenuTypeDirectory, Sort: 1, Status: 1}
	_ = biz.MenuCreate(ctx, parent)

	child1 := &Menu{Name: "子菜单1", Path: "/parent/child1", ParentID: parent.ID, Type: MenuTypeMenu, Sort: 1, Status: 1}
	_ = biz.MenuCreate(ctx, child1)

	child2 := &Menu{Name: "子菜单2", Path: "/parent/child2", ParentID: parent.ID, Type: MenuTypeMenu, Sort: 2, Status: 1}
	_ = biz.MenuCreate(ctx, child2)

	// 获取树
	tree, err := biz.MenuGetTree(ctx)
	if err != nil {
		t.Fatalf("MenuGetTree failed: %v", err)
	}

	if len(tree) == 0 {
		t.Error("expected at least one root menu")
	}
}

// TestBuildUserMenuTree 测试构建用户菜单树
func TestBuildUserMenuTree(t *testing.T) {
	menus := []*Menu{
		{ID: 1, ParentID: 0, Name: "系统管理", Path: "/system"},
		{ID: 2, ParentID: 1, Name: "用户管理", Path: "/system/user"},
		{ID: 3, ParentID: 1, Name: "角色管理", Path: "/system/role"},
		{ID: 4, ParentID: 0, Name: "内容管理", Path: "/content"},
		{ID: 5, ParentID: 4, Name: "文章管理", Path: "/content/article"},
	}

	tree := buildUserMenuTree(menus)

	// 验证根节点
	if len(tree) != 2 {
		t.Fatalf("expected 2 root nodes, got %d", len(tree))
	}

	// 验证子节点
	var sysMenu *Menu
	for _, m := range tree {
		if m.ID == 1 {
			sysMenu = m
			break
		}
	}
	if sysMenu == nil {
		t.Fatal("system menu not found")
	}
	if len(sysMenu.Children) != 2 {
		t.Errorf("expected 2 children for system menu, got %d", len(sysMenu.Children))
	}
}

// ── API Tests ───────────────────────────────────────────────────────────────

// TestAPICreate 测试API创建
func TestAPICreate(t *testing.T) {
	biz := newTestBiz()
	ctx := context.Background()

	api := &APIResource{
		Group:       "用户管理",
		Path:        "/api/user",
		Method:      "GET",
		Description: "获取用户列表",
	}

	err := biz.APICreate(ctx, api)
	if err != nil {
		t.Fatalf("APICreate failed: %v", err)
	}
	if api.ID == 0 {
		t.Error("expected API ID to be set")
	}
}

// TestAPIUpdate 测试API更新
func TestAPIUpdate(t *testing.T) {
	biz := newTestBiz()
	ctx := context.Background()

	// 创建
	api := &APIResource{
		Group:       "用户管理",
		Path:        "/api/user",
		Method:      "GET",
		Description: "获取用户列表",
	}
	_ = biz.APICreate(ctx, api)

	// 更新
	api.Description = "获取用户列表（更新）"
	err := biz.APIUpdate(ctx, api)
	if err != nil {
		t.Fatalf("APIUpdate failed: %v", err)
	}

	// 验证
	got, _ := biz.APIRepo.GetByID(ctx, api.ID)
	if got.Description != "获取用户列表（更新）" {
		t.Errorf("unexpected description: %s", got.Description)
	}
}

// TestAPIDelete 测试API删除
func TestAPIDelete(t *testing.T) {
	biz := newTestBiz()
	ctx := context.Background()

	// 创建
	api := &APIResource{
		Group:  "用户管理",
		Path:   "/api/user/delete",
		Method: "DELETE",
	}
	_ = biz.APICreate(ctx, api)

	// 删除
	err := biz.APIDelete(ctx, api.ID)
	if err != nil {
		t.Fatalf("APIDelete failed: %v", err)
	}

	// 验证
	_, err = biz.APIRepo.GetByID(ctx, api.ID)
	if err == nil {
		t.Error("expected API to be deleted")
	}
}

// TestAPIBatchDelete 测试API批量删除
func TestAPIBatchDelete(t *testing.T) {
	biz := newTestBiz()
	ctx := context.Background()

	// 创建多个API
	api1 := &APIResource{Group: "g1", Path: "/p1", Method: "GET"}
	api2 := &APIResource{Group: "g1", Path: "/p2", Method: "POST"}
	_ = biz.APICreate(ctx, api1)
	_ = biz.APICreate(ctx, api2)

	// 批量删除
	err := biz.APIBatchDelete(ctx, []int64{api1.ID, api2.ID})
	if err != nil {
		t.Fatalf("APIBatchDelete failed: %v", err)
	}

	// 验证
	all, _ := biz.APIRepo.GetAll(ctx)
	for _, a := range all {
		if a.ID == api1.ID || a.ID == api2.ID {
			t.Error("API should be deleted")
		}
	}
}

// ── Button Tests ─────────────────────────────────────────────────────────────

// TestButtonCreate 测试按钮创建
func TestButtonCreate(t *testing.T) {
	biz := newTestBiz()
	ctx := context.Background()

	btn := &Button{
		MenuID:      1,
		Name:        "新增用户",
		Code:        "user:add",
		Description: "新增用户按钮",
	}

	err := biz.ButtonCreate(ctx, btn)
	if err != nil {
		t.Fatalf("ButtonCreate failed: %v", err)
	}
	if btn.ID == 0 {
		t.Error("expected button ID to be set")
	}
}

// TestButtonUpdate 测试按钮更新
func TestButtonUpdate(t *testing.T) {
	biz := newTestBiz()
	ctx := context.Background()

	// 创建
	btn := &Button{
		MenuID:      1,
		Name:        "更新前",
		Code:        "btn:before",
		Description: "更新前描述",
	}
	_ = biz.ButtonCreate(ctx, btn)

	// 更新
	btn.Name = "更新后"
	btn.Code = "btn:after"
	err := biz.ButtonUpdate(ctx, btn)
	if err != nil {
		t.Fatalf("ButtonUpdate failed: %v", err)
	}

	// 验证
	btns, _ := biz.ButtonRepo.GetByMenuID(ctx, 1)
	for _, b := range btns {
		if b.ID == btn.ID && b.Name != "更新后" {
			t.Errorf("expected name '更新后', got '%s'", b.Name)
		}
	}
}

// TestButtonDelete 测试按钮删除
func TestButtonDelete(t *testing.T) {
	biz := newTestBiz()
	ctx := context.Background()

	// 创建
	btn := &Button{
		MenuID: 1,
		Name:   "待删除",
		Code:   "btn:delete",
	}
	_ = biz.ButtonCreate(ctx, btn)

	// 删除
	err := biz.ButtonDelete(ctx, btn.ID)
	if err != nil {
		t.Fatalf("ButtonDelete failed: %v", err)
	}

	// 验证
	btns, _ := biz.ButtonRepo.GetByMenuID(ctx, 1)
	for _, b := range btns {
		if b.ID == btn.ID {
			t.Error("button should be deleted")
		}
	}
}

// TestButtonGetByMenuID 测试根据菜单ID获取按钮
func TestButtonGetByMenuID(t *testing.T) {
	biz := newTestBiz()
	ctx := context.Background()

	// 创建多个按钮
	for i := 1; i <= 3; i++ {
		_ = biz.ButtonCreate(ctx, &Button{
			MenuID: 1,
			Name:   "按钮" + string(rune('0'+i)),
			Code:   "btn" + string(rune('0'+i)),
		})
	}

	// 查询
	btns, err := biz.ButtonGetByMenuID(ctx, 1)
	if err != nil {
		t.Fatalf("ButtonGetByMenuID failed: %v", err)
	}
	if len(btns) != 3 {
		t.Errorf("expected 3 buttons, got %d", len(btns))
	}
}

// ── Password Validation Tests ───────────────────────────────────────────────

// TestValidatePassword 测试密码验证
func TestValidatePassword(t *testing.T) {
	sysParamRepo := NewMockSysParamRepository()
	ctx := context.Background()

	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{"有效密码", "abc123", false},
		{"密码过短", "ab1", true},
		{"无字母", "123456", true},
		{"无数字", "abcdef", true},
		{"空密码", "", true},
		{"仅字母数字混合", "Test1234", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePassword(ctx, sysParamRepo, tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("validatePassword(%s) error = %v, wantErr %v", tt.password, err, tt.wantErr)
			}
		})
	}
}

// TestValidatePasswordWithCustomMinLength 测试自定义最小密码长度
func TestValidatePasswordWithCustomMinLength(t *testing.T) {
	sysParamRepo := NewMockSysParamRepository()
	ctx := context.Background()

	// 设置自定义最小长度
	_ = sysParamRepo.Create(ctx, &system.SysParam{
		Key:   "password_min_length",
		Value: "8",
	})

	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{"符合自定义长度", "abc12345", false},
		{"不符合自定义长度", "abc123", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePassword(ctx, sysParamRepo, tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("validatePassword(%s) error = %v, wantErr %v", tt.password, err, tt.wantErr)
			}
		})
	}
}

// TestHashPassword 测试密码哈希
func TestHashPassword(t *testing.T) {
	password := "test123"

	hashed, err := hashPassword(password)
	if err != nil {
		t.Fatalf("hashPassword failed: %v", err)
	}

	if hashed == password {
		t.Error("hashed password should not equal plain password")
	}
	if len(hashed) < 20 {
		t.Error("hashed password seems too short")
	}
}

// ── User Model Tests ─────────────────────────────────────────────────────────

// TestUserIsNormal 测试用户状态判断
func TestUserIsNormal(t *testing.T) {
	tests := []struct {
		name     string
		status   int8
		expected bool
	}{
		{"正常状态", UserStatusNormal, true},
		{"禁用状态", UserStatusDisabled, false},
		{"锁定状态", UserStatusLocked, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &User{Status: tt.status}
			if user.IsNormal() != tt.expected {
				t.Errorf("IsNormal() = %v, expected %v", user.IsNormal(), tt.expected)
			}
		})
	}
}

// ── Integration Tests ────────────────────────────────────────────────────────

// TestUserMenuIntegration 测试用户菜单关联
func TestUserMenuIntegration(t *testing.T) {
	biz := newTestBiz()
	ctx := context.Background()

	// 创建角色
	role := &Role{Name: "测试角色", Code: "test", Status: 1}
	_ = biz.RoleRepo.Create(ctx, role)

	// 创建菜单
	menu1 := &Menu{Name: "系统管理", Path: "/system", Type: MenuTypeDirectory, Sort: 1, Status: 1}
	menu2 := &Menu{Name: "用户管理", Path: "/system/user", Type: MenuTypeMenu, Sort: 1, Status: 1}
	_ = biz.MenuCreate(ctx, menu1)
	_ = biz.MenuCreate(ctx, menu2)

	// 分配菜单给角色
	_ = biz.RoleRepo.UpdateMenus(ctx, role.ID, []int64{menu1.ID, menu2.ID})

	// 创建用户并分配角色
	user := &User{Username: "integuser", Password: "abc123", Nickname: "集成测试", Status: UserStatusNormal}
	_ = biz.UserCreate(ctx, user)
	_ = biz.UserUpdateRoles(ctx, user.ID, []int64{role.ID})

	// 获取用户菜单
	menus, err := biz.MenuGetUserMenus(ctx, user.ID)
	if err != nil {
		t.Fatalf("MenuGetUserMenus failed: %v", err)
	}
	if len(menus) == 0 {
		t.Error("expected user to have menus")
	}
}

// TestMenuGetByID 测试根据ID获取菜单
func TestMenuGetByID(t *testing.T) {
	biz := newTestBiz()
	ctx := context.Background()

	// 创建菜单
	menu := &Menu{Name: "测试菜单", Path: "/test", Type: MenuTypeMenu, Sort: 1, Status: 1}
	_ = biz.MenuCreate(ctx, menu)

	// 获取
	got, err := biz.MenuRepo.GetByID(ctx, menu.ID)
	if err != nil {
		t.Fatalf("MenuGetByID failed: %v", err)
	}
	if got.Name != "测试菜单" {
		t.Errorf("expected '测试菜单', got '%s'", got.Name)
	}
}

// TestMenuList 测试菜单列表查询
func TestMenuList(t *testing.T) {
	biz := newTestBiz()
	ctx := context.Background()

	// 创建菜单
	for i := 1; i <= 3; i++ {
		_ = biz.MenuCreate(ctx, &Menu{
			Name:   "菜单" + string(rune('0'+i)),
			Path:   "/menu" + string(rune('0'+i)),
			Type:   MenuTypeMenu,
			Sort:   i,
			Status: 1,
		})
	}

	// 查询
	menus, total, err := biz.MenuList(ctx, 1, 10, "", "")
	if err != nil {
		t.Fatalf("MenuList failed: %v", err)
	}
	if total < 3 {
		t.Errorf("expected at least 3 menus, got %d", total)
	}
	if len(menus) < 3 {
		t.Errorf("expected at least 3 menus in list, got %d", len(menus))
	}
}

// TestMenuGetByRoleIDs 测试根据角色ID获取菜单
func TestMenuGetByRoleIDs(t *testing.T) {
	biz := newTestBiz()
	ctx := context.Background()

	// 创建菜单
	menu1 := &Menu{Name: "角色菜单1", Path: "/role1", Type: MenuTypeMenu, Sort: 1, Status: 1}
	menu2 := &Menu{Name: "角色菜单2", Path: "/role2", Type: MenuTypeMenu, Sort: 2, Status: 1}
	_ = biz.MenuCreate(ctx, menu1)
	_ = biz.MenuCreate(ctx, menu2)

	// 获取
	menus, err := biz.MenuGetByRoleIDs(ctx, []int64{1, 2})
	if err != nil {
		t.Fatalf("MenuGetByRoleIDs failed: %v", err)
	}
	if len(menus) < 2 {
		t.Errorf("expected at least 2 menus, got %d", len(menus))
	}
}

// TestAPIGetByID 测试根据ID获取API
func TestAPIGetByID(t *testing.T) {
	biz := newTestBiz()
	ctx := context.Background()

	api := &APIResource{Group: "用户", Path: "/api/user", Method: "GET"}
	_ = biz.APICreate(ctx, api)

	got, err := biz.APIRepo.GetByID(ctx, api.ID)
	if err != nil {
		t.Fatalf("APIGetByID failed: %v", err)
	}
	if got.Path != "/api/user" {
		t.Errorf("expected '/api/user', got '%s'", got.Path)
	}
}

// TestAPIList 测试API列表查询
func TestAPIList(t *testing.T) {
	biz := newTestBiz()
	ctx := context.Background()

	for i := 1; i <= 3; i++ {
		_ = biz.APICreate(ctx, &APIResource{
			Group:  "group" + string(rune('0'+i)),
			Path:   "/api/test" + string(rune('0'+i)),
			Method: "GET",
		})
	}

	apis, total, err := biz.APIList(ctx, 1, 10, "", "", "")
	if err != nil {
		t.Fatalf("APIList failed: %v", err)
	}
	if total < 3 {
		t.Errorf("expected at least 3 APIs, got %d", total)
	}
	if len(apis) < 3 {
		t.Errorf("expected at least 3 APIs in list, got %d", len(apis))
	}
}

// TestAPIGetAll 测试获取所有API
func TestAPIGetAll(t *testing.T) {
	biz := newTestBiz()
	ctx := context.Background()

	_ = biz.APICreate(ctx, &APIResource{Group: "g1", Path: "/p1", Method: "GET"})
	_ = biz.APICreate(ctx, &APIResource{Group: "g2", Path: "/p2", Method: "POST"})

	apis, err := biz.APIGetAll(ctx)
	if err != nil {
		t.Fatalf("APIGetAll failed: %v", err)
	}
	if len(apis) < 2 {
		t.Errorf("expected at least 2 APIs, got %d", len(apis))
	}
}

// TestButtonGetByRoleIDs 测试根据角色ID获取按钮
func TestButtonGetByRoleIDs(t *testing.T) {
	biz := newTestBiz()
	ctx := context.Background()

	btn1 := &Button{MenuID: 1, Name: "按钮1", Code: "btn1"}
	btn2 := &Button{MenuID: 1, Name: "按钮2", Code: "btn2"}
	_ = biz.ButtonCreate(ctx, btn1)
	_ = biz.ButtonCreate(ctx, btn2)

	btns, err := biz.ButtonGetByRoleIDs(ctx, []int64{1})
	if err != nil {
		t.Fatalf("ButtonGetByRoleIDs failed: %v", err)
	}
	if len(btns) < 2 {
		t.Errorf("expected at least 2 buttons, got %d", len(btns))
	}
}

// TestUserSetPassword 测试设置用户密码
func TestUserSetPassword(t *testing.T) {
	biz := newTestBiz()
	ctx := context.Background()

	user := &User{Username: "setpwd", Password: "abc123", Nickname: "设置密码", Status: UserStatusNormal}
	_ = biz.UserCreate(ctx, user)

	err := biz.UserSetPassword(ctx, user.ID, "newhashedpwd")
	if err != nil {
		t.Fatalf("UserSetPassword failed: %v", err)
	}

	got, _ := biz.UserRepo.GetByID(ctx, user.ID)
	if got.Password != "newhashedpwd" {
		t.Errorf("expected 'newhashedpwd', got '%s'", got.Password)
	}
}

// TestUserUpdateLastLogin 测试更新最后登录信息
func TestUserUpdateLastLogin(t *testing.T) {
	biz := newTestBiz()
	ctx := context.Background()

	user := &User{Username: "loginuser", Password: "abc123", Nickname: "登录用户", Status: UserStatusNormal}
	_ = biz.UserCreate(ctx, user)

	err := biz.UserUpdateLastLogin(ctx, user.ID, "192.168.1.1")
	if err != nil {
		t.Fatalf("UserUpdateLastLogin failed: %v", err)
	}

	got, _ := biz.UserRepo.GetByID(ctx, user.ID)
	if got.LastLoginIP != "192.168.1.1" {
		t.Errorf("expected '192.168.1.1', got '%s'", got.LastLoginIP)
	}
	if got.LastLoginAt == nil {
		t.Error("expected LastLoginAt to be set")
	}
}

// TestUserGetUserMenusNoRoles 测试用户无角色时获取菜单
func TestUserGetUserMenusNoRoles(t *testing.T) {
	biz := newTestBiz()
	ctx := context.Background()

	user := &User{Username: "noroleuser", Password: "abc123", Nickname: "无角色用户", Status: UserStatusNormal}
	_ = biz.UserCreate(ctx, user)

	menus, err := biz.MenuGetUserMenus(ctx, user.ID)
	if err != nil {
		t.Fatalf("MenuGetUserMenus failed: %v", err)
	}
	if len(menus) != 0 {
		t.Errorf("expected empty menus, got %d", len(menus))
	}
}

// TestBuildUserMenuTreeEmpty 测试空菜单树
func TestBuildUserMenuTreeEmpty(t *testing.T) {
	tree := buildUserMenuTree(nil)
	if tree != nil {
		t.Errorf("expected nil for nil input, got %v", tree)
	}

	tree = buildUserMenuTree([]*Menu{})
	if len(tree) != 0 {
		t.Errorf("expected empty slice, got %d", len(tree))
	}
}

// TestBuildUserMenuTreeNoRoots 测试无根节点时返回所有菜单
func TestBuildUserMenuTreeNoRoots(t *testing.T) {
	// 所有菜单都有父节点，但没有父节点存在
	menus := []*Menu{
		{ID: 1, ParentID: 99, Name: "孤儿菜单1"},
		{ID: 2, ParentID: 99, Name: "孤儿菜单2"},
	}

	tree := buildUserMenuTree(menus)
	// 当没有根节点时，返回原菜单列表
	if len(tree) != 2 {
		t.Errorf("expected 2 menus when no roots, got %d", len(tree))
	}
}

// TestMenuGetByIDDirect 测试MenuGetByID方法
func TestMenuGetByIDDirect(t *testing.T) {
	biz := newTestBiz()
	ctx := context.Background()

	// 创建菜单
	menu := &Menu{Name: "GetByID测试", Path: "/getbyid", Type: MenuTypeMenu, Sort: 1, Status: 1}
	_ = biz.MenuCreate(ctx, menu)

	// 正常获取
	got, err := biz.MenuGetByID(ctx, menu.ID)
	if err != nil {
		t.Fatalf("MenuGetByID failed: %v", err)
	}
	if got.Name != "GetByID测试" {
		t.Errorf("expected 'GetByID测试', got '%s'", got.Name)
	}

	// 获取不存在的菜单（通过直接调用Repo）
	_, err = biz.MenuRepo.GetByID(ctx, 9999)
	if err == nil {
		t.Error("expected error for non-existent menu")
	}
}

// TestAPIGetByIDDirect 测试APIGetByID方法
func TestAPIGetByIDDirect(t *testing.T) {
	biz := newTestBiz()
	ctx := context.Background()

	// 创建API
	api := &APIResource{Group: "GetByID", Path: "/api/getbyid", Method: "GET"}
	_ = biz.APICreate(ctx, api)

	// 正常获取
	got, err := biz.APIGetByID(ctx, api.ID)
	if err != nil {
		t.Fatalf("APIGetByID failed: %v", err)
	}
	if got.Path != "/api/getbyid" {
		t.Errorf("expected '/api/getbyid', got '%s'", got.Path)
	}

	// 获取不存在的API
	_, err = biz.APIRepo.GetByID(ctx, 9999)
	if err == nil {
		t.Error("expected error for non-existent API")
	}
}

// TestMenuGetUserMenusWithEmptyMenus 测试用户有角色但无菜单的情况
func TestMenuGetUserMenusWithEmptyMenus(t *testing.T) {
	biz := newTestBiz()
	ctx := context.Background()

	// 创建角色（不分配菜单）
	role := &Role{Name: "空菜单角色", Code: "empty_menu", Status: 1}
	_ = biz.RoleRepo.Create(ctx, role)

	// 创建用户并分配角色
	user := &User{Username: "emptymenuuser", Password: "abc123", Nickname: "空菜单用户", Status: UserStatusNormal}
	_ = biz.UserCreate(ctx, user)

	// 先清空菜单仓储
	menuRepo := biz.MenuRepo.(*MockMenuRepository)
	for id := range menuRepo.menus {
		delete(menuRepo.menus, id)
	}

	_ = biz.UserUpdateRoles(ctx, user.ID, []int64{role.ID})

	// 获取用户菜单
	menus, err := biz.MenuGetUserMenus(ctx, user.ID)
	if err != nil {
		t.Fatalf("MenuGetUserMenus failed: %v", err)
	}
	if len(menus) != 0 {
		t.Errorf("expected empty menus, got %d", len(menus))
	}
}
