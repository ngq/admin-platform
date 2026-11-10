package migrations

import (
	"io/fs"
	"strings"
	"testing"
)

func TestEmbeddedMigrationsContainAllDialectsAndTables(t *testing.T) {
	tables := []string{
		"sys_departments",
		"sys_users",
		"sys_roles",
		"sys_user_roles",
		"sys_menus",
		"sys_buttons",
		"sys_apis",
		"sys_role_menus",
		"sys_role_apis",
		"sys_role_buttons",
		"sys_params",
		"sys_operation_logs",
		"sys_login_logs",
		"sys_error_logs",
		"casbin_rule",
	}

	for _, dialect := range []string{"mysql", "postgres", "sqlite"} {
		entries, err := fs.Glob(FS, dialect+"/*.sql")
		if err != nil {
			t.Fatalf("glob %s migrations: %v", dialect, err)
		}
		if len(entries) == 0 {
			t.Fatalf("expected embedded SQL migration files for %s", dialect)
		}

		raw, err := fs.ReadFile(FS, dialect+"/00001_init_schema.sql")
		if err != nil {
			t.Fatalf("read %s/00001_init_schema.sql: %v", dialect, err)
		}
		sqlText := string(raw)

		upMarker := "-- +goose Up"
		downMarker := "-- +goose Down"
		upIdx := strings.Index(sqlText, upMarker)
		downIdx := strings.Index(sqlText, downMarker)
		if upIdx < 0 || downIdx < 0 || downIdx <= upIdx {
			t.Fatalf("%s/00001_init_schema.sql must contain valid %q and %q sections", dialect, upMarker, downMarker)
		}

		upSQL := sqlText[upIdx:downIdx]
		downSQL := sqlText[downIdx:]

		for _, table := range tables {
			if !strings.Contains(upSQL, table) {
				t.Errorf("%s Up migration missing table %s", dialect, table)
			}
			if !strings.Contains(downSQL, table) {
				t.Errorf("%s Down migration missing table %s", dialect, table)
			}
		}
	}
}
