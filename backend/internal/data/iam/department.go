package iam

import (
	"context"
	"fmt"

	"github.com/ngq/admin-platform/internal/biz/iam"
	"github.com/ngq/admin-platform/internal/data/po"
	"github.com/ngq/admin-platform/internal/datainfra"
)

type departmentRepo struct {
	data *datainfra.Data
}

func NewDepartmentRepo(d *datainfra.Data) iam.DepartmentRepository {
	return &departmentRepo{data: d}
}

func (r *departmentRepo) Create(ctx context.Context, dept *iam.Department) error {
	deptPO := po.DepartmentPOFromEntity(dept)
	if err := r.data.DB(ctx).Create(deptPO).Error; err != nil {
		return err
	}
	dept.ID = deptPO.ID
	return nil
}

func (r *departmentRepo) Update(ctx context.Context, dept *iam.Department) error {
	deptPO := po.DepartmentPOFromEntity(dept)
	return r.data.DB(ctx).Model(&po.DepartmentPO{}).
		Where("id = ?", dept.ID).
		Updates(map[string]interface{}{
			"parent_id":      deptPO.ParentID,
			"name":           deptPO.Name,
			"leader_user_id": deptPO.LeaderUserID,
			"sort":           deptPO.Sort,
			"status":         deptPO.Status,
		}).Error
}

func (r *departmentRepo) Delete(ctx context.Context, id int64) error {
	var childCount int64
	if err := r.data.DB(ctx).Model(&po.DepartmentPO{}).Where("parent_id = ?", id).Count(&childCount).Error; err != nil {
		return err
	}
	if childCount > 0 {
		return fmt.Errorf("存在下级部门，不能删除")
	}
	var userCount int64
	if err := r.data.DB(ctx).Model(&po.AdminUserPO{}).Where("dept_id = ?", id).Count(&userCount).Error; err != nil {
		return err
	}
	if userCount > 0 {
		return fmt.Errorf("存在归属该部门的用户，不能删除")
	}
	return r.data.DB(ctx).Delete(&po.DepartmentPO{}, id).Error
}

func (r *departmentRepo) GetByID(ctx context.Context, id int64) (*iam.Department, error) {
	var deptPO po.DepartmentPO
	if err := r.data.DB(ctx).First(&deptPO, id).Error; err != nil {
		return nil, err
	}
	return deptPO.ToEntity(), nil
}

func (r *departmentRepo) List(ctx context.Context, name, status string) ([]*iam.Department, error) {
	var list []po.DepartmentPO
	db := r.data.DB(ctx).Model(&po.DepartmentPO{})
	if name != "" {
		db = db.Where("name LIKE ?", "%"+name+"%")
	}
	if status != "" {
		db = db.Where("status = ?", status)
	}
	if err := db.Order("sort ASC, id ASC").Find(&list).Error; err != nil {
		return nil, err
	}
	depts := make([]*iam.Department, 0, len(list))
	for i := range list {
		depts = append(depts, list[i].ToEntity())
	}
	return depts, nil
}

func (r *departmentRepo) GetDescendantIDs(ctx context.Context, id int64) ([]int64, error) {
	all, err := r.List(ctx, "", "")
	if err != nil {
		return nil, err
	}
	children := make(map[int64][]int64, len(all))
	for _, dept := range all {
		children[dept.ParentID] = append(children[dept.ParentID], dept.ID)
	}
	result := []int64{id}
	queue := []int64{id}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		for _, childID := range children[current] {
			result = append(result, childID)
			queue = append(queue, childID)
		}
	}
	return result, nil
}

func (r *departmentRepo) GetManagedDeptIDs(ctx context.Context, leaderUserID int64) ([]int64, error) {
	if leaderUserID <= 0 {
		return []int64{}, nil
	}
	var ids []int64
	if err := r.data.DB(ctx).
		Model(&po.DepartmentPO{}).
		Where("leader_user_id = ?", leaderUserID).
		Pluck("id", &ids).Error; err != nil {
		return nil, err
	}
	return ids, nil
}
