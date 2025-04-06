package configs

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"
)

type App struct {
	Port            uint16        `json:"port"`
	TokenExpiryHour time.Duration `json:"tokenExpiryHour"`
}

type Secret struct {
	JWT string `json:"jwt"`
}

type Database struct {
	User     string `json:"user"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

type Config struct {
	App      App      `json:"app"`
	Secret   Secret   `json:"secret"`
	Database Database `json:"database"`
}

func Load() (*Config, error) {
	f, err := os.Open("./configs/config.json")
	if err != nil {
		return nil, fmt.Errorf("failed to open config.json: %w", err)
	}
	defer f.Close()

	b, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("failed to read config.json: %w", err)
	}

	config := new(Config)
	if err = json.Unmarshal(b, config); err != nil {
		return nil, fmt.Errorf("failed to parse config.json: %w", err)
	}

	return config, nil
}
