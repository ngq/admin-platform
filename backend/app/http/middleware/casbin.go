package middleware

import (
	"context"

	"github.com/casbin/casbin/v2"
	"github.com/ngq/admin-platform/internal/errors"
	"github.com/ngq/gorp"
)

type APIPermissionModeResolver func(ctx context.Context, path, method string) (string, bool)

func CasbinRBAC(enforcer *casbin.SyncedEnforcer) gorp.Middleware {
	return CasbinRBACWithPermissionMode(enforcer, nil)
}

func CasbinRBACWithPermissionMode(enforcer *casbin.SyncedEnforcer, resolveMode APIPermissionModeResolver) gorp.Middleware {
	return func(next gorp.Handler) gorp.Handler {
		return func(ctx gorp.Context) {
			claims, ok := gorp.ClaimsFromContext(ctx)
			if !ok || claims == nil {
				abortWithError(ctx, errors.Unauthorized("未认证，无法进行权限校验"))
				return
			}

			path := ctx.RoutePath()
			method := ""
			if ctx.Request() != nil {
				method = ctx.Request().Method
			}
			if path == "" {
				abortWithError(ctx, errors.InternalError("无法获取路由路径"))
				return
			}

			if resolveMode != nil {
				mode, ok := resolveMode(ctx.Context(), path, method)
				if ok && (mode == "auth" || mode == "ignored") {
					if next != nil {
						next(ctx)
					}
					return
				}
			}

			allowed := false
			for _, role := range claims.Roles {
				ok, err := enforcer.Enforce(role, path, method)
				if err != nil {
					abortWithError(ctx, errors.InternalError("权限校验内部错误"))
					return
				}
				if ok {
					allowed = true
					break
				}
			}

			if !allowed {
				abortWithError(ctx, errors.Forbidden("无权限访问该资源"))
				return
			}

			if next != nil {
				next(ctx)
			}
		}
	}
}
