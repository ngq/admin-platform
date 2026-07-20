// Package service - 认证服务
// 处理登录、登出、令牌刷新、验证码、当前用户查询等认证相关逻辑
package iam

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/mojocn/base64Captcha"

	bizAudit "github.com/ngq/admin-platform/internal/biz/audit"
	bizAuth "github.com/ngq/admin-platform/internal/biz/auth"
	bizIam "github.com/ngq/admin-platform/internal/biz/iam"
	bizSystem "github.com/ngq/admin-platform/internal/biz/system"
	"github.com/ngq/admin-platform/internal/errors"
	securitycontract "github.com/ngq/gorp/framework/contract/security"
	"golang.org/x/crypto/bcrypt"
)

// captchaStore 验证码存储（内存实现）
// 生产环境可替换为 Redis 实现
var captchaStore = base64Captcha.DefaultMemStore

// captchaMaker 验证码生成器
// 使用默认配置：字符验证码，宽 120，高 40
var captchaMaker = base64Captcha.NewCaptcha(base64Captcha.DefaultDriverDigit, captchaStore)

// AuthService 认证服务
// 协调 JWT 服务、认证仓库、用户仓库完成认证流程
type AuthService struct {
	iamBiz    *bizIam.Biz
	authBiz   *bizAuth.Biz
	auditBiz  *bizAudit.Biz
	systemBiz *bizSystem.Biz
	jwtSvc    securitycontract.JWTService
}

// NewAuthService 创建认证服务
func NewAuthService(iamBiz *bizIam.Biz, authBiz *bizAuth.Biz, auditBiz *bizAudit.Biz, systemBiz *bizSystem.Biz, jwtSvc securitycontract.JWTService) *AuthService {
	return &AuthService{iamBiz: iamBiz, authBiz: authBiz, auditBiz: auditBiz, systemBiz: systemBiz, jwtSvc: jwtSvc}
}

// Login 登录
// 校验验证码 → 查找用户 → 校验密码(bcrypt) → 校验状态 → 签发 JWT → 记录登录日志
func (s *AuthService) Login(ctx context.Context, username, password, captchaID, captchaCode, ip string) (*bizAuth.TokenPair, error) {
	// 从系统参数读取验证码启用状态
	captchaEnabled := true // 默认启用
	if s.systemBiz.SysParamRepo != nil {
		if param, err := s.systemBiz.SysParamRepo.GetByKey(ctx, "captcha_enabled"); err == nil {
			captchaEnabled = param.Value == "true"
		}
	}

	// 校验验证码（仅在启用时校验，关闭时直接跳过）
	if captchaEnabled {
		if captchaID == "" || captchaCode == "" {
			return nil, errors.BadRequest("请输入验证码")
		}
		if !captchaStore.Verify(captchaID, captchaCode, true) {
			return nil, errors.BadRequest("验证码错误或已过期")
		}
	}

	// 根据用户名查找用户
	user, err := s.iamBiz.UserRepo.GetByUsername(ctx, username)
	if err != nil {
		// 记录登录失败日志（对外统一消息，日志内部也用统一消息避免枚举用户名）
		s.recordLoginLog(ctx, 0, username, "", "", bizAudit.LoginStatusFail, "登录失败")
		return nil, errors.UserTokenInvalid("用户名或密码错误")
	}

	// 使用 bcrypt 校验密码哈希
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		// 记录登录失败日志（对外统一消息，日志内部也用统一消息避免枚举用户名）
		s.recordLoginLog(ctx, user.ID, username, "", "", bizAudit.LoginStatusFail, "登录失败")
		return nil, errors.UserPasswordWrong("用户名或密码错误")
	}

	// 校验用户状态
	if !user.IsNormal() {
		s.recordLoginLog(ctx, user.ID, username, "", "", bizAudit.LoginStatusFail, "用户已被禁用")
		return nil, errors.UserDisabled()
	}

	// 更新最后登录时间和 IP
	_ = s.iamBiz.UserRepo.UpdateLastLogin(ctx, user.ID, ip)

	// 构造 JWT claims，签发 access token
	roles := make([]string, 0, len(user.Roles))
	for _, r := range user.Roles {
		roles = append(roles, r.Code)
	}
	accessClaims := s.jwtSvc.NewClaims(user.ID, "user", user.Username, roles, 7200) // 2 小时
	accessToken, err := s.jwtSvc.Sign(accessClaims)
	if err != nil {
		return nil, errors.InternalError("签发令牌失败")
	}

	// 签发 refresh token，7 天有效期
	refreshClaims := s.jwtSvc.NewClaims(user.ID, "refresh", user.Username, nil, 604800)
	refreshToken, err := s.jwtSvc.Sign(refreshClaims)
	if err != nil {
		return nil, errors.InternalError("签发刷新令牌失败")
	}

	// 记录登录成功日志
	s.recordLoginLog(ctx, user.ID, username, "", "", bizAudit.LoginStatusSuccess, "登录成功")

	// 将 Unix 时间戳转换为 time.Time
	expiresAt := time.Unix(accessClaims.ExpiresAt, 0)

	return &bizAuth.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
	}, nil
}

