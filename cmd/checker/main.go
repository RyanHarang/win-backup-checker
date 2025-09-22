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
	Timestamp string                        `json:"timestamp"`
	Results   []winbackupchecker.ScanReport `json:"results"`
}

func main() {
	jsonOnly := flag.Bool("json", false, "Output results as JSON only (no human-readable logs)")
	jsonOut := flag.String("json-out", "", "Write JSON report to a file (NDJSON format if file already exists)")
	quiet := flag.Bool("quiet", false, "Suppress all console output (only write to --json-out if provided)")
	flag.Parse()

	// Config location
	configPath := filepath.Join("configs", "config.json")

	// Load config
	cfg, err := winbackupchecker.LoadConfig(configPath)
	if err != nil {
		if !*quiet {
			log.Printf("Error loading config: %v", err)
		}
		os.Exit(2)
	}

	if !*jsonOnly && !*quiet {
		fmt.Println("Loaded config:", cfg)
	}

	allReports := []winbackupchecker.ScanReport{}
	fatalErrors := []string{}

	// Run scan for each path
	for _, path := range cfg.BackupPaths {
		report, err := winbackupchecker.ScanFileBackupDir(path, *quiet)
		if err != nil {
			fatalErrors = append(fatalErrors, fmt.Sprintf("Scan failed for %s: %v", path, err))
			allReports = append(allReports, winbackupchecker.ScanReport{
				Root: path,
				Reports: []winbackupchecker.BackupReport{
					{
						BackupDir: path,
						Valid:     false,
						Errors:    []string{err.Error()},
						CheckedAt: time.Now().Format(time.RFC3339),
					},
				},
			})
			continue
		}
		allReports = append(allReports, *report)
	}

	runReport := RunReport{
		Timestamp: time.Now().Format(time.RFC3339),
		Results:   allReports,
	}

	jsonData, err := json.MarshalIndent(runReport, "", "  ")
	if err != nil {
		if !*quiet {
			log.Printf("Failed to marshal report: %v", err)
		}
		os.Exit(2)
	}

	// Console output if not quiet
	if !*quiet {
		if *jsonOnly {
			fmt.Println(string(jsonData))
		} else {
			fmt.Println("\n===== JSON Validation Report =====")
			fmt.Println(string(jsonData))
		}
	}

	// Optional file output
	if *jsonOut != "" {
		line, err := json.Marshal(runReport)
		if err != nil {
			if !*quiet {
				log.Printf("Failed to marshal run report for file output: %v", err)
			}
			os.Exit(2)
		}
		f, err := os.OpenFile(*jsonOut, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			if !*quiet {
				log.Printf("Failed to open JSON output file: %v", err)
			}
			os.Exit(2)
		}
		defer f.Close()

		if _, err := f.Write(append(line, '\n')); err != nil {
			if !*quiet {
				log.Printf("Failed to write to JSON output file: %v", err)
			}
			os.Exit(2)
		}

		if !*jsonOnly && !*quiet {
			fmt.Printf("Appended JSON report to %s\n", *jsonOut)
		}
	}

	os.Exit(decideExitCode(fatalErrors, allReports))
}

func decideExitCode(fatalErrors []string, allReports []winbackupchecker.ScanReport) int {
	if len(fatalErrors) > 0 {
		return 2
	}
	for _, sr := range allReports {
		for _, br := range sr.Reports {
			if !br.Valid {
				return 1
			}
		}
	}
	return 0
}

/*
Usage:
  go run ./cmd/checker/                              # Check file backups
  go run ./cmd/checker/ --json                       # JSON only
  go run ./cmd/checker/ --json-out=logs.json         # Append JSON report to logs.json
  go run ./cmd/checker/ --quiet --json-out=logs.json # Silent, writes JSON only to file

Exit codes:
  0 = all backups valid
  1 = some backups invalid
  2 = fatal error (config, scan, or IO failure)
*/
