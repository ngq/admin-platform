package auth

import (
	"context"
	"strings"
	"time"

	"github.com/mojocn/base64Captcha"
	datacontract "github.com/ngq/gorp/framework/contract/data"
)

// RedisCaptchaStore 基于框架 Redis 契约实现 base64Captcha.Store，支持多实例共享。
type RedisCaptchaStore struct {
	redis      datacontract.Redis
	prefix     string
	expiration time.Duration
}

// NewRedisCaptchaStore 创建验证码存储实例。
// 当 redis 为 nil 时，自动安全降级为 base64Captcha.DefaultMemStore 单机内存存储。
func NewRedisCaptchaStore(redis datacontract.Redis, expiration ...time.Duration) base64Captcha.Store {
	if redis == nil {
		return base64Captcha.DefaultMemStore
	}
	exp := 5 * time.Minute
	if len(expiration) > 0 && expiration[0] > 0 {
		exp = expiration[0]
	}
	return &RedisCaptchaStore{
		redis:      redis,
		prefix:     "captcha:",
		expiration: exp,
	}
}

// Set 写入验证码（带过期时间）
func (s *RedisCaptchaStore) Set(id string, value string) error {
	if s.redis == nil {
		return base64Captcha.DefaultMemStore.Set(id, value)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	return s.redis.Set(ctx, s.prefix+id, value, s.expiration)
}

// Get 读取验证码
func (s *RedisCaptchaStore) Get(id string, clear bool) string {
	if s.redis == nil {
		return base64Captcha.DefaultMemStore.Get(id, clear)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	val, err := s.redis.Get(ctx, s.prefix+id)
	if err != nil {
		return ""
	}
	if clear {
		_ = s.redis.Del(ctx, s.prefix+id)
	}
	return val
}

// Verify 校验验证码
func (s *RedisCaptchaStore) Verify(id, answer string, clear bool) bool {
	if s.redis == nil {
		return base64Captcha.DefaultMemStore.Verify(id, answer, clear)
	}
	v := s.Get(id, clear)
	return strings.ToLower(v) == strings.ToLower(answer) && v != ""
}
