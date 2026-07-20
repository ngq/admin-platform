package po

import (
	"time"

	"github.com/ngq/admin-platform/internal/biz/iam"
	"gorm.io/gorm"
)

// APIPO maps sys_apis.
type APIPO struct {
	ID             int64          `gorm:"primaryKey;autoIncrement" db:"id"`
	Group          string         `gorm:"type:varchar(64);not null;comment:API分组" db:"group_name"`
	Path           string         `gorm:"type:varchar(256);not null;uniqueIndex:idx_api_path_method" db:"path"`
	Method         string         `gorm:"type:varchar(16);not null;uniqueIndex:idx_api_path_method" db:"method"`
	Description    string         `gorm:"type:varchar(256)" db:"description"`
	MenuID         int64          `gorm:"type:bigint unsigned;not null;default:0;index;comment:所属菜单ID" db:"menu_id"`
	ActionType     string         `gorm:"type:varchar(32);not null;default:'other';comment:功能类型" db:"action_type"`
	PermissionMode string         `gorm:"type:varchar(16);not null;default:'casbin';index;comment:权限模式" db:"permission_mode"`
	Ignored        int8           `gorm:"type:tinyint;default:0;comment:1-忽略权限校验" db:"ignored"`
	CreatedAt      time.Time      `gorm:"autoCreateTime" db:"created_at"`
	UpdatedAt      time.Time      `gorm:"autoUpdateTime" db:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" db:"deleted_at"`
}

func (APIPO) TableName() string { return "sys_apis" }

func (a *APIPO) ToEntity() *iam.APIResource {
	return &iam.APIResource{
		ID:             a.ID,
		Group:          a.Group,
		Path:           a.Path,
		Method:         a.Method,
		Description:    a.Description,
		MenuID:         a.MenuID,
		ActionType:     normalizeAPIActionType(a.ActionType),
		PermissionMode: normalizeAPIPermissionMode(a.PermissionMode, a.Ignored),
		Ignored:        a.Ignored,
		CreatedAt:      a.CreatedAt,
		UpdatedAt:      a.UpdatedAt,
	}
}

func APIPOFromEntity(a *iam.APIResource) *APIPO {
	mode := normalizeAPIPermissionMode(a.PermissionMode, a.Ignored)
	ignored := a.Ignored
	if mode == "ignored" {
		ignored = 1
	}
	return &APIPO{
		ID:             a.ID,
		Group:          a.Group,
		Path:           a.Path,
		Method:         a.Method,
		Description:    a.Description,
		MenuID:         a.MenuID,
		ActionType:     normalizeAPIActionType(a.ActionType),
		PermissionMode: mode,
		Ignored:        ignored,
	}
}

func normalizeAPIActionType(actionType string) string {
	if actionType == "" {
		return "other"
	}
	return actionType
}

func normalizeAPIPermissionMode(mode string, ignored int8) string {
	if ignored == 1 {
		return "ignored"
	}
	if mode != "" {
		return mode
	}
	return "casbin"
}
