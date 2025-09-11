package main

import (
	"fmt"
	"log"
	"path/filepath"

	winbackupchecker "github.com/RyanHarang/win-backup-checker/internal/backup"
)

func main() {
	// Config file location
	configPath := filepath.Join("configs", "config.json")

	// Load config
	cfg, err := winbackupchecker.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	fmt.Println("Loaded config:", cfg)

	// Run scan for each path
	for _, path := range cfg.BackupPaths {
		if err := winbackupchecker.ScanBackupDir(path); err != nil {
			log.Printf("Scan failed for %s: %v", path, err)
		}
	}
}