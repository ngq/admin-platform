// Package setup provides database connection testing, initialization,
// goose migration execution, seed orchestration, and local config file persistence
// for the web-based database initialization wizard across MySQL/MariaDB, PostgreSQL, and SQLite.
package setup

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/ngq/admin-platform/internal/datainfra"
	"github.com/ngq/admin-platform/internal/errors"
	"github.com/ngq/admin-platform/migrations"
	gorp "github.com/ngq/gorp"
	gorpMigrate "github.com/ngq/gorp/framework/pkg/migrate"
	"github.com/subosito/gotenv"
	gormmysql "gorm.io/driver/mysql"
	gormpostgres "gorm.io/driver/postgres"
	gormsqlite "gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var validDBNamePattern = regexp.MustCompile(`^[a-zA-Z0-9_]{1,64}$`)

const (
	defaultDriver = "mysql"
)

// SeedFunc initializes default records and Casbin policies on a freshly migrated database.
type SeedFunc func(db *gorm.DB, d *datainfra.Data) error

// DatabaseInitInput represents database connection parameters submitted from the web setup wizard.
type DatabaseInitInput struct {
	Driver   string `json:"driver"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	Database string `json:"database" binding:"required"`
}

// Status describes whether the backend database has been configured and initialized.
type Status struct {
	Initialized bool   `json:"initialized"`
	Driver      string `json:"driver"`
	Host        string `json:"host,omitempty"`
	Port        int    `json:"port,omitempty"`
	Database    string `json:"database,omitempty"`
	ConfigFile  string `json:"config_file,omitempty"`
	Reason      string `json:"reason,omitempty"`
}

// TestConnectionResult reports database server connectivity and schema existence.
type TestConnectionResult struct {
	Connected      bool   `json:"connected"`
	DatabaseExists bool   `json:"database_exists"`
	ServerVersion  string `json:"server_version,omitempty"`
	Message        string `json:"message"`
}

// InitResult reports the outcome of database initialization and config persistence.
type InitResult struct {
	Initialized     bool   `json:"initialized"`
	ConfigFile      string `json:"config_file"`
	Driver          string `json:"driver"`
	Host            string `json:"host,omitempty"`
	Port            int    `json:"port,omitempty"`
	Database        string `json:"database"`
	DefaultUsername string `json:"default_username"`
	Message         string `json:"message"`
}

// Options configures the setup service.
type Options struct {
	Initialized   bool
	ProjectRoot   string
	Reason        string
	ActiveDSN     string
	ActiveDriver  string
	SeedFn        SeedFunc
	OnInitialized func()
}

// Service manages setup status inspection, connection probing, and database bootstrap.
type Service struct {
	mu            sync.Mutex
	initialized   atomic.Bool
	projectRoot   string
	reason        string
	activeDSN     string
	activeDriver  string
	seedFn        SeedFunc
	onInitialized func()
}

// NewService creates a setup Service instance.
func NewService(opts Options) *Service {
	root := strings.TrimSpace(opts.ProjectRoot)
	if root == "" {
		root = ResolveProjectRoot()
	}
	driver := strings.TrimSpace(strings.ToLower(opts.ActiveDriver))
	if driver == "" {
		driver = defaultDriver
	} else if canonical, err := migrations.NormalizeDriver(driver); err == nil {
		driver = canonical
	}
	s := &Service{
		projectRoot:   root,
		reason:        opts.Reason,
		activeDSN:     opts.ActiveDSN,
		activeDriver:  driver,
		seedFn:        opts.SeedFn,
		onInitialized: opts.OnInitialized,
	}
	s.initialized.Store(opts.Initialized)
	return s
}

// ResolveProjectRoot locates the backend project root directory containing config/app.yaml.
func ResolveProjectRoot() string {
	appYamlPath := gorp.ResolveResourcePath("config/app.yaml")
	if appYamlPath != "" {
		return filepath.Dir(filepath.Dir(appYamlPath))
	}
	if wd, err := os.Getwd(); err == nil {
		return wd
	}
	return "."
}

// IsInitialized returns whether the backend is already initialized.
func (s *Service) IsInitialized() bool {
	return s.initialized.Load()
}

// GetStatus returns sanitized initialization status without exposing credentials.
func (s *Service) GetStatus() *Status {
	s.mu.Lock()
	dsn := s.activeDSN
	driver := s.activeDriver
	reason := s.reason
	s.mu.Unlock()

	if s.initialized.Load() {
		host, port, dbName := ParseDSNInfo(driver, dsn)
		return &Status{
			Initialized: true,
			Driver:      driver,
			Host:        host,
			Port:        port,
			Database:    dbName,
			ConfigFile:  "config/app.yaml",
		}
	}

	return &Status{
		Initialized: false,
		Driver:      driver,
		ConfigFile:  "config/app.yaml",
		Reason:      reason,
	}
}

// TestConnection verifies database connectivity and checks whether the target database exists.
func (s *Service) TestConnection(ctx context.Context, input DatabaseInitInput, allowWhenInitialized ...bool) (*TestConnectionResult, error) {
	allowReinit := len(allowWhenInitialized) > 0 && allowWhenInitialized[0]
	if s.initialized.Load() && !allowReinit {
		return nil, errors.Forbidden("系统已完成初始化，禁止在未授权状态下修改数据库配置")
	}

	normalized, err := validateInput(input)
	if err != nil {
		return nil, err
	}

	probeCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	switch normalized.Driver {
	case "mysql":
		return s.testMySQLConnection(probeCtx, normalized)
	case "postgres":
		return s.testPostgresConnection(probeCtx, normalized)
	case "sqlite":
		return s.testSQLiteConnection(probeCtx, normalized)
	default:
		return nil, errors.BadRequest("不支持的数据库类型")
	}
}

func (s *Service) testMySQLConnection(ctx context.Context, in DatabaseInitInput) (*TestConnectionResult, error) {
	serverDSN := BuildMySQLDSN(in, "", 5*time.Second)
	db, err := sql.Open("mysql", serverDSN)
	if err != nil {
		return nil, errors.BadRequest(fmt.Sprintf("创建 MySQL 连接失败: %v", err))
	}
	defer db.Close()

	if err := db.PingContext(ctx); err != nil {
		return nil, errors.BadRequest(fmt.Sprintf("连接 MySQL 服务器失败: %v", err))
	}

	var version string
	_ = db.QueryRowContext(ctx, "SELECT VERSION()").Scan(&version)

	var schemaName string
	err = db.QueryRowContext(ctx, "SELECT SCHEMA_NAME FROM INFORMATION_SCHEMA.SCHEMATA WHERE SCHEMA_NAME = ?", in.Database).Scan(&schemaName)
	dbExists := err == nil && schemaName == in.Database

	msg := "MySQL 连接成功，目标数据库已存在（将执行幂等迁移）"
	if !dbExists {
		msg = fmt.Sprintf("MySQL 服务器连接成功，数据库 %q 尚不存在，初始化时将自动创建", in.Database)
	}
	return &TestConnectionResult{
		Connected:      true,
		DatabaseExists: dbExists,
		ServerVersion:  version,
		Message:        msg,
	}, nil
}

func (s *Service) testPostgresConnection(ctx context.Context, in DatabaseInitInput) (*TestConnectionResult, error) {
	adminDSN := BuildPostgresDSN(in, "postgres", 5*time.Second)
	db, err := sql.Open("pgx", adminDSN)
	if err != nil {
		return nil, errors.BadRequest(fmt.Sprintf("创建 PostgreSQL 连接失败: %v", err))
	}
	defer db.Close()

	if err := db.PingContext(ctx); err != nil {
		return nil, errors.BadRequest(fmt.Sprintf("连接 PostgreSQL 服务器失败: %v", err))
	}

	var version string
	_ = db.QueryRowContext(ctx, "SELECT version()").Scan(&version)

	var exists int
	err = db.QueryRowContext(ctx, "SELECT 1 FROM pg_database WHERE datname = $1", in.Database).Scan(&exists)
	dbExists := err == nil && exists == 1

	msg := "PostgreSQL 连接成功，目标数据库已存在（将执行幂等迁移）"
	if !dbExists {
		msg = fmt.Sprintf("PostgreSQL 服务器连接成功，数据库 %q 尚不存在，初始化时将自动创建", in.Database)
	}
	return &TestConnectionResult{
		Connected:      true,
		DatabaseExists: dbExists,
		ServerVersion:  version,
		Message:        msg,
	}, nil
}

func (s *Service) testSQLiteConnection(ctx context.Context, in DatabaseInitInput) (*TestConnectionResult, error) {
	relDSN, absPath := resolveSQLitePaths(s.projectRoot, in.Database)
	info, statErr := os.Stat(absPath)
	dbExists := statErr == nil && !info.IsDir()

	// 验证目标目录可写入且 SQLite 驱动可正常执行查询
	probeTarget := ":memory:"
	if dbExists {
		probeTarget = absPath
	} else {
		parent := filepath.Dir(absPath)
		if err := os.MkdirAll(parent, 0o755); err != nil {
			return nil, errors.BadRequest(fmt.Sprintf("SQLite 目录不可用 (%s): %v", parent, err))
		}
	}

	db, err := sql.Open("sqlite3", probeTarget)
	if err != nil {
		return nil, errors.BadRequest(fmt.Sprintf("打开 SQLite 失败: %v", err))
	}
	defer db.Close()

	var version string
	if err := db.QueryRowContext(ctx, "SELECT sqlite_version()").Scan(&version); err != nil {
		return nil, errors.BadRequest(fmt.Sprintf("SQLite 引擎探测失败: %v", err))
	}

	msg := fmt.Sprintf("SQLite 文件已存在（%s），将执行幂等迁移", relDSN)
	if !dbExists {
		msg = fmt.Sprintf("SQLite 路径校验通过（%s），初始化时将自动创建数据库文件", relDSN)
	}
	return &TestConnectionResult{
		Connected:      true,
		DatabaseExists: dbExists,
		ServerVersion:  "SQLite " + version,
		Message:        msg,
	}, nil
}

// InitializeDatabase creates the database if missing, runs goose migrations, seeds initial data,
// persists the DSN and driver to configuration files, and triggers runtime transition.
func (s *Service) InitializeDatabase(ctx context.Context, input DatabaseInitInput, allowWhenInitialized ...bool) (*InitResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	allowReinit := len(allowWhenInitialized) > 0 && allowWhenInitialized[0]
	if s.initialized.Load() && !allowReinit {
		return nil, errors.Forbidden("系统已完成初始化，禁止在未授权状态下重复初始化数据库")
	}

	normalized, err := validateInput(input)
	if err != nil {
		return nil, err
	}

	initCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	gormDB, runtimeDSN, displayDB, err := s.openAndPrepareDatabase(initCtx, normalized)
	if err != nil {
		return nil, err
	}
	targetSQLDB, err := gormDB.DB()
	if err != nil {
		return nil, errors.InternalError(fmt.Sprintf("获取底层 SQL 连接失败: %v", err))
	}
	defer func() { _ = targetSQLDB.Close() }()

	migDir, err := migrations.DirForDriver(normalized.Driver)
	if err != nil {
		return nil, errors.BadRequest(err.Error())
	}
	if err := gorpMigrate.RunUpFS(initCtx, targetSQLDB, normalized.Driver, migrations.FS, migDir); err != nil {
		return nil, errors.InternalError(fmt.Sprintf("执行 goose 数据库迁移失败: %v", err))
	}

	casbinModelPath := filepath.Join(s.projectRoot, "config", "casbin_model.conf")
	dataInfra := datainfra.NewData(gormDB, casbinModelPath)
	if s.seedFn != nil {
		if err := s.seedFn(gormDB, dataInfra); err != nil {
			return nil, errors.InternalError(fmt.Sprintf("写入初始种子数据失败: %v", err))
		}
	}

	configFile, err := PersistDatabaseConfig(s.projectRoot, normalized.Driver, runtimeDSN)
	if err != nil {
		return nil, errors.InternalError(fmt.Sprintf("写入配置文件失败: %v", err))
	}

	s.activeDSN = runtimeDSN
	s.activeDriver = normalized.Driver
	s.reason = ""
	s.initialized.Store(true)

	if s.onInitialized != nil {
		cb := s.onInitialized
		go func() {
			time.Sleep(200 * time.Millisecond)
			cb()
		}()
	}

	return &InitResult{
		Initialized:     true,
		ConfigFile:      configFile,
		Driver:          normalized.Driver,
		Host:            normalized.Host,
		Port:            normalized.Port,
		Database:        displayDB,
		DefaultUsername: "admin",
		Message:         "数据库初始化完成，配置已保存，正在切换服务状态",
	}, nil
}

func (s *Service) openAndPrepareDatabase(ctx context.Context, in DatabaseInitInput) (*gorm.DB, string, string, error) {
	switch in.Driver {
	case "mysql":
		serverDSN := BuildMySQLDSN(in, "", 5*time.Second)
		serverDB, err := sql.Open("mysql", serverDSN)
		if err != nil {
			return nil, "", "", errors.BadRequest(fmt.Sprintf("打开 MySQL 服务器连接失败: %v", err))
		}
		if err := serverDB.PingContext(ctx); err != nil {
			_ = serverDB.Close()
			return nil, "", "", errors.BadRequest(fmt.Sprintf("连接 MySQL 服务器失败: %v", err))
		}
		createSQL := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci", in.Database)
		if _, err := serverDB.ExecContext(ctx, createSQL); err != nil {
			_ = serverDB.Close()
			return nil, "", "", errors.InternalError(fmt.Sprintf("创建 MySQL 数据库失败: %v", err))
		}
		_ = serverDB.Close()

		runtimeDSN := BuildMySQLDSN(in, in.Database, 0)
		connectDSN := BuildMySQLDSN(in, in.Database, 10*time.Second)
		gormDB, err := gorm.Open(gormmysql.Open(connectDSN), &gorm.Config{})
		if err != nil {
			return nil, "", "", errors.InternalError(fmt.Sprintf("连接 MySQL 目标数据库失败: %v", err))
		}
		return gormDB, runtimeDSN, in.Database, nil

	case "postgres":
		adminDSN := BuildPostgresDSN(in, "postgres", 5*time.Second)
		adminDB, err := sql.Open("pgx", adminDSN)
		if err != nil {
			return nil, "", "", errors.BadRequest(fmt.Sprintf("打开 PostgreSQL 连接失败: %v", err))
		}
		if err := adminDB.PingContext(ctx); err != nil {
			_ = adminDB.Close()
			return nil, "", "", errors.BadRequest(fmt.Sprintf("连接 PostgreSQL 服务器失败: %v", err))
		}
		var exists int
		checkErr := adminDB.QueryRowContext(ctx, "SELECT 1 FROM pg_database WHERE datname = $1", in.Database).Scan(&exists)
		if checkErr == sql.ErrNoRows {
			createSQL := fmt.Sprintf(`CREATE DATABASE "%s"`, in.Database)
			if _, err := adminDB.ExecContext(ctx, createSQL); err != nil {
				_ = adminDB.Close()
				return nil, "", "", errors.InternalError(fmt.Sprintf("创建 PostgreSQL 数据库失败: %v", err))
			}
		} else if checkErr != nil {
			_ = adminDB.Close()
			return nil, "", "", errors.InternalError(fmt.Sprintf("检查 PostgreSQL 数据库失败: %v", checkErr))
		}
		_ = adminDB.Close()

		runtimeDSN := BuildPostgresDSN(in, in.Database, 0)
		connectDSN := BuildPostgresDSN(in, in.Database, 10*time.Second)
		gormDB, err := gorm.Open(gormpostgres.Open(connectDSN), &gorm.Config{})
		if err != nil {
			return nil, "", "", errors.InternalError(fmt.Sprintf("连接 PostgreSQL 目标数据库失败: %v", err))
		}
		return gormDB, runtimeDSN, in.Database, nil

	case "sqlite":
		relDSN, absPath := resolveSQLitePaths(s.projectRoot, in.Database)
		if err := os.MkdirAll(filepath.Dir(absPath), 0o755); err != nil {
			return nil, "", "", errors.InternalError(fmt.Sprintf("创建 SQLite 目录失败: %v", err))
		}
		gormDB, err := gorm.Open(gormsqlite.Open(absPath), &gorm.Config{})
		if err != nil {
			return nil, "", "", errors.InternalError(fmt.Sprintf("打开 SQLite 数据库失败: %v", err))
		}
		return gormDB, relDSN, relDSN, nil

	default:
		return nil, "", "", errors.BadRequest("不支持的数据库类型")
	}
}

func validateInput(input DatabaseInitInput) (DatabaseInitInput, error) {
	rawDriver := strings.ToLower(strings.TrimSpace(input.Driver))
	if rawDriver == "" {
		rawDriver = defaultDriver
	}
	canonicalDriver, err := migrations.NormalizeDriver(rawDriver)
	if err != nil {
		return input, errors.BadRequest("当前支持 MySQL、MariaDB、PostgreSQL、SQLite 数据库初始化")
	}

	in := DatabaseInitInput{
		Driver:   canonicalDriver,
		Host:     strings.TrimSpace(input.Host),
		Port:     input.Port,
		Username: strings.TrimSpace(input.Username),
		Password: input.Password,
		Database: strings.TrimSpace(input.Database),
	}

	if in.Driver == "sqlite" {
		if in.Database == "" || containsControlChars(in.Database) {
			return in, errors.BadRequest("请输入有效的 SQLite 数据库文件路径或名称")
		}
		in.Host = ""
		in.Port = 0
		in.Username = ""
		in.Password = ""
		return in, nil
	}

	if in.Host == "" || containsControlChars(in.Host) || strings.ContainsAny(in.Host, " \t") {
		return in, errors.BadRequest("请输入有效的数据库主机/IP")
	}
	if in.Port < 1 || in.Port > 65535 {
		return in, errors.BadRequest("数据库端口必须在 1-65535 之间")
	}
	if in.Username == "" || containsControlChars(in.Username) {
		return in, errors.BadRequest("请输入有效的数据库账号")
	}
	if containsControlChars(in.Password) {
		return in, errors.BadRequest("数据库密码不能包含换行或控制字符")
	}
	if !validDBNamePattern.MatchString(in.Database) {
		return in, errors.BadRequest("数据库名称仅允许字母、数字和下划线，长度 1-64 位")
	}
	return in, nil
}

func containsControlChars(s string) bool {
	for _, r := range s {
		if r == '\r' || r == '\n' || r == 0 {
			return true
		}
	}
	return false
}

func resolveSQLitePaths(projectRoot, dbInput string) (configDSN string, absPath string) {
	cleaned := strings.TrimSpace(dbInput)
	if !strings.ContainsAny(cleaned, `/\`) && !strings.HasSuffix(strings.ToLower(cleaned), ".db") && !strings.HasSuffix(strings.ToLower(cleaned), ".sqlite") && !strings.HasSuffix(strings.ToLower(cleaned), ".sqlite3") {
		cleaned = filepath.ToSlash(filepath.Join("data", cleaned+".db"))
	} else {
		cleaned = filepath.ToSlash(cleaned)
	}
	if filepath.IsAbs(cleaned) {
		return cleaned, filepath.Clean(cleaned)
	}
	return cleaned, filepath.Join(projectRoot, filepath.FromSlash(cleaned))
}

// BuildMySQLDSN formats a safe MySQL DSN using go-sql-driver/mysql Config.
func BuildMySQLDSN(in DatabaseInitInput, dbName string, timeout time.Duration) string {
	cfg := mysql.NewConfig()
	cfg.User = in.Username
	cfg.Passwd = in.Password
	cfg.Net = "tcp"
	cfg.Addr = net.JoinHostPort(in.Host, strconv.Itoa(in.Port))
	cfg.DBName = dbName
	cfg.ParseTime = true
	cfg.Loc = time.Local
	cfg.Params = map[string]string{
		"charset": "utf8mb4",
	}
	if timeout > 0 {
		cfg.Timeout = timeout
	}
	return cfg.FormatDSN()
}

// BuildPostgresDSN formats a safe PostgreSQL connection URL.
func BuildPostgresDSN(in DatabaseInitInput, dbName string, timeout time.Duration) string {
	q := url.Values{}
	q.Set("sslmode", "disable")
	q.Set("TimeZone", "Asia/Shanghai")
	if timeout > 0 {
		secs := int(timeout.Seconds())
		if secs < 1 {
			secs = 1
		}
		q.Set("connect_timeout", strconv.Itoa(secs))
	}
	u := &url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(in.Username, in.Password),
		Host:     net.JoinHostPort(in.Host, strconv.Itoa(in.Port)),
		Path:     "/" + dbName,
		RawQuery: q.Encode(),
	}
	return u.String()
}

// ParseDSNInfo extracts host, port, and database name from a DSN without exposing credentials.
func ParseDSNInfo(driver, dsn string) (host string, port int, dbName string) {
	trimmed := strings.TrimSpace(dsn)
	if trimmed == "" {
		return "", 0, ""
	}
	canonicalDriver, err := migrations.NormalizeDriver(driver)
	if err != nil {
		canonicalDriver = "mysql"
	}

	switch canonicalDriver {
	case "sqlite":
		return "", 0, trimmed
	case "postgres":
		if u, err := url.Parse(trimmed); err == nil && u.Scheme != "" {
			dbName = strings.TrimPrefix(u.Path, "/")
			if h, pStr, splitErr := net.SplitHostPort(u.Host); splitErr == nil {
				host = h
				if p, convErr := strconv.Atoi(pStr); convErr == nil {
					port = p
				}
			} else {
				host = u.Host
			}
			return host, port, dbName
		}
		// Fallback for keyword=value DSN
		for _, part := range strings.Fields(trimmed) {
			kv := strings.SplitN(part, "=", 2)
			if len(kv) != 2 {
				continue
			}
			switch strings.ToLower(kv[0]) {
			case "host":
				host = kv[1]
			case "port":
				if p, convErr := strconv.Atoi(kv[1]); convErr == nil {
					port = p
				}
			case "dbname":
				dbName = kv[1]
			}
		}
		return host, port, dbName
	default:
		cfg, err := mysql.ParseDSN(trimmed)
		if err != nil {
			return "", 0, ""
		}
		dbName = cfg.DBName
		if h, pStr, splitErr := net.SplitHostPort(cfg.Addr); splitErr == nil {
			host = h
			if p, convErr := strconv.Atoi(pStr); convErr == nil {
				port = p
			}
		} else {
			host = cfg.Addr
		}
		return host, port, dbName
	}
}

// PersistDatabaseConfig writes database.driver and database.dsn directly into
// <projectRoot>/config/app.yaml, removes any stale DATABASE_DSN override from .env,
// and ensures non-placeholder JWT/service secrets exist in .env when required.
func PersistDatabaseConfig(projectRoot, driver, dsn string) (string, error) {
	if strings.TrimSpace(projectRoot) == "" {
		return "", fmt.Errorf("project root is empty")
	}
	canonicalDriver, err := migrations.NormalizeDriver(driver)
	if err != nil {
		return "", err
	}

	appYamlPath := filepath.Join(projectRoot, "config", "app.yaml")
	if err := writeDatabaseToAppYAML(appYamlPath, canonicalDriver, dsn); err != nil {
		return "", err
	}

	envPath := filepath.Join(projectRoot, ".env")
	if err := syncEnvSecretsAndClearDSN(envPath); err != nil {
		return "", err
	}
	_ = os.Unsetenv("DATABASE_DSN")
	if _, statErr := os.Stat(envPath); statErr == nil {
		_ = gotenv.OverLoad(envPath)
	}

	return "config/app.yaml", nil
}

func syncEnvSecretsAndClearDSN(envPath string) error {
	var existingLines []string
	if raw, err := os.ReadFile(envPath); err == nil {
		normalized := strings.ReplaceAll(string(raw), "\r\n", "\n")
		existingLines = strings.Split(normalized, "\n")
		if len(existingLines) > 0 && existingLines[len(existingLines)-1] == "" {
			existingLines = existingLines[:len(existingLines)-1]
		}
	} else if !os.IsNotExist(err) {
		return err
	}

	seenJWT := false
	seenSvcAuth := false

	outLines := make([]string, 0, len(existingLines)+2)
	for _, line := range existingLines {
		trimmed := strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(trimmed, "DATABASE_DSN="):
			// 数据库连接统一写入 config/app.yaml，移除 .env 中的 DATABASE_DSN 防止 Viper AutomaticEnv 覆盖
			continue
		case strings.HasPrefix(trimmed, "JWT_SECRET="):
			val := strings.TrimSpace(strings.TrimPrefix(trimmed, "JWT_SECRET="))
			if isPlaceholderSecret(val) {
				val = generateRandomHex(32)
			}
			outLines = append(outLines, "JWT_SECRET="+val)
			_ = os.Setenv("JWT_SECRET", val)
			seenJWT = true
		case strings.HasPrefix(trimmed, "SERVICE_AUTH_SECRET="):
			val := strings.TrimSpace(strings.TrimPrefix(trimmed, "SERVICE_AUTH_SECRET="))
			if isPlaceholderSecret(val) {
				val = generateRandomHex(32)
			}
			outLines = append(outLines, "SERVICE_AUTH_SECRET="+val)
			_ = os.Setenv("SERVICE_AUTH_SECRET", val)
			seenSvcAuth = true
		default:
			outLines = append(outLines, line)
		}
	}

	if !seenJWT {
		secret := os.Getenv("JWT_SECRET")
		if isPlaceholderSecret(secret) {
			secret = generateRandomHex(32)
		}
		outLines = append(outLines, "JWT_SECRET="+secret)
		_ = os.Setenv("JWT_SECRET", secret)
	}
	if !seenSvcAuth {
		secret := os.Getenv("SERVICE_AUTH_SECRET")
		if isPlaceholderSecret(secret) {
			secret = generateRandomHex(32)
		}
		outLines = append(outLines, "SERVICE_AUTH_SECRET="+secret)
		_ = os.Setenv("SERVICE_AUTH_SECRET", secret)
	}

	content := strings.Join(outLines, "\n") + "\n"
	return os.WriteFile(envPath, []byte(content), 0o600)
}

func writeDatabaseToAppYAML(appYamlPath, driver, dsn string) error {
	if err := os.MkdirAll(filepath.Dir(appYamlPath), 0o755); err != nil {
		return err
	}

	raw, err := os.ReadFile(appYamlPath)
	if err != nil {
		if os.IsNotExist(err) {
			content := fmt.Sprintf("database:\n  driver: %s\n  dsn: %s\n", driver, strconv.Quote(dsn))
			return os.WriteFile(appYamlPath, []byte(content), 0o644)
		}
		return err
	}

	text := strings.ReplaceAll(string(raw), "\r\n", "\n")
	lines := strings.Split(text, "\n")

	inDatabase := false
	hasDatabaseSection := false
	driverUpdated := false
	dsnUpdated := false
	databaseSectionEnd := len(lines)

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		if !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t") {
			if inDatabase {
				databaseSectionEnd = i
			}
			inDatabase = trimmed == "database:"
			if inDatabase {
				hasDatabaseSection = true
			}
			continue
		}
		if !inDatabase {
			continue
		}
		databaseSectionEnd = i + 1
		if strings.HasPrefix(trimmed, "driver:") {
			lines[i] = "  driver: " + driver
			driverUpdated = true
		} else if strings.HasPrefix(trimmed, "dsn:") {
			lines[i] = "  dsn: " + strconv.Quote(dsn)
			dsnUpdated = true
		}
	}

	if !hasDatabaseSection {
		if len(lines) > 0 && lines[len(lines)-1] != "" {
			lines = append(lines, "")
		}
		lines = append(lines, "database:", "  driver: "+driver, "  dsn: "+strconv.Quote(dsn))
	} else {
		var missing []string
		if !driverUpdated {
			missing = append(missing, "  driver: "+driver)
		}
		if !dsnUpdated {
			missing = append(missing, "  dsn: "+strconv.Quote(dsn))
		}
		if len(missing) > 0 {
			updated := make([]string, 0, len(lines)+len(missing))
			updated = append(updated, lines[:databaseSectionEnd]...)
			updated = append(updated, missing...)
			updated = append(updated, lines[databaseSectionEnd:]...)
			lines = updated
		}
	}

	return os.WriteFile(appYamlPath, []byte(strings.Join(lines, "\n")), 0o644)
}

func isPlaceholderSecret(val string) bool {
	v := strings.TrimSpace(val)
	return v == "" || strings.HasPrefix(v, "replace-with-") || strings.HasPrefix(v, "setup-placeholder-")
}

func generateRandomHex(numBytes int) string {
	b := make([]byte, numBytes)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("gorp-secret-%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}
