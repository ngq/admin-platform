package main

import (
	"context"
	"testing"
	"testing/fstest"

	"github.com/glebarez/sqlite"
	"github.com/ngq/admin-platform/internal/data/po"
	"github.com/ngq/admin-platform/migrations"
	gorp "github.com/ngq/gorp"
	datacontract "github.com/ngq/gorp/framework/contract/data"
	gorpMigrate "github.com/ngq/gorp/framework/pkg/migrate"
	"gorm.io/gorm"
)

type stubConfig struct {
	values map[string]string
}

func (c stubConfig) Env() string                                   { return c.values["app.env"] }
func (c stubConfig) Get(key string) any                            { return c.values[key] }
func (c stubConfig) GetString(key string) string                   { return c.values[key] }
func (c stubConfig) GetInt(string) int                             { return 0 }
func (c stubConfig) GetBool(string) bool                           { return false }
func (c stubConfig) GetFloat(string) float64                       { return 0 }
func (c stubConfig) Unmarshal(string, any) error                   { return nil }
func (c stubConfig) Watch(context.Context, string) (datacontract.ConfigWatcher, error) {
	return nil, nil
}
func (c stubConfig) Reload(context.Context) error { return nil }

func openTestSQLiteDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open in-memory sqlite: %v", err)
	}
	return db
}

func TestMigrateWithFS_ProductionUsesGooseAndSkipsAutoMigrate(t *testing.T) {
	db := openTestSQLiteDB(t)
	rt := &gorp.HTTPRuntime{
		DB: db,
		Config: stubConfig{values: map[string]string{
			"app.env":         "production",
			"database.driver": "sqlite",
		}},
	}

	if err := migrateWithFS(rt, migrations.FS, ""); err != nil {
		t.Fatalf("production sqlite migrate failed: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("unwrap sqlDB: %v", err)
	}
	ver, err := gorpMigrate.GetDBVersion(context.Background(), sqlDB, "sqlite")
	if err != nil {
		t.Fatalf("GetDBVersion failed: %v", err)
	}
	if ver != 1 {
		t.Fatalf("expected goose version 1, got %d", ver)
	}
	if !db.Migrator().HasTable(&po.AdminUserPO{}) || !db.Migrator().HasTable(&po.CasbinRulePO{}) {
		t.Fatal("expected goose migration to create sys_users and casbin_rule tables")
	}
}

func TestMigrateWithFS_DevHybridRunsGooseAndAutoMigrateAndSupportsGroupName(t *testing.T) {
	db := openTestSQLiteDB(t)
	rt := &gorp.HTTPRuntime{
		DB: db,
		Config: stubConfig{values: map[string]string{
			"app.env":               "dev",
			"database.driver":       "sqlite",
			"database.migrate_mode": "hybrid",
		}},
	}

	if err := migrateWithFS(rt, migrations.FS, ""); err != nil {
		t.Fatalf("dev hybrid migrate failed: %v", err)
	}

	for _, model := range []any{
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
	} {
		if !db.Migrator().HasTable(model) {
			t.Fatalf("expected table for %T to exist after dev migrate", model)
		}
	}

	api := po.APIPO{
		Group:  "user",
		Path:   "/api/v1/users",
		Method: "GET",
	}
	if err := db.Create(&api).Error; err != nil {
		t.Fatalf("create APIPO: %v", err)
	}
	var count int64
	if err := db.Model(&po.APIPO{}).Where("group_name = ?", "user").Count(&count).Error; err != nil {
		t.Fatalf("query by group_name: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 row matching group_name='user', got %d", count)
	}
}

func TestRunGooseUp_AppliesSQLMigrations(t *testing.T) {
	db := openTestSQLiteDB(t)
	rt := &gorp.HTTPRuntime{DB: db}

	memFS := fstest.MapFS{
		"00001_create_items.sql": &fstest.MapFile{
			Data: []byte(`-- +goose Up
CREATE TABLE IF NOT EXISTS items (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS items;
`),
		},
	}

	if err := runGooseUp(context.Background(), rt, "sqlite", memFS, "."); err != nil {
		t.Fatalf("runGooseUp failed: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("unwrap sqlDB: %v", err)
	}
	ver, err := gorpMigrate.GetDBVersion(context.Background(), sqlDB, "sqlite")
	if err != nil {
		t.Fatalf("GetDBVersion failed: %v", err)
	}
	if ver != 1 {
		t.Fatalf("expected migration version 1, got %d", ver)
	}
	if !db.Migrator().HasTable("items") {
		t.Fatal("expected items table to exist after goose up")
	}
}

func TestMaybeHandleMigrateCLI(t *testing.T) {
	if handled, err := maybeHandleMigrateCLI(nil); handled || err != nil {
		t.Fatalf("expected unhandled for empty args, got handled=%v err=%v", handled, err)
	}
	if handled, err := maybeHandleMigrateCLI([]string{"serve"}); handled || err != nil {
		t.Fatalf("expected unhandled for non-migrate args, got handled=%v err=%v", handled, err)
	}
	if handled, err := maybeHandleMigrateCLI([]string{"migrate"}); !handled || err == nil {
		t.Fatalf("expected handled usage error for bare migrate command, got handled=%v err=%v", handled, err)
	}
}
