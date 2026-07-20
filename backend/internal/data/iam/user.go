// 用户仓储实现
// 基于 GORM 实现 iam.UserRepository 接口
package iam

import (
	"context"
	"fmt"
	"github.com/ngq/admin-platform/internal/datainfra"
	"time"

	"github.com/ngq/admin-platform/internal/biz/iam"
	"github.com/ngq/admin-platform/internal/data/po"
	"gorm.io/gorm"
)

// userRepo 用户仓储实现
type userRepo struct {
	data *datainfra.Data
}

// NewUserRepo 创建用户仓储实例
func NewUserRepo(d *datainfra.Data) iam.UserRepository {
	return &userRepo{data: d}
}

// Create 创建用户，同时插入用户-角色关联
func (r *userRepo) Create(ctx context.Context, user *iam.User) error {
	userPO := po.AdminUserPOFromEntity(user)
	// 使用事务保证用户记录与角色关联的原子性
	return r.data.DB(ctx).Transaction(func(tx *gorm.DB) error {
		// 先创建用户记录
		if err := tx.Create(userPO).Error; err != nil {
			return err
		}
		// 如果有角色关联，则批量插入
		if len(userPO.Roles) > 0 {
			if err := tx.Model(userPO).Association("Roles").Replace(userPO.Roles); err != nil {
				return err
			}
		}
		user.ID = userPO.ID
		return nil
	})
}

// Update 更新用户基本信息（不包含角色和密码）
func (r *userRepo) Update(ctx context.Context, user *iam.User) error {
	userPO := po.AdminUserPOFromEntity(user)
	return r.data.DB(ctx).Model(&po.AdminUserPO{}).
		Where("id = ?", user.ID).
		Updates(map[string]interface{}{
			"username": userPO.Username,
			"nickname": userPO.Nickname,
			"dept_id":  userPO.DeptID,
			"email":    userPO.Email,
			"phone":    userPO.Phone,
			"avatar":   userPO.Avatar,
			"status":   userPO.Status,
		}).Error
}

// Delete 软删除用户，同时清理角色关联表
func (r *userRepo) Delete(ctx context.Context, id int64) error {
	return r.data.DB(ctx).Transaction(func(tx *gorm.DB) error {
		// 先清理用户-角色关联表，避免孤立记录
		if err := tx.Table("sys_user_roles").Where("user_id = ?", id).Delete(nil).Error; err != nil {
			return err
		}
		return tx.Delete(&po.AdminUserPO{}, id).Error
	})
}

// GetByID 根据ID获取用户，预加载角色列表
func (r *userRepo) GetByID(ctx context.Context, id int64) (*iam.User, error) {
	var userPO po.AdminUserPO
	if err := r.data.DB(ctx).Preload("Roles").First(&userPO, id).Error; err != nil {
		return nil, err
	}
	user := userPO.ToEntity()
	r.attachDeptNames(ctx, []*iam.User{user})
	return user, nil
}

// GetByUsername 根据用户名获取用户，预加载角色列表
func (r *userRepo) GetByUsername(ctx context.Context, username string) (*iam.User, error) {
	var userPO po.AdminUserPO
	if err := r.data.DB(ctx).Preload("Roles").Where("username = ?", username).First(&userPO).Error; err != nil {
		return nil, err
	}
	user := userPO.ToEntity()
	r.attachDeptNames(ctx, []*iam.User{user})
	return user, nil
}

// List 分页查询用户列表，支持按用户名模糊搜索和状态过滤
func (r *userRepo) List(ctx context.Context, page, pageSize int, username, status string) ([]*iam.User, int64, error) {
	var list []po.AdminUserPO
	var count int64

	db := r.data.DB(ctx).Model(&po.AdminUserPO{})

	// 按用户名模糊搜索
	if username != "" {
		db = db.Where("username LIKE ?", "%"+username+"%")
	}
	// 按状态过滤
	if status != "" {
		db = db.Where("status = ?", status)
	}

	// 分页查询并预加载角色
	if err := datainfra.Paginate(db.Preload("Roles"), page, pageSize, "id DESC", &list, &count); err != nil {
		return nil, 0, err
	}

	// 转换为领域实体
	users := make([]*iam.User, 0, len(list))
	for i := range list {
		users = append(users, list[i].ToEntity())
	}
	r.attachDeptNames(ctx, users)
	return users, count, nil
}

