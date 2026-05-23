package db

import (
	"database/sql"
	"embed"
	"fmt"
	"log/slog"
	"sort"
	"strconv"
	"strings"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// Migrate applies any embedded migration files not yet recorded in the
// schema_migrations table. Versions are derived from the filename prefix
// (e.g. "001_initial_schema.sql" -> 1). Each migration runs inside its own
// transaction; a failure aborts that migration and stops the process.
func Migrate(conn *sql.DB) error {
	if _, err := conn.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version    integer PRIMARY KEY,
			name       text NOT NULL,
			applied_at timestamptz NOT NULL DEFAULT now()
		)
	`); err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	entries, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}

	type migration struct {
		version int
		name    string
		path    string
	}
	var pending []migration
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".sql") {
			continue
		}
		prefix, _, ok := strings.Cut(e.Name(), "_")
		if !ok {
			return fmt.Errorf("malformed migration filename %q (expected NNN_name.sql)", e.Name())
		}
		v, err := strconv.Atoi(prefix)
		if err != nil {
			return fmt.Errorf("migration %q: parse version: %w", e.Name(), err)
		}
		pending = append(pending, migration{version: v, name: e.Name(), path: "migrations/" + e.Name()})
	}
	sort.Slice(pending, func(i, j int) bool { return pending[i].version < pending[j].version })

	for _, m := range pending {
		var exists bool
		if err := conn.QueryRow("SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version=$1)", m.version).Scan(&exists); err != nil {
			return fmt.Errorf("check migration %d: %w", m.version, err)
		}
		if exists {
			continue
		}

		body, err := migrationsFS.ReadFile(m.path)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", m.name, err)
		}

		slog.Info("applying migration", "version", m.version, "name", m.name)

		tx, err := conn.Begin()
		if err != nil {
			return fmt.Errorf("begin tx for %s: %w", m.name, err)
		}
		if _, err := tx.Exec(string(body)); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("exec migration %s: %w", m.name, err)
		}
		if _, err := tx.Exec("INSERT INTO schema_migrations(version, name) VALUES($1, $2)", m.version, m.name); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("record migration %s: %w", m.name, err)
		}
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit migration %s: %w", m.name, err)
		}
	}

	return nil
}
