package auth

import (
	"context"
	"testing"
	"time"
)

type mockRedisClient struct {
	data map[string]string
}

func (m *mockRedisClient) Ping(ctx context.Context) error { return nil }
func (m *mockRedisClient) Get(ctx context.Context, key string) (string, error) {
	if val, ok := m.data[key]; ok {
		return val, nil
	}
	return "", nil
}
func (m *mockRedisClient) Set(ctx context.Context, key, val string, ttl time.Duration) error {
	m.data[key] = val
	return nil
}
func (m *mockRedisClient) Del(ctx context.Context, key string) error {
	delete(m.data, key)
	return nil
}
func (m *mockRedisClient) MGet(ctx context.Context, keys ...string) (map[string]string, error) {
	res := make(map[string]string)
	for _, k := range keys {
		if v, ok := m.data[k]; ok {
			res[k] = v
		}
	}
	return res, nil
}
func (m *mockRedisClient) MSet(ctx context.Context, kvs map[string]string) error {
	for k, v := range kvs {
		m.data[k] = v
	}
	return nil
}
func (m *mockRedisClient) Expire(ctx context.Context, key string, ttl time.Duration) error {
	return nil
}

func TestRedisCaptchaStore(t *testing.T) {
	mock := &mockRedisClient{data: make(map[string]string)}
	store := NewRedisCaptchaStore(mock, time.Minute)

	err := store.Set("c1", "1234")
	if err != nil {
		t.Fatalf("unexpected set error: %v", err)
	}

	if val := store.Get("c1", false); val != "1234" {
		t.Fatalf("expected 1234, got %s", val)
	}

	if !store.Verify("c1", "1234", true) {
		t.Fatal("expected verify to succeed")
	}

	if val := store.Get("c1", false); val != "" {
		t.Fatalf("expected cleared captcha to be empty, got %s", val)
	}
}

func TestRedisCaptchaStore_NilFallback(t *testing.T) {
	store := NewRedisCaptchaStore(nil)
	if err := store.Set("mem1", "5678"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !store.Verify("mem1", "5678", true) {
		t.Fatal("expected fallback memory verify to succeed")
	}
}