// Logout 登出
// 将 access token 加入黑名单，使其立即失效
func (s *AuthService) Logout(ctx context.Context, tokenString string) error {
	// 解析令牌获取过期时间
	claims, err := s.jwtSvc.Verify(tokenString)
	if err != nil {
		// 令牌已无效，无需加入黑名单
		return nil
	}
	return s.authBiz.AuthRepo.TokenBlacklist(ctx, tokenString, claims.ExpiresAt)
}

// GenerateCaptcha 生成验证码
// 使用 base64Captcha 库生成数字验证码，返回 ID 和 base64 图片
func (s *AuthService) GenerateCaptcha(ctx context.Context) (*bizAuth.Captcha, error) {
	id, b64s, _, err := captchaMaker.Generate()
	if err != nil {
		return nil, errors.InternalError("生成验证码失败")
	}
	return &bizAuth.Captcha{
		ID:    id,
		Image: b64s,
	}, nil
}

// VerifyCaptcha 校验验证码
// captchaID: 验证码 ID；captchaCode: 用户输入的验证码；clear: 校验后是否清除
func (s *AuthService) VerifyCaptcha(captchaID, captchaCode string, clear bool) bool {
	return captchaStore.Verify(captchaID, captchaCode, clear)
}

// GetCurrentUser 获取当前登录用户详细信息
// 根据 claims 中的用户 ID 查询用户并返回
func (s *AuthService) GetCurrentUser(ctx context.Context, userID int64) (*bizIam.User, error) {
	user, err := s.iamBiz.UserRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, errors.NotFound("用户不存在")
	}
	return user, nil
}

// RefreshToken 刷新令牌
// 校验 refresh token 有效性，签发新的 access + refresh token
func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*bizAuth.TokenPair, error) {
	// 校验 refresh token 签名
	claims, err := s.jwtSvc.Verify(refreshToken)
	if err != nil {
		return nil, errors.UserTokenInvalid("刷新令牌无效")
	}

	// 检查黑名单
	blacklisted, err := s.authBiz.AuthRepo.IsTokenBlacklisted(ctx, refreshToken)
	if err != nil {
		return nil, errors.InternalError("令牌校验失败")
	}
	if blacklisted {
		return nil, errors.UserTokenInvalid("刷新令牌已失效")
	}

	// 校验令牌类型
	if claims.SubjectType != "refresh" {
		return nil, errors.UserTokenInvalid("令牌类型错误")
	}

	// 查询用户信息（含角色）
	user, err := s.iamBiz.UserRepo.GetByID(ctx, claims.SubjectID)
	if err != nil {
		return nil, errors.NotFound("用户不存在")
	}

	// 签发新 access token
	roles := make([]string, 0, len(user.Roles))
	for _, r := range user.Roles {
		roles = append(roles, r.Code)
	}
	newAccessClaims := s.jwtSvc.NewClaims(user.ID, "user", user.Username, roles, 7200)
	newAccessToken, err := s.jwtSvc.Sign(newAccessClaims)
	if err != nil {
		return nil, errors.InternalError("签发令牌失败")
	}

	// 签发新 refresh token
	newRefreshClaims := s.jwtSvc.NewClaims(user.ID, "refresh", user.Username, nil, 604800)
	newRefreshToken, err := s.jwtSvc.Sign(newRefreshClaims)
	if err != nil {
		return nil, errors.InternalError("签发刷新令牌失败")
	}

	// 将旧 refresh token 加入黑名单
	_ = s.authBiz.AuthRepo.TokenBlacklist(ctx, refreshToken, claims.ExpiresAt)

	// 将 Unix 时间戳转换为 time.Time
	expiresAt := time.Unix(newAccessClaims.ExpiresAt, 0)

	return &bizAuth.TokenPair{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
		ExpiresAt:    expiresAt,
	}, nil
}

