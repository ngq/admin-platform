// Package auth — 认证领域子聚合
package auth

// Biz 认证领域子聚合
type Biz struct {
	AuthRepo AuthRepository
}

// NewBiz 创建认证领域子聚合
func NewBiz(authRepo AuthRepository) *Biz {
	return &Biz{AuthRepo: authRepo}
}