// ListByDataScope 分页查询用户列表，并按当前登录人的数据范围过滤。
func (r *userRepo) ListByDataScope(ctx context.Context, page, pageSize int, username, status string, deptID int64, scope iam.DataPermissionScope) ([]*iam.User, int64, error) {
	var list []po.AdminUserPO
	var count int64

	db := r.data.DB(ctx).Model(&po.AdminUserPO{})
	if username != "" {
		db = db.Where("username LIKE ?", "%"+username+"%")
	}
	if status != "" {
		db = db.Where("status = ?", status)
	}
	if deptID > 0 {
		db = db.Where("dept_id = ?", deptID)
	}

	switch scope.DataScope {
	case iam.DataScopeAll:
		// no-op
	case iam.DataScopeDept:
		if scope.DeptID <= 0 {
			db = db.Where("id = ?", scope.UserID)
		} else {
			db = db.Where("dept_id = ?", scope.DeptID)
		}
	case iam.DataScopeDeptSub:
		deptIDs := scope.DeptIDs
		if len(deptIDs) == 0 && scope.DeptID > 0 {
			deptIDs = []int64{scope.DeptID}
		}
		if len(deptIDs) == 0 {
			db = db.Where("id = ?", scope.UserID)
		} else {
			db = db.Where("dept_id IN ?", deptIDs)
		}
	default:
		db = db.Where("id = ?", scope.UserID)
	}

	if err := datainfra.Paginate(db.Preload("Roles"), page, pageSize, "id DESC", &list, &count); err != nil {
		return nil, 0, err
	}

	users := make([]*iam.User, 0, len(list))
	for i := range list {
		users = append(users, list[i].ToEntity())
	}
	r.attachDeptNames(ctx, users)
	return users, count, nil
}

// UpdateRoles 更新用户角色关联，采用替换模式
// 验证传入的角色 ID 全部存在，否则返回错误
func (r *userRepo) UpdateRoles(ctx context.Context, userID int64, roleIDs []int64) error {
	return r.data.DB(ctx).Transaction(func(tx *gorm.DB) error {
		var userPO po.AdminUserPO
		if err := tx.First(&userPO, userID).Error; err != nil {
			return err
		}
		// 查询需要关联的角色
		var roles []po.RolePO
		if len(roleIDs) > 0 {
			if err := tx.Where("id IN ?", roleIDs).Find(&roles).Error; err != nil {
				return err
			}
			// 验证所有角色 ID 都存在，避免静默忽略不存在的角色
			if len(roles) != len(roleIDs) {
				return fmt.Errorf("部分角色ID不存在，期望 %d 个，实际找到 %d 个", len(roleIDs), len(roles))
			}
		}
		// 替换角色关联
		return tx.Model(&userPO).Association("Roles").Replace(roles)
	})
}

// SetPassword 设置用户密码（哈希值）
func (r *userRepo) SetPassword(ctx context.Context, id int64, hashedPassword string) error {
	return r.data.DB(ctx).Model(&po.AdminUserPO{}).
		Where("id = ?", id).
		Update("password", hashedPassword).Error
}

// UpdateLastLogin 更新最后登录时间和IP
func (r *userRepo) UpdateLastLogin(ctx context.Context, id int64, ip string) error {
	now := time.Now()
	return r.data.DB(ctx).Model(&po.AdminUserPO{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"last_login_at": now,
			"last_login_ip": ip,
		}).Error
}

// Count 查询用户总数（用于仪表盘统计，避免 List 的额外开销）
func (r *userRepo) Count(ctx context.Context) (int64, error) {
	var count int64
	return count, r.data.DB(ctx).Model(&po.AdminUserPO{}).Count(&count).Error
}

func (r *userRepo) attachDeptNames(ctx context.Context, users []*iam.User) {
	deptIDs := make([]int64, 0)
	seen := make(map[int64]struct{})
	for _, user := range users {
		if user.DeptID <= 0 {
			continue
		}
		if _, ok := seen[user.DeptID]; ok {
			continue
		}
		seen[user.DeptID] = struct{}{}
		deptIDs = append(deptIDs, user.DeptID)
	}
	if len(deptIDs) == 0 {
		return
	}
	var depts []po.DepartmentPO
	if err := r.data.DB(ctx).Where("id IN ?", deptIDs).Find(&depts).Error; err != nil {
		return
	}
	names := make(map[int64]string, len(depts))
	for _, dept := range depts {
		names[dept.ID] = dept.Name
	}
	for _, user := range users {
		user.DeptName = names[user.DeptID]
	}
}
