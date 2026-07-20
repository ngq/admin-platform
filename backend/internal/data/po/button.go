// Domain: IAM
// 按钮权限持久化对象
// ButtonPO 对应 sys_buttons 表，存储页面按钮级别的权限标识
package po

import (
	"time"

	"github.com/ngq/admin-platform/internal/biz/iam"
)

// ButtonPO 按钮权限表持久化对象
type ButtonPO struct {
	ID          int64     `gorm:"primaryKey;autoIncrement" db:"id"`
	MenuID      int64     `gorm:"not null;uniqueIndex:idx_menu_button_code" db:"menu_id"`
	Name        string    `gorm:"type:varchar(64);not null" db:"name"`
	Code        string    `gorm:"type:varchar(64);not null;uniqueIndex:idx_menu_button_code" db:"code"`
	Description string    `gorm:"type:varchar(256)" db:"description"`
	CreatedAt   time.Time `gorm:"autoCreateTime" db:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" db:"updated_at"`
}

// TableName 指定表名
func (ButtonPO) TableName() string { return "sys_buttons" }

// ToEntity 将 PO 转换为领域实体
func (b *ButtonPO) ToEntity() *iam.Button {
	return &iam.Button{
		ID:          b.ID,
		MenuID:      b.MenuID,
		Name:        b.Name,
		Code:        b.Code,
		Description: b.Description,
		CreatedAt:   b.CreatedAt,
		UpdatedAt:   b.UpdatedAt,
	}
}

// ButtonPOFromEntity 将领域实体转换为 PO
func ButtonPOFromEntity(b *iam.Button) *ButtonPO {
	return &ButtonPO{
		ID:          b.ID,
		MenuID:      b.MenuID,
		Name:        b.Name,
		Code:        b.Code,
		Description: b.Description,
	}
}
