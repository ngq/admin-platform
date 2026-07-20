package audit

import (
	"github.com/ngq/admin-platform/internal/biz/audit"
	"github.com/ngq/admin-platform/internal/datainfra"
)

// NewBiz 创建审计领域子聚合（DI 桥接）
func NewBiz(d *datainfra.Data) *audit.Biz {
	return audit.NewBiz(
		NewOperationLogRepo(d),
		NewLoginLogRepo(d),
		NewErrorLogRepo(d),
	)
}
