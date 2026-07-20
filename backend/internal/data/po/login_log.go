// Domain: Audit
// 登录日志持久化对象
// LoginLogPO 对应 sys_login_logs 表，记录用户登录/登出事件
package po

import (
	"time"

	"github.com/ngq/admin-platform/internal/biz/audit"
)

// LoginLogPO 登录日志表持久化对象
type LoginLogPO struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" db:"id"`
	UserID    int64     `gorm:"comment:登录用户ID" db:"user_id"`
	Username  string    `gorm:"type:varchar(64)" db:"username"`
	IP        string    `gorm:"type:varchar(45)" db:"ip"`
	Location  string    `gorm:"type:varchar(128)" db:"location"`
	UserAgent string    `gorm:"type:varchar(256)" db:"user_agent"`
	Status    int8      `gorm:"type:tinyint;not null;comment:1-成功 2-失败" db:"status"`
	Message   string    `gorm:"type:varchar(256)" db:"message"`
	CreatedAt time.Time `gorm:"autoCreateTime" db:"created_at"`
}

// TableName 指定表名
func (LoginLogPO) TableName() string { return "sys_login_logs" }

// ToEntity 将 PO 转换为领域实体
func (l *LoginLogPO) ToEntity() *audit.LoginLog {
	return &audit.LoginLog{
		ID:        l.ID,
		UserID:    l.UserID,
		Username:  l.Username,
		IP:        l.IP,
		Location:  l.Location,
		UserAgent: l.UserAgent,
		Status:    l.Status,
		Message:   l.Message,
		CreatedAt: l.CreatedAt,
	}
}

// LoginLogPOFromEntity 将领域实体转换为 PO
func LoginLogPOFromEntity(l *audit.LoginLog) *LoginLogPO {
	return &LoginLogPO{
		ID:        l.ID,
		UserID:    l.UserID,
		Username:  l.Username,
		IP:        l.IP,
		Location:  l.Location,
		UserAgent: l.UserAgent,
		Status:    l.Status,
		Message:   l.Message,
	}
}
