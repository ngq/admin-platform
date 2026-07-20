package auth

import "time"

// LoginRequest 登录请求
type LoginRequest struct {
	Username    string `json:"username" validate:"required"` // 用户名
	Password    string `json:"password" validate:"required"` // 密码
	CaptchaID   string `json:"captcha_id"`                   // 验证码ID
	CaptchaCode string `json:"captcha_code"`                 // 验证码答案
}

// TokenPair JWT令牌对
type TokenPair struct {
	AccessToken  string    `json:"access_token"`  // 访问令牌
	RefreshToken string    `json:"refresh_token"` // 刷新令牌
	ExpiresAt    time.Time `json:"expires_at"`    // 过期时间
}

// Session 会话信息
type Session struct {
	UserID   int64    `json:"user_id"`  // 用户ID
	Username string   `json:"username"` // 用户名
	Nickname string   `json:"nickname"` // 昵称
	Avatar   string   `json:"avatar"`   // 头像URL
	Roles    []string `json:"roles"`    // 角色编码列表
}

// Captcha 验证码
type Captcha struct {
	ID     string `json:"captcha_id"` // 验证码ID
	Image  string `json:"image"`      // 验证码图片（base64编码）
	Answer string `json:"-"`          // 验证码答案（不序列化）
}
