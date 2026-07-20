package iam

import "time"

// Department 组织部门实体，用于后台数据范围过滤。
type Department struct {
	ID           int64         `json:"id"`
	ParentID     int64         `json:"parent_id"`
	Name         string        `json:"name"`
	LeaderUserID int64         `json:"leader_user_id"`
	Sort         int           `json:"sort"`
	Status       int8          `json:"status"`
	CreatedAt    time.Time     `json:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at"`
	Children     []*Department `json:"children,omitempty"`
}

const (
	DepartmentStatusNormal   int8 = 1
	DepartmentStatusDisabled int8 = 2
)
