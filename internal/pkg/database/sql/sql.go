package sql

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/dev-pt-bai/cataloging/configs"
	_ "github.com/go-sql-driver/mysql"
)

func NewClient(config *configs.Config) (*sql.DB, error) {
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@/%s", config.Database.SQL.User, config.Database.SQL.Password, config.Database.SQL.Name))
	if err != nil {
		return nil, fmt.Errorf("failed to open sql database: %w", err)
	}

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to verify connection to sql database: %w", err)
	}

	db.SetConnMaxLifetime(3 * time.Minute)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)

	return db, nil
}

type migration struct {
	name    string
	version int
}

func Migrate(config *configs.Config) error {
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@/%s?multiStatements=true", config.Database.SQL.User, config.Database.SQL.Password, config.Database.SQL.Name))
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (version BIGINT NOT NULL, dirty TINYINT(1) NOT NULL, PRIMARY KEY (version))`)
	if err != nil {
		return fmt.Errorf("failed to ensure schema_migrations table: %w", err)
	}

	_, err = db.Exec(`INSERT INTO schema_migrations (version, dirty) SELECT 0, 0 WHERE NOT EXISTS (SELECT * FROM schema_migrations)`)
	if err != nil {
		return fmt.Errorf("failed to initialize schema_migrations row: %w", err)
	}

	var version int
	var isDirty bool
	if err := db.QueryRow("SELECT version, dirty FROM schema_migrations").Scan(&version, &isDirty); err != nil {
		return fmt.Errorf("failed to read latest migration version: %w", err)
	}

	if isDirty {
		return fmt.Errorf("dirty migrations")
	}

	entries, err := os.ReadDir("./migrations")
	if err != nil {
		return fmt.Errorf("failed to open migrations directory: %w", err)
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	pending := make([]migration, 0, 5)
	for _, entry := range entries {
		entryName := entry.Name()
		if !strings.HasSuffix(entryName, ".up.sql") {
			continue
		}

		v, err := parseMigrationVersion(entryName)
		if err != nil {
			return err
		}

		if v <= version {
			continue
		}

		pending = append(pending, migration{name: entryName, version: v})
	}

	if len(pending) == 0 {
		return nil
	}

	for i := range pending {
		if err := up(db, pending[i]); err != nil {
			return err
		}
	}

	return nil
}

func up(db *sql.DB, migration migration) error {
	b, err := os.ReadFile(filepath.Join("./migrations", migration.name))
	if err != nil {
		return fmt.Errorf("failed to read migration script: %s, cause: %w", migration.name, err)
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start database transaction: %w", err)
	}
	defer tx.Rollback()

	if _, err = tx.Exec(string(b)); err != nil {
		if _, errDirty := tx.Exec("UPDATE schema_migrations SET version = ?, dirty = 1", migration.version); errDirty != nil {
			return fmt.Errorf("failed to update migration version: %w", errDirty)
		}

		if errCommit := tx.Commit(); errCommit != nil {
			return fmt.Errorf("failed to commit transaction: %w", errCommit)
		}

		return fmt.Errorf("failed to run migration script: %s, cause: %w", migration.name, err)
	}

	if _, err = tx.Exec("UPDATE schema_migrations SET version = ?", migration.version); err != nil {
		return fmt.Errorf("failed to update migration version: %w", err)
	}

	if errCommit := tx.Commit(); errCommit != nil {
		return fmt.Errorf("failed to commit transaction: %w", errCommit)
	}

	return nil
}

func parseMigrationVersion(name string) (int, error) {
	if len(name) < 6 {
		return 0, fmt.Errorf("invalid migration filename: %s", name)
	}
	return strconv.Atoi(name[:6])
}
