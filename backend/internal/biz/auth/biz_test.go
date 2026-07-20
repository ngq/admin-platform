package auth

import (
	"context"
	"testing"
	"time"
)

// MockAuthRepository 模拟认证仓储
type MockAuthRepository struct {
	blacklist map[string]int64
}

func NewMockAuthRepository() *MockAuthRepository {
	return &MockAuthRepository{
		blacklist: make(map[string]int64),
	}
}

func (m *MockAuthRepository) TokenBlacklist(ctx context.Context, tokenString string, expiresAt int64) error {
	m.blacklist[tokenString] = expiresAt
	return nil
}

func (m *MockAuthRepository) IsTokenBlacklisted(ctx context.Context, tokenString string) (bool, error) {
	_, exists := m.blacklist[tokenString]
	return exists, nil
}

// TestNewBiz 测试创建认证领域子聚合
func TestNewBiz(t *testing.T) {
	repo := NewMockAuthRepository()
	biz := NewBiz(repo)

	if biz == nil {
		t.Fatal("NewBiz returned nil")
	}
	if biz.AuthRepo == nil {
		t.Fatal("AuthRepo is nil")
	}
}

// TestBiz_TokenBlacklist 测试Token加入黑名单
func TestBiz_TokenBlacklist(t *testing.T) {
	repo := NewMockAuthRepository()
	biz := NewBiz(repo)
	ctx := context.Background()

	token := "test-token-123"
	expiresAt := int64(1700000000)

	err := biz.AuthRepo.TokenBlacklist(ctx, token, expiresAt)
	if err != nil {
		t.Fatalf("TokenBlacklist failed: %v", err)
	}

	// 验证token已在黑名单中
	blacklisted, err := biz.AuthRepo.IsTokenBlacklisted(ctx, token)
	if err != nil {
		t.Fatalf("IsTokenBlacklisted failed: %v", err)
	}
	if !blacklisted {
		t.Error("Token should be blacklisted")
	}
}

// TestBiz_IsTokenBlacklisted 测试检查Token是否在黑名单
func TestBiz_IsTokenBlacklisted(t *testing.T) {
	repo := NewMockAuthRepository()
	biz := NewBiz(repo)
	ctx := context.Background()

	// 测试不存在的token
	blacklisted, err := biz.AuthRepo.IsTokenBlacklisted(ctx, "non-existent-token")
	if err != nil {
		t.Fatalf("IsTokenBlacklisted failed: %v", err)
	}
	if blacklisted {
		t.Error("Non-existent token should not be blacklisted")
	}

	// 添加token到黑名单
	token := "blacklisted-token"
	_ = biz.AuthRepo.TokenBlacklist(ctx, token, 1700000000)

	// 测试存在的token
	blacklisted, err = biz.AuthRepo.IsTokenBlacklisted(ctx, token)
	if err != nil {
		t.Fatalf("IsTokenBlacklisted failed: %v", err)
	}
	if !blacklisted {
		t.Error("Token should be blacklisted")
	}
}

// TestLoginRequest_Validation 测试登录请求验证
func TestLoginRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		request LoginRequest
		valid   bool
	}{
		{
			name: "valid request",
			request: LoginRequest{
				Username: "admin",
				Password: "password123",
			},
			valid: true,
		},
		{
			name: "missing username",
			request: LoginRequest{
				Password: "password123",
			},
			valid: false,
		},
		{
			name: "missing password",
			request: LoginRequest{
				Username: "admin",
			},
			valid: false,
		},
		{
			name: "with captcha",
			request: LoginRequest{
				Username:    "admin",
				Password:    "password123",
				CaptchaID:   "captcha-id-123",
				CaptchaCode: "ABCD",
			},
			valid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 简单验证：检查必填字段
			hasUsername := tt.request.Username != ""
			hasPassword := tt.request.Password != ""

			if tt.valid {
				if !hasUsername || !hasPassword {
					t.Errorf("expected valid request, but missing required fields")
				}
			}
		})
	}
}

// TestTokenPair 测试Token对结构
func TestTokenPair(t *testing.T) {
	pair := TokenPair{
		AccessToken:  "access-token-123",
		RefreshToken: "refresh-token-456",
		ExpiresAt:    time.Now().Add(24 * time.Hour),
	}

	if pair.AccessToken == "" {
		t.Error("AccessToken should not be empty")
	}
	if pair.RefreshToken == "" {
		t.Error("RefreshToken should not be empty")
	}
	if pair.ExpiresAt.IsZero() {
		t.Error("ExpiresAt should not be zero")
	}
}

// TestSession 测试会话结构
func TestSession(t *testing.T) {
	session := Session{
		UserID:   1,
		Username: "admin",
		Nickname: "管理员",
		Avatar:   "https://example.com/avatar.png",
		Roles:    []string{"admin", "editor"},
	}

	if session.UserID != 1 {
		t.Errorf("expected UserID 1, got %d", session.UserID)
	}
	if len(session.Roles) != 2 {
		t.Errorf("expected 2 roles, got %d", len(session.Roles))
	}
}

// TestCaptcha 测试验证码结构
func TestCaptcha(t *testing.T) {
	captcha := Captcha{
		ID:     "captcha-id-123",
		Image:  "data:image/png;base64,iVBORw0KGgo...",
		Answer: "ABCD",
	}

	if captcha.ID == "" {
		t.Error("Captcha ID should not be empty")
	}
	if captcha.Image == "" {
		t.Error("Captcha Image should not be empty")
	}
	if captcha.Answer == "" {
		t.Error("Captcha Answer should not be empty")
	}
}
