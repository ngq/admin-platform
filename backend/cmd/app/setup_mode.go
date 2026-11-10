package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync/atomic"
	"time"

	hdlrsetup "github.com/ngq/admin-platform/app/http/handler/setup"
	"github.com/ngq/admin-platform/cmd/app/seed"
	"github.com/ngq/admin-platform/internal/datainfra"
	svcsetup "github.com/ngq/admin-platform/internal/service/setup"
	"github.com/ngq/admin-platform/migrations"
	gorp "github.com/ngq/gorp"
	datacontract "github.com/ngq/gorp/framework/contract/data"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	"github.com/ngq/gorp/framework/httpx"
	"github.com/subosito/gotenv"
	"gopkg.in/yaml.v3"
	"gorm.io/gorm"
)

var reSetupEnvPlaceholder = regexp.MustCompile(`\$\{([a-zA-Z0-9_.-]+)(?::-?([^}]*))?\}|env\(([^)]+)\)`)

type rawDatabaseYAML struct {
	Database struct {
		Driver string `yaml:"driver"`
		DSN    string `yaml:"dsn"`
	} `yaml:"database"`
}

type readinessProbeResult struct {
	needsSetup bool
	driver     string
	dsn        string
	reason     string
}

func probeDatabaseReadiness(projectRoot string) readinessProbeResult {
	if strings.TrimSpace(projectRoot) == "" {
		projectRoot = svcsetup.ResolveProjectRoot()
	}

	envPath := filepath.Join(projectRoot, ".env")
	if _, err := os.Stat(envPath); err == nil {
		_ = gotenv.OverLoad(envPath)
	}

	rawDriver, rawDSN, err := readDatabaseConfigFromYAML(filepath.Join(projectRoot, "config", "app.yaml"))
	if err != nil {
		return readinessProbeResult{
			needsSetup: true,
			driver:     "mysql",
			reason:     fmt.Sprintf("读取配置文件失败: %v", err),
		}
	}

	driver, normErr := migrations.NormalizeDriver(rawDriver)
	if normErr != nil {
		return readinessProbeResult{
			needsSetup: true,
			driver:     "mysql",
			reason:     normErr.Error(),
		}
	}

	resolvedDSN, unresolved := expandEnvPlaceholders(rawDSN)
	if unresolved || strings.TrimSpace(resolvedDSN) == "" || isPlaceholderDSN(resolvedDSN) {
		return readinessProbeResult{
			needsSetup: true,
			driver:     driver,
			reason:     "尚未配置有效的数据库连接信息，请通过初始化向导完成配置",
		}
	}

	sqlDriver, probeTarget, err := resolveProbeTarget(driver, projectRoot, resolvedDSN)
	if err != nil {
		return readinessProbeResult{
			needsSetup: true,
			driver:     driver,
			dsn:        resolvedDSN,
			reason:     err.Error(),
		}
	}

	db, err := sql.Open(sqlDriver, probeTarget)
	if err != nil {
		return readinessProbeResult{
			needsSetup: true,
			driver:     driver,
			dsn:        resolvedDSN,
			reason:     fmt.Sprintf("数据库连接配置格式有误: %v", err),
		}
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return readinessProbeResult{
			needsSetup: true,
			driver:     driver,
			dsn:        resolvedDSN,
			reason:     fmt.Sprintf("当前配置的数据库不可用（%v），请检查连接参数或初始化数据库", err),
		}
	}

	return readinessProbeResult{
		needsSetup: false,
		driver:     driver,
		dsn:        resolvedDSN,
	}
}

func resolveProbeTarget(driver, projectRoot, dsn string) (sqlDriver string, target string, err error) {
	switch driver {
	case "mysql":
		return "mysql", dsn, nil
	case "postgres":
		return "pgx", dsn, nil
	case "sqlite":
		target = dsn
		if !strings.HasPrefix(target, ":memory:") && !strings.HasPrefix(target, "file:") && !filepath.IsAbs(target) {
			target = filepath.Join(projectRoot, filepath.FromSlash(target))
		}
		if !strings.HasPrefix(target, ":memory:") && !strings.HasPrefix(target, "file:") {
			if _, statErr := os.Stat(target); os.IsNotExist(statErr) {
				return "", "", fmt.Errorf("SQLite 数据库文件 %q 尚未初始化，请通过向导完成初始化", dsn)
			}
		}
		return "sqlite3", target, nil
	default:
		return "", "", fmt.Errorf("不支持的数据库驱动 %q", driver)
	}
}

func readDatabaseConfigFromYAML(appYamlPath string) (driver, dsn string, err error) {
	raw, err := os.ReadFile(appYamlPath)
	if err != nil {
		return "", "", err
	}
	var cfg rawDatabaseYAML
	if err := yaml.Unmarshal(raw, &cfg); err != nil {
		return "", "", err
	}
	return strings.ToLower(strings.TrimSpace(cfg.Database.Driver)), strings.TrimSpace(cfg.Database.DSN), nil
}

func expandEnvPlaceholders(value string) (string, bool) {
	if value == "" {
		return "", false
	}
	unresolved := false
	out := reSetupEnvPlaceholder.ReplaceAllStringFunc(value, func(m string) string {
		sm := reSetupEnvPlaceholder.FindStringSubmatch(m)
		var key, defVal string
		var hasDef bool
		if len(sm) > 1 && sm[1] != "" {
			key = strings.TrimSpace(sm[1])
			if len(sm) > 2 && sm[2] != "" {
				defVal = sm[2]
				hasDef = true
			}
		} else if len(sm) > 3 && sm[3] != "" {
			key = strings.TrimSpace(sm[3])
		}
		if val, ok := os.LookupEnv(key); ok && val != "" {
			return val
		}
		if hasDef {
			return defVal
		}
		unresolved = true
		return ""
	})
	return out, unresolved
}

