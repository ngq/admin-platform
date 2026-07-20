package iam

import "time"

// Role 角色领域实体
type Role struct {
	ID          int64     `json:"id"`          // 角色唯一标识
	Name        string    `json:"name"`        // 角色名称
	Code        string    `json:"code"`        // 角色编码（唯一标识符）
	ParentID    int64     `json:"parent_id"`   // 父角色ID
	Sort        int       `json:"sort"`        // 排序序号
	DataScope   int8      `json:"data_scope"`  // 数据权限范围：1-全部 2-本部门 3-本部门及下级 4-仅本人
	Status      int8      `json:"status"`      // 状态
	Description string    `json:"description"` // 角色描述
	CreatedAt   time.Time `json:"created_at"`  // 创建时间
	UpdatedAt   time.Time `json:"updated_at"`  // 更新时间
}

// DataScope 数据权限范围常量
const (
	DataScopeAll     int8 = 1 // 全部数据
	DataScopeDept    int8 = 2 // 本部门数据
	DataScopeDeptSub int8 = 3 // 本部门及下级数据
	DataScopeSelf    int8 = 4 // 仅本人数据
)
