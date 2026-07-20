// 认证仓储实现
// 使用内存 map + sync.RWMutex 实现 JWT 黑名单
// 生产环境建议替换为 Redis 实现
package auth

import (
	"context"
	"sync"
	"time"

	"github.com/ngq/admin-platform/internal/biz/auth"
)

// authRepo 认证仓储实现（基于内存的 JWT 黑名单）
type authRepo struct {
	mu        sync.RWMutex     // 读写锁，保证并发安全
	blacklist map[string]int64 // 黑名单：token 字符串 -> 过期时间戳（Unix 秒）
	stopCh    chan struct{}    // 停止信号通道
}

// NewAuthRepo 创建认证仓储实例
func NewAuthRepo() auth.AuthRepository {
	repo := &authRepo{
		blacklist: make(map[string]int64),
		stopCh:    make(chan struct{}),
	}
	// 启动后台清理协程，定期清除已过期的黑名单条目
	go repo.cleanupExpired()
	return repo
}

// Close 停止后台清理协程，实现优雅关闭
func (r *authRepo) Close() {
	close(r.stopCh)
}

// TokenBlacklist 将 JWT 加入黑名单
// tokenString: JWT 字符串
// expiresAt: Token 的原始过期时间（Unix 时间戳，秒），过期后自动清理
func (r *authRepo) TokenBlacklist(ctx context.Context, tokenString string, expiresAt int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.blacklist[tokenString] = expiresAt
	return nil
}

// IsTokenBlacklisted 检查 Token 是否在黑名单中
func (r *authRepo) IsTokenBlacklisted(ctx context.Context, tokenString string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	expiresAt, exists := r.blacklist[tokenString]
	if !exists {
		return false, nil
	}
	// 如果 Token 已过期，视为不在黑名单中（已自动失效）
	if time.Now().Unix() > expiresAt {
		return false, nil
	}
	return true, nil
}

// cleanupExpired 后台定期清理已过期的黑名单条目
// 每分钟执行一次清理，避免黑名单无限增长
func (r *authRepo) cleanupExpired() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			r.mu.Lock()
			now := time.Now().Unix()
			for tokenStr, expiresAt := range r.blacklist {
				if now > expiresAt {
					delete(r.blacklist, tokenStr)
				}
			}
			r.mu.Unlock()
		case <-r.stopCh:
			// 收到停止信号，退出协程
			return
		}
	}
}