// ChangePassword 修改当前用户密码
// 校验旧密码 → 校验新密码复杂度 → 设置新密码
// userID 从 JWT claims 获取，不依赖 URL 参数，消除 IDOR 风险
func (s *AuthService) ChangePassword(ctx context.Context, userID int64, oldPassword, newPassword string) error {
	// 获取用户当前密码哈希
	user, err := s.iamBiz.UserRepo.GetByID(ctx, userID)
	if err != nil {
		return errors.UserNotFound()
	}

	// 校验旧密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword)); err != nil {
		return errors.UserPasswordWrong("旧密码错误")
	}

	// 校验新密码复杂度（复用 UserService.validatePassword 的逻辑）
	if err := validatePasswordComplexity(ctx, s.systemBiz.SysParamRepo, newPassword); err != nil {
		return err
	}

	// 哈希化新密码
	hashedPwd, err := HashPassword(newPassword)
	if err != nil {
		return errors.InternalError("密码哈希失败")
	}

	return s.iamBiz.UserRepo.SetPassword(ctx, userID, hashedPwd)
}

// GetUserPermissionCodes 获取用户权限码列表
// 根据用户 ID 查询其关联的角色，再查询角色关联的按钮权限码，合并去重后返回
// 用于前端按钮级权限控制
func (s *AuthService) GetUserPermissionCodes(ctx context.Context, userID int64) ([]string, error) {
	// 1. 查询用户关联的角色
	roles, err := s.iamBiz.RoleRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, errors.InternalError("查询用户角色失败")
	}
	if len(roles) == 0 {
		return []string{}, nil
	}

	// 2. 提取角色 ID 列表
	roleIDs := make([]int64, 0, len(roles))
	for _, r := range roles {
		roleIDs = append(roleIDs, r.ID)
	}

	// 3. 查询角色关联的按钮权限
	buttons, err := s.iamBiz.ButtonRepo.GetByRoleIDs(ctx, roleIDs)
	if err != nil {
		return nil, errors.InternalError("查询按钮权限失败")
	}

	// 4. 提取权限码并去重
	codeSet := make(map[string]struct{})
	for _, b := range buttons {
		if b.Code != "" {
			codeSet[b.Code] = struct{}{}
		}
	}

	// 5. 转换为切片返回
	codes := make([]string, 0, len(codeSet))
	for code := range codeSet {
		codes = append(codes, code)
	}
	return codes, nil
}

// HashPassword 使用 bcrypt 对明文密码进行哈希
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// validatePasswordComplexity 校验密码复杂度（独立函数，供 AuthService 和 CustomerService 复用）
func validatePasswordComplexity(ctx context.Context, sysParamRepo bizSystem.SysParamRepository, password string) error {
	minLength := 6
	if sysParamRepo != nil {
		if param, err := sysParamRepo.GetByKey(ctx, "password_min_length"); err == nil {
			if length, err := strconv.Atoi(param.Value); err == nil && length > 0 {
				minLength = length
			}
		}
	}

	if len(password) < minLength {
		return errors.BadRequest(fmt.Sprintf("密码长度不能少于 %d 位", minLength))
	}

	hasLetter := false
	hasDigit := false
	for _, c := range password {
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') {
			hasLetter = true
		}
		if c >= '0' && c <= '9' {
			hasDigit = true
		}
	}
	if !hasLetter || !hasDigit {
		return errors.BadRequest("密码必须同时包含字母和数字")
	}

	return nil
}

// recordLoginLog 记录登录日志（异步，不阻塞主流程）
// 如果 LoginLogRepo 为 nil 则跳过
func (s *AuthService) recordLoginLog(ctx context.Context, userID int64, username, ip, userAgent string, status int8, message string) {
	if s.auditBiz.LoginLogRepo == nil {
		return
	}
	log := &bizAudit.LoginLog{
		UserID:    userID,
		Username:  username,
		IP:        ip,
		UserAgent: userAgent,
		Status:    status,
		Message:   message,
	}
	// 忽略错误，登录日志不应阻塞认证流程
	_ = s.auditBiz.LoginLogRepo.Create(ctx, log)
}
