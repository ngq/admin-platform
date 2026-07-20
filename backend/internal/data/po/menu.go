// Domain: IAM
// 菜单持久化对象
// MenuPO 对应 sys_menus 表，存储菜单树结构（目录/菜单/按钮三级）
package po

import (
	"time"

	"github.com/ngq/admin-platform/internal/biz/iam"
	"gorm.io/gorm"
)

// MenuPO 菜单表持久化对象
type MenuPO struct {
	ID        int64          `gorm:"primaryKey;autoIncrement" db:"id"`
	ParentID  int64          `gorm:"default:0;comment:父级ID, 0为顶级" db:"parent_id"`
	Name      string         `gorm:"type:varchar(64);not null" db:"name"`
	Path      string         `gorm:"type:varchar(256)" db:"path"`
	Component string         `gorm:"type:varchar(256)" db:"component"`
	Redirect  string         `gorm:"type:varchar(256)" db:"redirect"`
	Icon      string         `gorm:"type:varchar(64)" db:"icon"`
	Sort      int            `gorm:"default:0" db:"sort"`
	Type      int8           `gorm:"type:tinyint;not null;comment:1-目录 2-菜单 3-按钮" db:"type"` // 1-目录 2-菜单 3-按钮
	Visible   int8           `gorm:"type:tinyint;default:1;comment:1-显示 2-隐藏" db:"visible"`
	Status    int8           `gorm:"type:tinyint;default:1;comment:1-正常 2-禁用" db:"status"`
	CreatedAt time.Time      `gorm:"autoCreateTime" db:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" db:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" db:"deleted_at"`
	Children  []MenuPO       `gorm:"-" db:"-"` // 子菜单，不映射数据库
}

// TableName 指定表名
func (MenuPO) TableName() string { return "sys_menus" }

// ToEntity 将 PO 转换为领域实体
func (m *MenuPO) ToEntity() *iam.Menu {
	menu := &iam.Menu{
		ID:        m.ID,
		ParentID:  m.ParentID,
		Name:      m.Name,
		Path:      m.Path,
		Component: m.Component,
		Redirect:  m.Redirect,
		Icon:      m.Icon,
		Sort:      m.Sort,
		Type:      m.Type,
		Visible:   m.Visible,
		Status:    m.Status,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
	for _, c := range m.Children {
		menu.Children = append(menu.Children, c.ToEntity())
	}
	return menu
}

// MenuPOFromEntity 将领域实体转换为 PO
func MenuPOFromEntity(m *iam.Menu) *MenuPO {
	return &MenuPO{
		ID:        m.ID,
		ParentID:  m.ParentID,
		Name:      m.Name,
		Path:      m.Path,
		Component: m.Component,
		Redirect:  m.Redirect,
		Icon:      m.Icon,
		Sort:      m.Sort,
		Type:      m.Type,
		Visible:   m.Visible,
		Status:    m.Status,
	}
}
