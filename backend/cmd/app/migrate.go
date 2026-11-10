package main

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/ngq/admin-platform/internal/data/po"
	"github.com/ngq/admin-platform/migrations"
	gorp "github.com/ngq/gorp"
	gorpMigrate "github.com/ngq/gorp/framework/pkg/migrate"
)

const migrateTimeout = 30 * time.Second

func migrate(rt *gorp.HTTPRuntime) error {
	return migrateWithFS(rt, migrations.FS, "")
}

func migrateWithFS(rt *gorp.HTTPRuntime, fsys fs.FS, dirOverride string) error {
	if rt == nil || rt.DB == nil {
		return nil
	}

	env, driver, mode := resolveMigrateSettings(rt)

	ctx, cancel := context.WithTimeout(context.Background(), migrateTimeout)
	defer cancel()

	// 生产环境或显式指定 goose 模式：严格使用版本化 SQL 迁移，禁止 AutoMigrate
	if env == "production" || mode == "goose" {
		return runGooseUp(ctx, rt, driver, fsys, dirOverride)
	}

	// 开发/测试环境：默认 hybrid（按当前驱动执行对应方言的 goose 脚本；对 mysql/sqlite 补充 AutoMigrate 方便开发迭代）
	if mode != "auto" {
		if err := runGooseUp(ctx, rt, driver, fsys, dirOverride); err != nil {
			return err
		}
	}

	// PostgreSQL 下部分 PO 含 MySQL 专属 tag（如 tinyint/bigint unsigned），由 goose 迁移文件统一维护表结构
	if driver == "postgres" {
		return nil
	}

	return autoMigrateModels(rt)
}

func resolveMigrateSettings(rt *gorp.HTTPRuntime) (env, driver, mode string) {
	if rt.DB != nil && rt.DB.Dialector != nil {
		driver = strings.ToLower(strings.TrimSpace(rt.DB.Dialector.Name()))
	}
	if rt.Config != nil {
		env = strings.ToLower(strings.TrimSpace(rt.Config.GetString("app.env")))
		if cfgDriver := strings.ToLower(strings.TrimSpace(rt.Config.GetString("database.driver"))); cfgDriver != "" {
			driver = cfgDriver
		}
		mode = strings.ToLower(strings.TrimSpace(rt.Config.GetString("database.migrate_mode")))
	}
	if canonical, err := migrations.NormalizeDriver(driver); err == nil {
		driver = canonical
	}
	return env, driver, mode
}

func runGooseUp(ctx context.Context, rt *gorp.HTTPRuntime, driver string, fsys fs.FS, dirOverride string) error {
	canonicalDriver, err := migrations.NormalizeDriver(driver)
	if err != nil {
		return err
	}
	dir := strings.TrimSpace(dirOverride)
	if dir == "" {
		dir, err = migrations.DirForDriver(canonicalDriver)
		if err != nil {
			return err
		}
	}
	if fsys == nil {
		fsys = migrations.FS
	}
	sqlDB, err := rt.DB.DB()
	if err != nil {
		return fmt.Errorf("unwrap sql.DB for goose migrate: %w", err)
	}
	if err := gorpMigrate.RunUpFS(ctx, sqlDB, canonicalDriver, fsys, dir); err != nil {
		return fmt.Errorf("goose migrate up (%s): %w", canonicalDriver, err)
	}
	return nil
}

func autoMigrateModels(rt *gorp.HTTPRuntime) error {
	rt.DB.Config.DisableForeignKeyConstraintWhenMigrating = true
	return rt.DB.AutoMigrate(
		&po.AdminUserPO{},
		&po.DepartmentPO{},
		&po.RolePO{},
		&po.UserRolePO{},
		&po.MenuPO{},
		&po.ButtonPO{},
		&po.APIPO{},
		&po.RoleMenuPO{},
		&po.RoleAPIPO{},
		&po.RoleButtonPO{},
		&po.SysParamPO{},
		&po.OperationLogPO{},
		&po.LoginLogPO{},
		&po.ErrorLogPO{},
	)
}

func maybeHandleMigrateCLI(args []string) (bool, error) {
	if len(args) == 0 || args[0] != "migrate" {
		return false, nil
	}
	if len(args) < 2 {
		return true, fmt.Errorf("usage: app migrate <up|down|status|version>")
	}
	return true, runMigrateCommand(args[1], os.Stdout)
}

func runMigrateCommand(action string, out io.Writer) error {
	rt, err := gorp.BuildHTTPRuntime(gorp.WithMonoGovernance(), gorp.WithContractHTTPMode())
	if err != nil {
		return err
	}
	if rt.Container != nil {
		defer func() { _ = rt.Container.Destroy() }()
	}
	if rt.DB == nil {
		return fmt.Errorf("database runtime is nil; ensure database config and DSN are set")
	}

	_, driver, _ := resolveMigrateSettings(rt)
	dir, err := migrations.DirForDriver(driver)
	if err != nil {
		return err
	}
	sqlDB, err := rt.DB.DB()
	if err != nil {
		return fmt.Errorf("unwrap sql.DB: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), migrateTimeout)
	defer cancel()

	switch strings.ToLower(strings.TrimSpace(action)) {
	case "up":
		if err := gorpMigrate.RunUpFS(ctx, sqlDB, driver, migrations.FS, dir); err != nil {
			return err
		}
		_, _ = fmt.Fprintf(out, "migrations applied (%s)\n", driver)
		return nil
	case "down":
		if err := gorpMigrate.RunDownFS(ctx, sqlDB, driver, migrations.FS, dir); err != nil {
			return err
		}
		_, _ = fmt.Fprintf(out, "latest migration rolled back (%s)\n", driver)
		return nil
	case "status":
		statuses, err := gorpMigrate.GetStatus(ctx, sqlDB, driver, migrations.FS, dir)
		if err != nil {
			return err
		}
		w := tabwriter.NewWriter(out, 0, 0, 3, ' ', 0)
		_, _ = fmt.Fprintln(w, "VERSION\tSTATUS\tAPPLIED AT\tSOURCE")
		_, _ = fmt.Fprintln(w, "-------\t------\t----------\t------")
		for _, s := range statuses {
			appliedAt := "-"
			if !s.AppliedAt.IsZero() {
				appliedAt = s.AppliedAt.Format("2006-01-02 15:04:05")
			}
			version := int64(0)
			sourceName := "-"
			if s.Source != nil {
				version = s.Source.Version
				sourceName = filepath.Base(s.Source.Path)
			}
			_, _ = fmt.Fprintf(w, "%d\t%s\t%s\t%s\n", version, s.State, appliedAt, sourceName)
		}
		return w.Flush()
	case "version":
		ver, err := gorpMigrate.GetDBVersion(ctx, sqlDB, driver)
		if err != nil {
			return err
		}
		_, _ = fmt.Fprintf(out, "version: %d (%s)\n", ver, driver)
		return nil
	default:
		return fmt.Errorf("unknown migrate action %q (supported: up, down, status, version)", action)
	}
}
