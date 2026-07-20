package auth

import "context"

// AuthRepository 认证仓储接口（黑名单等）
type AuthRepository interface {
	// TokenBlacklist 将JWT加入黑名单
	TokenBlacklist(ctx context.Context, tokenString string, expiresAt int64) error
	// IsTokenBlacklisted 检查Token是否在黑名单中
	IsTokenBlacklisted(ctx context.Context, tokenString string) (bool, error)
}
