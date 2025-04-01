package database

import (
	"database/sql"
	"fmt"

	"github.com/dev-pt-bai/cataloging/configs"
	_ "github.com/lib/pq"
)

func NewClient(config *configs.Config) (*sql.DB, error) {
	db, err := sql.Open("postgres", fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=%s", config.Database.User, config.Database.Password, config.Database.Host, config.Database.Name, config.Database.SSLMode))
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to verify connection to database: %v", err)
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)

	return db, nil
}
