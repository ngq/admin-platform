package iam

import "time"

// Menu 菜单领域实体
type Menu struct {
	ID        int64     `json:"id"`         // 菜单唯一标识
	ParentID  int64     `json:"parent_id"`  // 父菜单ID
	Name      string    `json:"name"`       // 菜单名称
	Path      string    `json:"path"`       // 路由路径
	Component string    `json:"component"`  // 前端组件路径
	Redirect  string    `json:"redirect"`   // 重定向地址
	Icon      string    `json:"icon"`       // 菜单图标
	Sort      int       `json:"sort"`       // 排序序号
	Type      int8      `json:"type"`       // 菜单类型：1-目录 2-菜单 3-按钮
	Visible   int8      `json:"visible"`    // 是否可见
	Status    int8      `json:"status"`     // 状态
	CreatedAt time.Time `json:"created_at"` // 创建时间
	UpdatedAt time.Time `json:"updated_at"` // 更新时间
	Children  []*Menu   `json:"children"`   // 子菜单列表
}

// MenuType 菜单类型常量
const (
	MenuTypeDirectory int8 = 1 // 目录
	MenuTypeMenu      int8 = 2 // 菜单
	MenuTypeButton    int8 = 3 // 按钮
)
