package database

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/dev-pt-bai/cataloging/configs"
	_ "github.com/go-sql-driver/mysql"
)

func NewClient(config *configs.Config) (*sql.DB, error) {
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@/%s", config.Database.User, config.Database.Password, config.Database.Name))
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to verify connection to database: %v", err)
	}

	db.SetConnMaxLifetime(3 * time.Minute)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)

	return db, nil
}
