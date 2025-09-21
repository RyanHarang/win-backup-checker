package winbackupchecker

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type Config struct {
	BackupPaths []string `json:"backup_paths"`
	CheckHash   bool     `json:"check_hash"`
}

// BackupReport represents validation details for single backup folder
type BackupReport struct {
	BackupDir string   `json:"backup_dir"`
	Valid     bool     `json:"valid"`
	Errors    []string `json:"errors"`
	CheckedAt string   `json:"checked_at"`
}

// ScanReport represents results for one root path
type ScanReport struct {
	Root    string         `json:"root"`
	Reports []BackupReport `json:"reports"`
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

// Helper to create timestamp consistently
func NowRFC3339() string {
	return time.Now().Format(time.RFC3339)
}
