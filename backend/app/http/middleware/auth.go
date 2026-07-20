// Package middleware - JWT 认证中间件（带黑名单检查）
//
// 中文说明：
// - 封装框架 jwtprovider.AuthMiddleware，增加 token 黑名单检查；
// - 框架中间件只校验签名和过期时间，不感知业务侧黑名单；
// - 本中间件在框架校验通过后，额外查询黑名单，已退出的 token 立即失效。
package middleware

import (
	"strings"

	"github.com/ngq/admin-platform/internal/biz/auth"
	"github.com/ngq/admin-platform/internal/errors"
	gorp "github.com/ngq/gorp"
	securitycontract "github.com/ngq/gorp/framework/contract/security"
	jwtprovider "github.com/ngq/gorp/framework/provider/auth/jwt"
)

// AuthMiddleware 创建带黑名单检查的 JWT 认证中间件。
// 先调用框架 AuthMiddleware 校验签名/过期/subjectType，
// 再查询黑名单确保已退出的 token 无法继续使用。
//
// 参数：
//   - jwtSvc: JWT 签发/校验服务
//   - expectedSubjectType: 期望的主体类型（如 "user"）
//   - authRepo: 认证仓储，用于查询黑名单
func AuthMiddleware(jwtSvc securitycontract.JWTService, expectedSubjectType string, authRepo auth.AuthRepository) gorp.Middleware {
	// 框架中间件处理签名校验、过期校验、subjectType 校验、claims 注入
	inner := jwtprovider.AuthMiddleware(jwtSvc, expectedSubjectType)

	return func(next gorp.Handler) gorp.Handler {
		return func(ctx gorp.Context) {
			// 先执行框架中间件：校验签名、过期时间、subjectType，注入 claims
			// 框架中间件在 token 无效时会直接返回 401，不会调用 next
			// 我们用一个标记位来判断框架中间件是否通过了校验
			passed := false
			inner(func(c gorp.Context) {
				passed = true
				// 不调用 next，先检查黑名单
			})(ctx)

			if !passed {
				ctx.Abort(401)
				// 框架中间件已返回 401，无需继续
				return
			}

			// 框架校验通过，检查黑名单
			token := extractBearerToken(ctx.GetHeader("Authorization"))
			if token != "" && authRepo != nil {
				blacklisted, err := authRepo.IsTokenBlacklisted(ctx.Context(), token)
				if err != nil {
					abortWithError(ctx, errors.InternalError("令牌校验失败"))
					return
				}
				if blacklisted {
					abortWithError(ctx, errors.Unauthorized("token已失效，请重新登录"))
					return
				}
			}

			// 黑名单检查通过，执行后续 handler
			if next != nil {
				next(ctx)
			}
		}
	}
}

// extractBearerToken 从 Authorization header 提取 Bearer token
func extractBearerToken(header string) string {
	const prefix = "Bearer "
	if !strings.HasPrefix(strings.TrimSpace(header), prefix) {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(header), prefix))
}
