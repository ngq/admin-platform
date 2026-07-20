package main

import (
	"context"
	"fmt"

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
	d := datainfra.NewData(rt.DB)
	systemBiz := datasystem.NewBiz(d, datasystem.NewDashboardRepo(d))
	iamBiz := dataiam.NewBiz(d, systemBiz.SysParamRepo)
	authBiz := dataauth.NewBiz()
	auditBiz := dataaudit.NewBiz(d)
	authSvc := svciam.NewAuthService(iamBiz, authBiz, auditBiz, systemBiz, rt.JWT)
	roleSvc := svciam.NewRoleService(iamBiz, d.Casbin)
	routeRegistry := routesync.NewRegistry()
	resolveMode := func(ctx context.Context, path, method string) (string, bool) {
		var row struct {
			PermissionMode string
			Ignored        int8
		}
		err := rt.DB.WithContext(ctx).Table("sys_apis").Select("permission_mode, ignored").Where("path = ? AND method = ? AND deleted_at IS NULL", path, method).Take(&row).Error
		if err != nil {
			return "", false
		}
		if row.Ignored == 1 {
			return "ignored", true
		}
		if row.PermissionMode == "" {
			return "casbin", true
		}
		return row.PermissionMode, true
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
	return nil
}
