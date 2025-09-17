package winbackupchecker

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	BackupPaths []string `json:"backup_paths"`
	CheckHash   bool     `json:"check_hash"`
}

// LoadConfig loads JSON config file from given path
func LoadConfig(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	var cfg Config
	if err := decoder.Decode(&cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	if len(cfg.BackupPaths) == 0 {
		return nil, fmt.Errorf("no backup paths specified in config")
	}

	return &cfg, nil
}
