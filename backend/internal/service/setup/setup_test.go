package setup

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ngq/admin-platform/internal/data/po"
	"github.com/ngq/admin-platform/internal/datainfra"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func TestValidateInput(t *testing.T) {
	tests := []struct {
		name    string
		input   DatabaseInitInput
		wantErr bool
	}{
		{
			name: "valid mysql input",
			input: DatabaseInitInput{
				Driver:   "mysql",
				Host:     "127.0.0.1",
				Port:     3306,
				Username: "root",
				Password: "p@ss:word/with#chars",
				Database: "admin_platform",
			},
		},
		{
			name: "valid mariadb alias",
			input: DatabaseInitInput{
				Driver:   "mariadb",
				Host:     "127.0.0.1",
				Port:     3306,
				Username: "root",
				Database: "admin_platform",
			},
		},
		{
			name: "valid postgres input",
			input: DatabaseInitInput{
				Driver:   "postgres",
				Host:     "127.0.0.1",
				Port:     5432,
				Username: "postgres",
				Password: "secret",
				Database: "admin_pg",
			},
		},
		{
			name: "valid sqlite input without host port user",
			input: DatabaseInitInput{
				Driver:   "sqlite",
				Database: "data/admin.db",
			},
		},
		{
			name: "reject sql injection in database name",
			input: DatabaseInitInput{
				Driver:   "mysql",
				Host:     "127.0.0.1",
				Port:     3306,
				Username: "root",
				Database: "admin; DROP TABLE sys_users;--",
			},
			wantErr: true,
		},
		{
			name: "reject newline in password",
			input: DatabaseInitInput{
				Driver:   "postgres",
				Host:     "127.0.0.1",
				Port:     5432,
				Username: "postgres",
				Password: "abc\nJWT_SECRET=hacked",
				Database: "admin_platform",
			},
			wantErr: true,
		},
		{
			name: "reject invalid port",
			input: DatabaseInitInput{
				Driver:   "mysql",
				Host:     "127.0.0.1",
				Port:     70000,
				Username: "root",
				Database: "admin_platform",
			},
			wantErr: true,
		},
		{
			name: "reject unsupported driver",
			input: DatabaseInitInput{
				Driver:   "oracle",
				Host:     "127.0.0.1",
				Port:     1521,
				Username: "system",
				Database: "admin_platform",
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := validateInput(tc.input)
			if tc.wantErr && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestBuildAndParseDSNs(t *testing.T) {
	mysqlIn := DatabaseInitInput{
		Driver:   "mysql",
		Host:     "10.0.0.8",
		Port:     3307,
		Username: "db_user",
		Password: "p@ss/word:123?abc",
		Database: "admin_prod",
	}
	mysqlDSN := BuildMySQLDSN(mysqlIn, mysqlIn.Database, 0)
	host, port, dbName := ParseDSNInfo("mysql", mysqlDSN)
	if host != "10.0.0.8" || port != 3307 || dbName != "admin_prod" {
		t.Fatalf("mysql ParseDSNInfo mismatch: got host=%q port=%d db=%q", host, port, dbName)
	}

	pgIn := DatabaseInitInput{
		Driver:   "postgres",
		Host:     "10.0.0.9",
		Port:     5432,
		Username: "pg_user",
		Password: "p@ss/word:123?abc",
		Database: "admin_pg",
	}
	pgDSN := BuildPostgresDSN(pgIn, pgIn.Database, 0)
	host, port, dbName = ParseDSNInfo("postgres", pgDSN)
	if host != "10.0.0.9" || port != 5432 || dbName != "admin_pg" {
		t.Fatalf("postgres ParseDSNInfo mismatch: got host=%q port=%d db=%q (dsn=%s)", host, port, dbName, pgDSN)
	}
	if strings.Contains(pgDSN, "p@ss/word") {
		t.Fatalf("postgres password should be URL-encoded, got %s", pgDSN)
	}

	host, port, dbName = ParseDSNInfo("sqlite", "data/admin_platform.db")
	if host != "" || port != 0 || dbName != "data/admin_platform.db" {
		t.Fatalf("sqlite ParseDSNInfo mismatch: got host=%q port=%d db=%q", host, port, dbName)
	}
}

func TestPersistDatabaseConfig_WritesAppYAMLAndClearsEnvDSN(t *testing.T) {
	tmp := t.TempDir()
	configDir := filepath.Join(tmp, "config")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatalf("mkdir config: %v", err)
	}

	appYaml := `app:
  name: admin-platform
database:
  driver: mysql
  dsn: env(DATABASE_DSN)
  migrate_mode: hybrid
`
	if err := os.WriteFile(filepath.Join(configDir, "app.yaml"), []byte(appYaml), 0o644); err != nil {
		t.Fatalf("write app.yaml: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmp, ".env"), []byte("DATABASE_DSN=old-dsn\nJWT_SECRET=keep-jwt\n"), 0o600); err != nil {
		t.Fatalf("write .env: %v", err)
	}

	pgDSN := "postgres://pg:secret@127.0.0.1:5432/admin_pg?sslmode=disable"
	written, err := PersistDatabaseConfig(tmp, "postgres", pgDSN)
	if err != nil {
		t.Fatalf("PersistDatabaseConfig failed: %v", err)
	}
	if written != "config/app.yaml" {
		t.Fatalf("expected config/app.yaml, got %q", written)
	}

	yamlBytes, err := os.ReadFile(filepath.Join(configDir, "app.yaml"))
	if err != nil {
		t.Fatalf("read app.yaml: %v", err)
	}
	yamlText := string(yamlBytes)
	if !strings.Contains(yamlText, "  driver: postgres") || !strings.Contains(yamlText, `  dsn: "`+pgDSN+`"`) {
		t.Fatalf("app.yaml driver/dsn update mismatch:\n%s", yamlText)
	}
	if !strings.Contains(yamlText, "migrate_mode: hybrid") {
		t.Fatalf("app.yaml should preserve migrate_mode:\n%s", yamlText)
	}

	envBytes, err := os.ReadFile(filepath.Join(tmp, ".env"))
	if err != nil {
		t.Fatalf("read .env: %v", err)
	}
	envText := string(envBytes)
	if strings.Contains(envText, "DATABASE_DSN=") {
		t.Fatalf(".env should not retain DATABASE_DSN override:\n%s", envText)
	}
	if !strings.Contains(envText, "JWT_SECRET=keep-jwt") {
		t.Fatalf(".env should preserve JWT_SECRET:\n%s", envText)
	}
}

func TestService_EndToEndSQLiteTestAndInitialize(t *testing.T) {
	tmp := t.TempDir()
	configDir := filepath.Join(tmp, "config")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatalf("mkdir config: %v", err)
	}
	appYaml := `app:
  name: admin-platform
database:
  driver: mysql
  dsn: env(DATABASE_DSN)
`
	if err := os.WriteFile(filepath.Join(configDir, "app.yaml"), []byte(appYaml), 0o644); err != nil {
		t.Fatalf("write app.yaml: %v", err)
	}

	t.Setenv("DATABASE_DSN", "")
	t.Setenv("JWT_SECRET", "")
	t.Setenv("SERVICE_AUTH_SECRET", "")

	seeded := false
	svc := NewService(Options{
		Initialized: false,
		ProjectRoot: tmp,
		SeedFn: func(db *gorm.DB, _ *datainfra.Data) error {
			seeded = true
			return db.Where(clause.Eq{Column: clause.Column{Name: "key"}, Value: "captcha_enabled"}).
				FirstOrCreate(&po.SysParamPO{Key: "captcha_enabled", Value: "false", Category: "security", Status: 1}).Error
		},
	})

	input := DatabaseInitInput{
		Driver:   "sqlite",
		Database: "data/test_setup.db",
	}

	probeRes, err := svc.TestConnection(context.Background(), input)
	if err != nil {
		t.Fatalf("TestConnection sqlite failed: %v", err)
	}
	if !probeRes.Connected || probeRes.DatabaseExists {
		t.Fatalf("unexpected sqlite probe result: %+v", probeRes)
	}

	initRes, err := svc.InitializeDatabase(context.Background(), input)
	if err != nil {
		t.Fatalf("InitializeDatabase sqlite failed: %v", err)
	}
	if !initRes.Initialized || initRes.Driver != "sqlite" || !seeded {
		t.Fatalf("unexpected sqlite init result: %+v seeded=%v", initRes, seeded)
	}
	if _, err := os.Stat(filepath.Join(tmp, "data", "test_setup.db")); err != nil {
		t.Fatalf("expected sqlite file to be created: %v", err)
	}

	yamlText, _ := os.ReadFile(filepath.Join(configDir, "app.yaml"))
	if !strings.Contains(string(yamlText), "  driver: sqlite") {
		t.Fatalf("expected app.yaml driver to be sqlite, got:\n%s", string(yamlText))
	}
}

func TestService_GuardsWhenAlreadyInitialized(t *testing.T) {
	svc := NewService(Options{
		Initialized:  true,
		ActiveDriver: "mysql",
		ActiveDSN:    "root:secret-pwd@tcp(10.1.2.3:3306)/admin_platform?charset=utf8mb4",
	})

	status := svc.GetStatus()
	if !status.Initialized || status.Host != "10.1.2.3" || status.Port != 3306 || status.Database != "admin_platform" {
		t.Fatalf("unexpected initialized status: %#v", status)
	}

	in := DatabaseInitInput{
		Driver:   "mysql",
		Host:     "127.0.0.1",
		Port:     3306,
		Username: "root",
		Database: "admin_platform",
	}
	if _, err := svc.TestConnection(context.Background(), in); err == nil {
		t.Fatal("TestConnection should reject calls when already initialized")
	}
	if _, err := svc.InitializeDatabase(context.Background(), in); err == nil {
		t.Fatal("InitializeDatabase should reject calls when already initialized")
	}
}
