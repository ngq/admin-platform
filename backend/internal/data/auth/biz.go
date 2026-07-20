package auth

import (
	"github.com/ngq/admin-platform/internal/biz/auth"
)

// NewBiz 创建认证领域子聚合（DI 桥接）
func NewBiz() *auth.Biz {
	return auth.NewBiz(NewAuthRepo())
}
