package iam

import "time"

// Button 按钮权限领域实体
type Button struct {
	ID          int64     `json:"id"`          // 按钮唯一标识
	MenuID      int64     `json:"menu_id"`     // 所属菜单ID
	Name        string    `json:"name"`        // 按钮名称
	Code        string    `json:"code"`        // 按钮权限编码
	Description string    `json:"description"` // 按钮描述
	CreatedAt   time.Time `json:"created_at"`  // 创建时间
	UpdatedAt   time.Time `json:"updated_at"`  // 更新时间
}
