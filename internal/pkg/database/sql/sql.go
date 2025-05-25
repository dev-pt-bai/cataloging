package sql

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/dev-pt-bai/cataloging/configs"
	_ "github.com/go-sql-driver/mysql"
)

func NewClient(config *configs.Config) (*sql.DB, error) {
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@/%s", config.Database.User, config.Database.Password, config.Database.Name))
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to verify connection to database: %w", err)
	}

	db.SetConnMaxLifetime(3 * time.Minute)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)

	return db, nil
}

func Migrate(config *configs.Config) error {
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@/%s?multiStatements=true", config.Database.User, config.Database.Password, config.Database.Name))
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

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

	lastIndex := len(entries) - 1
	for i := lastIndex; i >= 0; i-- {
		entryName := entries[i].Name()
		if !strings.HasSuffix(entryName, ".up.sql") {
			continue
		}

		v, err := strconv.Atoi(entryName[:6])
		if err != nil {
			return fmt.Errorf("failed to parse migration script version: %w", err)
		}

		if v <= version {
			break
		}

		lastIndex--
	}

	if lastIndex >= len(entries)-2 {
		return nil
	}

	for i := lastIndex; i <= len(entries)-1; i++ {
		entryName := entries[i].Name()
		if !strings.HasSuffix(entryName, ".up.sql") {
			continue
		}

		v, err := strconv.Atoi(entryName[:6])
		if err != nil {
			return fmt.Errorf("failed to parse migration script version: %w", err)
		}

		b, err := os.ReadFile(fmt.Sprintf("./migrations/%s", entryName))
		if err != nil {
			return fmt.Errorf("failed to read migration script: %s, cause: %w", entryName, err)
		}

		if _, err = db.Exec(string(b)); err != nil {
			if _, errDirty := db.Exec(fmt.Sprintf("BEGIN;UPDATE schema_migrations SET version = %d, dirty = 1;COMMIT;", v)); errDirty != nil {
				return fmt.Errorf("failed to update migration version: %w", errDirty)
			}

			return fmt.Errorf("failed to run migration script: %s, cause: %w", entryName, err)
		}

		if _, err = db.Exec(fmt.Sprintf("BEGIN;UPDATE schema_migrations SET version = %d;COMMIT;", v)); err != nil {
			return fmt.Errorf("failed to update migration version: %w", err)
		}
	}

	return nil
}
