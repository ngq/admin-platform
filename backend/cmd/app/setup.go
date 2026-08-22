package main

import (
	"context"
	"fmt"
	"sync"

	apphttp "github.com/ngq/admin-platform/app/http"
	"github.com/ngq/admin-platform/cmd/app/routesync"
	"github.com/ngq/admin-platform/cmd/app/seed"
	dataaudit "github.com/ngq/admin-platform/internal/data/audit"
	dataauth "github.com/ngq/admin-platform/internal/data/auth"
	dataiam "github.com/ngq/admin-platform/internal/data/iam"
	datasystem "github.com/ngq/admin-platform/internal/data/system"
	"github.com/ngq/admin-platform/internal/datainfra"
	svciam "github.com/ngq/admin-platform/internal/service/iam"
	gorp "github.com/ngq/gorp"
)

func setup(rt *gorp.HTTPRuntime) error {
	casbinModelPath := ""
	if rt.Config != nil {
		casbinModelPath = rt.Config.GetString("casbin.model_path")
	}
	d := datainfra.NewData(rt.DB, casbinModelPath)
	systemBiz := datasystem.NewBiz(d, datasystem.NewDashboardRepo(d))
	iamBiz := dataiam.NewBiz(d, systemBiz.SysParamRepo)
	authBiz := dataauth.NewBiz()
	auditBiz := dataaudit.NewBiz(d)

	// 使用基于框架 Redis 契约的验证码存储（支持多实例共享，无 Redis 时自动安全降级为内存存储）
	captchaStore := dataauth.NewRedisCaptchaStore(rt.Redis)
	authSvc := svciam.NewAuthService(iamBiz, authBiz, auditBiz, systemBiz, rt.JWT, captchaStore)
	roleSvc := svciam.NewRoleService(iamBiz, d.Casbin)
	routeRegistry := routesync.NewRegistry()

	// 初始化 API 权限模式内存高速缓存（消除中间件每次 HTTP 请求直连 MySQL 的开销）
	var permissionCache sync.Map
	reloadPermissionCache := func(ctx context.Context) {
		var rows []struct {
			Path           string `gorm:"column:path"`
			Method         string `gorm:"column:method"`
			PermissionMode string `gorm:"column:permission_mode"`
			Ignored        int8   `gorm:"column:ignored"`
		}
		if err := rt.DB.WithContext(ctx).Table("sys_apis").Select("path, method, permission_mode, ignored").Where("deleted_at IS NULL").Find(&rows).Error; err == nil {
			for _, row := range rows {
				mode := "casbin"
				if row.Ignored == 1 {
					mode = "ignored"
				} else if row.PermissionMode != "" {
					mode = row.PermissionMode
				}
				permissionCache.Store(row.Method+":"+row.Path, mode)
			}
		}
	}

	resolveMode := func(ctx context.Context, path, method string) (string, bool) {
		key := method + ":" + path
		if val, ok := permissionCache.Load(key); ok {
			return val.(string), true
		}
		// 缓存未命中时安全回退查库并写入缓存
		var row struct {
			PermissionMode string
			Ignored        int8
		}
		err := rt.DB.WithContext(ctx).Table("sys_apis").Select("permission_mode, ignored").Where("path = ? AND method = ? AND deleted_at IS NULL", path, method).Take(&row).Error
		if err != nil {
			return "", false
		}
		mode := "casbin"
		if row.Ignored == 1 {
			mode = "ignored"
		} else if row.PermissionMode != "" {
			mode = row.PermissionMode
		}
		permissionCache.Store(key, mode)
		return mode, true
	}

	apphttp.RegisterRoutes(routeRegistry.Wrap(rt.Router), authSvc, roleSvc, iamBiz, systemBiz, auditBiz, rt.JWT, d.Casbin, authBiz.AuthRepo, resolveMode)
	if err := routesync.Sync(routeRegistry.Routes(), d); err != nil {
		return fmt.Errorf("route sync: %w", err)
	}
	if err := seed.Run(rt, d, d.Casbin); err != nil {
		return fmt.Errorf("seed: %w", err)
	}
	// Seed creates the super-admin role. Run route sync again so that role is
	// linked to every protected API before Casbin policies are rebuilt.
	if err := routesync.Sync(routeRegistry.Routes(), d); err != nil {
		return fmt.Errorf("post-seed route sync: %w", err)
	}
	if err := roleSvc.SyncAllCasbinPolicies(context.Background()); err != nil {
		return fmt.Errorf("casbin sync: %w", err)
	}

	// 预热加载全部已同步的 API 权限规则到内存缓存
	reloadPermissionCache(context.Background())
	return nil
}