func isPlaceholderDSN(dsn string) bool {
	trimmed := strings.TrimSpace(dsn)
	return trimmed == "" ||
		strings.HasPrefix(trimmed, "user:password@") ||
		strings.HasPrefix(trimmed, "setup-placeholder@")
}

func ensureSetupModePlaceholderEnv() func() {
	type envBackup struct {
		existed bool
		value   string
	}
	keys := []string{"DATABASE_DSN", "JWT_SECRET", "SERVICE_AUTH_SECRET"}
	backups := make(map[string]envBackup, len(keys))
	defaults := map[string]string{
		"DATABASE_DSN":        "setup-placeholder@tcp(127.0.0.1:3306)/setup",
		"JWT_SECRET":          "setup-placeholder-jwt-secret",
		"SERVICE_AUTH_SECRET": "setup-placeholder-service-auth-secret",
	}

	for _, k := range keys {
		val, ok := os.LookupEnv(k)
		backups[k] = envBackup{existed: ok, value: val}
		if !ok || strings.TrimSpace(val) == "" {
			_ = os.Setenv(k, defaults[k])
		}
	}

	return func() {
		for _, k := range keys {
			cur := os.Getenv(k)
			if cur == defaults[k] {
				if backups[k].existed {
					_ = os.Setenv(k, backups[k].value)
				} else {
					_ = os.Unsetenv(k)
				}
			}
		}
	}
}

func runSetupWizardServer(projectRoot string, probe readinessProbeResult) (bool, error) {
	restoreEnv := ensureSetupModePlaceholderEnv()
	defer restoreEnv()

	setupCtx, cancelSetup := context.WithCancel(context.Background())
	defer cancelSetup()

	var completed atomic.Bool
	setupSvc := svcsetup.NewService(svcsetup.Options{
		Initialized:  false,
		ProjectRoot:  projectRoot,
		Reason:       probe.reason,
		ActiveDriver: probe.driver,
		ActiveDSN:    probe.dsn,
		SeedFn: func(_ *gorm.DB, d *datainfra.Data) error {
			return seed.Run(nil, d, d.Casbin)
		},
		OnInitialized: func() {
			completed.Store(true)
			cancelSetup()
		},
	})

	fmt.Fprintln(os.Stdout, "[setup] database not initialized; starting database setup wizard endpoints at /api/v1/setup/*")
	err := gorp.RunContext(setupCtx, setupWizardRuntimeOptions(setupSvc)...)
	return completed.Load(), err
}

func setupWizardRuntimeOptions(setupSvc *svcsetup.Service) []gorp.Option {
	return []gorp.Option{
		gorp.HTTP(gorp.HTTPServiceOptions{
			DisableGorm:  true,
			DisableRedis: true,
		}),
		// Setup 模式允许 app.yaml 中暂为空 DSN 或占位符，由 setupConfigOverrideProvider 跳过 database 必填校验。
		gorp.WithProviders(&setupConfigOverrideProvider{}),
		gorp.WithMonoGovernance(),
		gorp.WithContractHTTPMode(),
		gorp.WithSetup(func(rt *gorp.HTTPRuntime) error {
			return registerSetupWizardRoutes(rt, setupSvc)
		}),
	}
}

func registerSetupWizardRoutes(rt *gorp.HTTPRuntime, setupSvc *svcsetup.Service) error {
	hdlrsetup.RegisterRoutes(rt.Router.Group("/api/v1"), setupSvc)
	rt.Router.NoRoute(func(ctx gorp.Context) {
		ctx.AbortWithJSON(http.StatusServiceUnavailable, httpx.Response{
			Code:    httpx.CodeFailedPrecondition,
			Reason:  "SETUP_REQUIRED",
			Message: "数据库尚未初始化，请先通过 /setup 页面完成配置",
		})
	})
	return nil
}

type setupConfigOverrideProvider struct{}

func (p *setupConfigOverrideProvider) Name() string        { return "setup-config-override" }
func (p *setupConfigOverrideProvider) IsDefer() bool       { return false }
func (p *setupConfigOverrideProvider) Provides() []string  { return []string{datacontract.ConfigKey} }
func (p *setupConfigOverrideProvider) DependsOn() []string { return nil }
func (p *setupConfigOverrideProvider) Boot(runtimecontract.Container) error {
	return nil
}

func (p *setupConfigOverrideProvider) Register(c runtimecontract.Container) error {
	raw, err := c.Make(datacontract.ConfigKey)
	if err != nil {
		return err
	}
	baseCfg, ok := raw.(datacontract.Config)
	if !ok || baseCfg == nil {
		return nil
	}
	wrapped := &setupModeConfig{Config: baseCfg}
	c.Bind(datacontract.ConfigKey, func(runtimecontract.Container) (any, error) {
		return wrapped, nil
	}, true)
	return nil
}

// setupModeConfig 在 Setup Wizard 阶段包装底层 Config，使 ValidateCriticalConfig
// 不会因为 database.driver 非空而 database.dsn 为空字符串导致无法进入初始化向导。
type setupModeConfig struct {
	datacontract.Config
}

func (c *setupModeConfig) Get(key string) any {
	if key == "database" {
		return nil
	}
	return c.Config.Get(key)
}
