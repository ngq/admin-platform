// Domain: Audit
// 错误日志持久化对象
// ErrorLogPO 对应 sys_error_logs 表，记录系统运行时错误
package po

import (
	"time"

	"github.com/ngq/admin-platform/internal/biz/audit"
)

// ErrorLogPO 错误日志表持久化对象
type ErrorLogPO struct {
	ID           int64     `gorm:"primaryKey;autoIncrement" db:"id"`
	UserID       int64     `gorm:"comment:关联用户ID" db:"user_id"`
	Username     string    `gorm:"type:varchar(64)" db:"username"`
	Method       string    `gorm:"type:varchar(16)" db:"method"`
	Path         string    `gorm:"type:varchar(256)" db:"path"`
	ErrorMessage string    `gorm:"type:text" db:"error_message"`
	Stack        string    `gorm:"type:text" db:"stack"`
	Status       int8      `gorm:"type:tinyint;default:1;comment:1-未处理 2-已处理" db:"status"`
	CreatedAt    time.Time `gorm:"autoCreateTime" db:"created_at"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime" db:"updated_at"`
}

// TableName 指定表名
func (ErrorLogPO) TableName() string { return "sys_error_logs" }

// ToEntity 将 PO 转换为领域实体
func (e *ErrorLogPO) ToEntity() *audit.ErrorLog {
	return &audit.ErrorLog{
		ID:           e.ID,
		UserID:       e.UserID,
		Username:     e.Username,
		Method:       e.Method,
		Path:         e.Path,
		ErrorMessage: e.ErrorMessage,
		Stack:        e.Stack,
		Status:       e.Status,
		CreatedAt:    e.CreatedAt,
		UpdatedAt:    e.UpdatedAt,
	}
}

// ErrorLogPOFromEntity 将领域实体转换为 PO
func ErrorLogPOFromEntity(e *audit.ErrorLog) *ErrorLogPO {
	return &ErrorLogPO{
		ID:           e.ID,
		UserID:       e.UserID,
		Username:     e.Username,
		Method:       e.Method,
		Path:         e.Path,
		ErrorMessage: e.ErrorMessage,
		Stack:        e.Stack,
		Status:       e.Status,
	}
}
