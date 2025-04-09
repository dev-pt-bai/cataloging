package database

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/dev-pt-bai/cataloging/configs"
	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/golang-migrate/migrate/v4/source/file"
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

	sourceInstance, err := new(file.File).Open("file://migrations")
	if err != nil {
		return fmt.Errorf("failed to create source instance: %w", err)
	}

	driver, err := mysql.WithInstance(db, &mysql.Config{})
	if err != nil {
		return fmt.Errorf("failed to create database instance: %w", err)
	}

	m, err := migrate.NewWithInstance("file", sourceInstance, "mysql", driver)
	if err != nil {
		return fmt.Errorf("failed to create migration instance: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to execute migration: %w", err)
	}

	return nil
}
