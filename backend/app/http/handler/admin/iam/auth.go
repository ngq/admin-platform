// Package handler - 认证处理器
// 处理登录、登出、令牌刷新、验证码、当前用户查询等 HTTP 请求
package iam

import (
	"github.com/ngq/admin-platform/internal/errors"
	svciam "github.com/ngq/admin-platform/internal/service/iam"
	"github.com/ngq/gorp"
	"github.com/ngq/gorp/framework/httpx"
)

// LoginRequest 登录请求
type LoginRequest struct {
	Username    string `json:"username" binding:"required"`
	Password    string `json:"password" binding:"required"`
	CaptchaID   string `json:"captchaId"`
	CaptchaCode string `json:"captchaCode"`
}

// RefreshTokenRequest 刷新令牌请求
type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

// AuthHandler 认证处理器
type AuthHandler struct {
	svc *svciam.AuthService
}

// NewAuthHandler 创建认证处理器
func NewAuthHandler(svc *svciam.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

// Login 登录
// POST /api/v1/auth/login
func (h *AuthHandler) Login(ctx gorp.Context) {
	var req LoginRequest
	if err := ctx.Bind(&req); err != nil {
		httpx.Error(ctx, errors.BadRequest("请求参数错误"))
		return
	}

	// 调用认证服务登录（验证码校验在 service 内完成）
	ip := ctx.GetHeader("X-Real-IP")
	if ip == "" {
		ip = ctx.GetHeader("X-Forwarded-For")
	}
	tokenPair, err := h.svc.Login(ctx.Context(), req.Username, req.Password, req.CaptchaID, req.CaptchaCode, ip)
	if err != nil {
		httpx.Error(ctx, err)
		return
	}

	httpx.Success(ctx, tokenPair)
}

// Logout 登出
// POST /api/v1/auth/logout
func (h *AuthHandler) Logout(ctx gorp.Context) {
	token := extractToken(ctx)
	if token == "" {
		httpx.Error(ctx, errors.BadRequest("缺少认证令牌"))
		return
	}

	if err := h.svc.Logout(ctx.Context(), token); err != nil {
		httpx.Error(ctx, err)
		return
	}

	httpx.Success(ctx, nil)
}

// RefreshToken 刷新令牌
// POST /api/v1/auth/refresh
func (h *AuthHandler) RefreshToken(ctx gorp.Context) {
	var req RefreshTokenRequest
	if err := ctx.Bind(&req); err != nil {
		httpx.Error(ctx, errors.BadRequest("请求参数错误"))
		return
	}

	tokenPair, err := h.svc.RefreshToken(ctx.Context(), req.RefreshToken)
	if err != nil {
		httpx.Error(ctx, err)
		return
	}

	httpx.Success(ctx, tokenPair)
}

// Captcha 获取验证码
// GET /api/v1/auth/captcha
func (h *AuthHandler) Captcha(ctx gorp.Context) {
	captcha, err := h.svc.GenerateCaptcha(ctx.Context())
	if err != nil {
		httpx.Error(ctx, err)
		return
	}

	httpx.Success(ctx, captcha)
}

// GetCurrentUser 获取当前登录用户
// GET /api/v1/auth/current
func (h *AuthHandler) GetCurrentUser(ctx gorp.Context) {
	// 从 JWT claims 中获取用户 ID
	jwtClaims, ok := gorp.ClaimsFromContext(ctx)
	if !ok {
		httpx.Error(ctx, errors.BadRequest("未获取到认证信息"))
		return
	}

	user, err := h.svc.GetCurrentUser(ctx.Context(), jwtClaims.SubjectID)
	if err != nil {
		httpx.Error(ctx, err)
		return
	}

	httpx.Success(ctx, user)
}

// AuthChangePasswordRequest 修改密码请求
type AuthChangePasswordRequest struct {
	OldPassword string `json:"oldPassword" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required,min=6"`
}

// ChangePassword 修改当前用户密码
// POST /api/v1/auth/password
// 用户 ID 从 JWT claims 获取，不从 URL 参数取，彻底消除 IDOR
func (h *AuthHandler) ChangePassword(ctx gorp.Context) {
	claims, ok := gorp.ClaimsFromContext(ctx)
	if !ok || claims == nil {
		httpx.Error(ctx, errors.Unauthorized("未认证"))
		return
	}

	var req AuthChangePasswordRequest
	if err := ctx.Bind(&req); err != nil {
		httpx.Error(ctx, errors.BadRequest("请求参数错误"))
		return
	}

	if err := h.svc.ChangePassword(ctx.Context(), claims.SubjectID, req.OldPassword, req.NewPassword); err != nil {
		httpx.Error(ctx, err)
		return
	}

	httpx.Success(ctx, nil)
}

// GetAccessCodes 获取当前用户权限码
// GET /api/v1/auth/codes
// 返回用户所有按钮级权限码，用于前端权限控制
func (h *AuthHandler) GetAccessCodes(ctx gorp.Context) {
	claims, ok := gorp.ClaimsFromContext(ctx)
	if !ok || claims == nil {
		httpx.Error(ctx, errors.Unauthorized("未认证"))
		return
	}

	codes, err := h.svc.GetUserPermissionCodes(ctx.Context(), claims.SubjectID)
	if err != nil {
		httpx.Error(ctx, err)
		return
	}

	httpx.Success(ctx, codes)
}

// extractToken 从请求头提取 Bearer token
func extractToken(ctx gorp.Context) string {
	authHeader := ctx.GetHeader("Authorization")
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		return authHeader[7:]
	}
	return ""
}
