package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	winbackupchecker "github.com/RyanHarang/win-backup-checker/internal/backup"
)

type RunReport struct {
	Timestamp string                            `json:"timestamp"`
	Results   []winbackupchecker.ScanReport     `json:"results"`
}

func main() {
	// Flags
	jsonOnly := flag.Bool("json", false, "Output results as JSON only (no human-readable logs)")
	jsonOut := flag.String("json-out", "", "Write JSON report to a file (NDJSON format if file already exists)")
	flag.Parse()

	// Config file location
	configPath := filepath.Join("configs", "config.json")

	// Load config
	cfg, err := winbackupchecker.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	if !*jsonOnly {
		fmt.Println("Loaded config:", cfg)
	}

	allReports := []winbackupchecker.ScanReport{}

	// Run scan for each path
	for _, path := range cfg.BackupPaths {
		report, err := winbackupchecker.ScanBackupDir(path)
		if err != nil {
			log.Printf("Scan failed for %s: %v", path, err)
			continue
		}
		allReports = append(allReports, *report)
	}

	// Wrap in run-level report with timestamp
	runReport := RunReport{
		Timestamp: time.Now().Format(time.RFC3339),
		Results:   allReports,
	}

	// Marshal JSON
	jsonData, err := json.MarshalIndent(runReport, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal report: %v", err)
	}

	// Console output
	if *jsonOnly {
		fmt.Println(string(jsonData))
	} else {
		fmt.Println("\n===== JSON Validation Report =====")
		fmt.Println(string(jsonData))
	}

	// Optional file output
	if *jsonOut != "" {
		line, err := json.Marshal(runReport)
		if err != nil {
			log.Fatalf("Failed to marshal run report for file output: %v", err)
		}
		f, err := os.OpenFile(*jsonOut, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			log.Fatalf("Failed to open JSON output file: %v", err)
		}
		defer f.Close()

		if _, err := f.Write(append(line, '\n')); err != nil {
			log.Fatalf("Failed to write to JSON output file: %v", err)
		}

		if !*jsonOnly {
			fmt.Printf("Appended JSON report to %s\n", *jsonOut)
		}
	}
}

/*
Usage:
  go run ./cmd/checker/                            	# Human-readable logs + JSON summary
  go run ./cmd/checker/ --json                     	# JSON only (no extra logs)
  go run ./cmd/checker/ --json-out=logs.json     		# Appends JSON report to logs.json
  go run ./cmd/checker/ --json --json-out=logs.json # JSON only, also appends to logs.json

Note:
  - The JSON log file uses NDJSON (one JSON object per line).
  - Each run is timestamped so you can track history.
*/
