// Package migrations embeds dialect-specific SQL migration files managed by goose.
package migrations

import (
	"embed"
	"fmt"
	"strings"
)

// FS holds embedded SQL migration scripts grouped by database dialect (mysql, postgres, sqlite).
//
//go:embed mysql/*.sql postgres/*.sql sqlite/*.sql
var FS embed.FS

// NormalizeDriver normalizes user-facing / GORM driver aliases to gorp's canonical driver names.
func NormalizeDriver(driver string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(driver)) {
	case "", "mysql", "mariadb", "tidb":
		return "mysql", nil
	case "postgres", "postgresql", "pgsql", "pgx":
		return "postgres", nil
	case "sqlite", "sqlite3":
		return "sqlite", nil
	default:
		return "", fmt.Errorf("unsupported database driver %q (supported: mysql, mariadb, postgres, sqlite)", driver)
	}
}

// DirForDriver maps a database driver name to its embedded migrations subdirectory.
func DirForDriver(driver string) (string, error) {
	return NormalizeDriver(driver)
}
