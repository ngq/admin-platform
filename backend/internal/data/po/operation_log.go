// Domain: Audit
// 操作日志持久化对象
// OperationLogPO 对应 sys_operation_logs 表，记录用户 API 请求操作
package po

import (
	"time"

	"github.com/ngq/admin-platform/internal/biz/audit"
)

// OperationLogPO 操作日志表持久化对象
type OperationLogPO struct {
	ID             int64     `gorm:"primaryKey;autoIncrement" db:"id"`
	UserID         int64     `gorm:"comment:操作用户ID" db:"user_id"`
	Username       string    `gorm:"type:varchar(64)" db:"username"`
	Method         string    `gorm:"type:varchar(16);not null" db:"method"`
	Path           string    `gorm:"type:varchar(256);not null" db:"path"`
	IP             string    `gorm:"type:varchar(45)" db:"ip"`
	UserAgent      string    `gorm:"type:varchar(256)" db:"user_agent"`
	RequestBody    string    `gorm:"type:text" db:"request_body"`
	ResponseStatus int       `gorm:"comment:HTTP响应状态码" db:"response_status"`
	LatencyMs      int       `gorm:"comment:请求耗时(毫秒)" db:"latency_ms"`
	CreatedAt      time.Time `gorm:"autoCreateTime;not null" db:"created_at"`
}

// TableName 指定表名
func (OperationLogPO) TableName() string { return "sys_operation_logs" }

// ToEntity 将 PO 转换为领域实体
func (o *OperationLogPO) ToEntity() *audit.OperationLog {
	return &audit.OperationLog{
		ID:             o.ID,
		UserID:         o.UserID,
		Username:       o.Username,
		Method:         o.Method,
		Path:           o.Path,
		IP:             o.IP,
		UserAgent:      o.UserAgent,
		RequestBody:    o.RequestBody,
		ResponseStatus: o.ResponseStatus,
		LatencyMs:      o.LatencyMs,
		CreatedAt:      o.CreatedAt,
	}
}

// OperationLogPOFromEntity 将领域实体转换为 PO
func OperationLogPOFromEntity(o *audit.OperationLog) *OperationLogPO {
	return &OperationLogPO{
		ID:             o.ID,
		UserID:         o.UserID,
		Username:       o.Username,
		Method:         o.Method,
		Path:           o.Path,
		IP:             o.IP,
		UserAgent:      o.UserAgent,
		RequestBody:    o.RequestBody,
		ResponseStatus: o.ResponseStatus,
		LatencyMs:      o.LatencyMs,
	}
}
