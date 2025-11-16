package main

import (
	"context"
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
	Summary   ScanSummary                   `json:"summary"`
}

type ScanSummary struct {
	TotalBackups   int `json:"total_backups"`
	ValidBackups   int `json:"valid_backups"`
	InvalidBackups int `json:"invalid_backups"`
	FailedScans    int `json:"failed_scans"`
}

func main() {
	jsonOnly := flag.Bool("json", false, "Output results as JSON only (no human-readable logs)")
	jsonOut := flag.String("json-out", "logs.json", "Write JSON report to a file (NDJSON format)")
	noLog := flag.Bool("no-log", false, "Disable writing to log file")
	parallel := flag.Int("parallel", 4, "Number of backup sets to validate concurrently")
	timeout := flag.Duration("timeout", 30*time.Minute, "Timeout for entire scan operation")
	noEmail := flag.Bool("no-email", false, "Disable email notifications even if configured")
	flag.Parse()

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	// Config location
	configPath := filepath.Join("configs", "config.json")
	emailConfigPath := filepath.Join("configs", "email.config.json")

	// Load config
	cfg, err := winbackupchecker.LoadConfig(configPath)
	if err != nil {
		log.Printf("Error loading config: %v", err)
		os.Exit(2)
	}

	// Load email config (optional)
	emailCfg, err := winbackupchecker.LoadEmailConfig(emailConfigPath)
	if err != nil {
		log.Printf("Error loading email config: %v", err)
		os.Exit(2)
	}

	if !*jsonOnly {
		fmt.Printf("Loaded config with %d backup paths, parallel workers: %d\n", len(cfg.BackupPaths), *parallel)
		if emailCfg != nil && emailCfg.Enabled && !*noEmail {
			fmt.Printf("Email notifications: enabled (to: %v)\n", emailCfg.To)
		}
		if !*noLog {
			fmt.Printf("Logging to: %s\n", *jsonOut)
		}
	}

	allReports := []winbackupchecker.ScanReport{}
	fatalErrors := []string{}

	// Run scan for each path with controlled concurrency
	for _, path := range cfg.BackupPaths {
		report, err := winbackupchecker.ScanFileBackupDir(ctx, path, *parallel)
		if err != nil {
			fatalErrors = append(fatalErrors, fmt.Sprintf("Scan failed for %s: %v", path, err))
			allReports = append(allReports, winbackupchecker.ScanReport{
				Root: path,
				Reports: []winbackupchecker.BackupReport{
					{
						BackupDir: path,
						Valid:     false,
						Issues: []winbackupchecker.ValidationIssue{
							winbackupchecker.NewValidationIssue(
								winbackupchecker.SeverityCritical,
								err.Error(),
								path,
								"check path accessibility and permissions",
							),
						},
						CheckedAt: time.Now().Format(time.RFC3339),
					},
				},
			})
			continue
		}
		allReports = append(allReports, *report)
	}

	summary := calculateSummary(allReports, fatalErrors)
	runReport := RunReport{
		Timestamp: time.Now().Format(time.RFC3339),
		Results:   allReports,
		Summary:   summary,
	}

	jsonData, err := json.MarshalIndent(runReport, "", "  ")
	if err != nil {
		log.Printf("Failed to marshal report: %v", err)
		os.Exit(2)
	}

	if *jsonOnly {
		fmt.Println(string(jsonData))
	} else {
		printSummary(summary)
		fmt.Println("\n===== JSON Validation Report =====")
		fmt.Println(string(jsonData))
	}

	// Write to log file (default behavior unless --no-log is set)
	if !*noLog {
		if err := writeJSONOutput(*jsonOut, runReport); err != nil {
			log.Printf("Failed to write JSON output: %v", err)
			os.Exit(2)
		}
		if !*jsonOnly {
			fmt.Printf("\nAppended report to %s\n", *jsonOut)
		}
	}

	if !*noEmail && emailCfg != nil && emailCfg.Enabled {
		if !*jsonOnly {
			fmt.Println("\nSending email notification...")
		}

		emailSummary := winbackupchecker.ScanSummary{
			TotalBackups:   summary.TotalBackups,
			ValidBackups:   summary.ValidBackups,
			InvalidBackups: summary.InvalidBackups,
			FailedScans:    summary.FailedScans,
		}

		if err := winbackupchecker.SendEmailAlert(emailCfg, emailSummary, allReports); err != nil {
			log.Printf("Failed to send email alert: %v", err)
		} else if !*jsonOnly {
			fmt.Println("Email notification sent successfully")
		}
	}

	os.Exit(decideExitCode(fatalErrors, allReports))
}

func calculateSummary(allReports []winbackupchecker.ScanReport, fatalErrors []string) ScanSummary {
	summary := ScanSummary{
		FailedScans: len(fatalErrors),
	}

	for _, sr := range allReports {
		for _, br := range sr.Reports {
			summary.TotalBackups++
			if br.Valid {
				summary.ValidBackups++
			} else {
				summary.InvalidBackups++
			}
		}
	}

	return summary
}

func printSummary(summary ScanSummary) {
	fmt.Printf("\n===== Backup Validation Summary =====\n")
	fmt.Printf("Total Backups: %d\n", summary.TotalBackups)
	fmt.Printf("Valid Backups: %d\n", summary.ValidBackups)
	fmt.Printf("Invalid Backups: %d\n", summary.InvalidBackups)
	fmt.Printf("Failed Scans: %d\n", summary.FailedScans)

	if summary.TotalBackups > 0 {
		validPercent := float64(summary.ValidBackups) / float64(summary.TotalBackups) * 100
		fmt.Printf("Success Rate: %.1f%%\n", validPercent)
	}
}

func writeJSONOutput(filename string, report RunReport) error {
	// Marshal with indentation for readability
	line, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal run report: %w", err)
	}

	// Check if file exists to determine format
	fileExists := false
	if info, err := os.Stat(filename); err == nil && info.Size() > 0 {
		fileExists = true
	}

	f, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open JSON output file: %w", err)
	}
	defer f.Close()

	if fileExists {
		if _, err := f.WriteString("\n---\n\n"); err != nil {
			return fmt.Errorf("failed to write separator: %w", err)
		}
	}

	if _, err := f.Write(append(line, '\n')); err != nil {
		return fmt.Errorf("failed to write to JSON output file: %w", err)
	}

	return nil
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
  go run ./cmd/checker/                                    # Check file backups (writes to logs.json by default)
  go run ./cmd/checker/ --json                             # JSON only output
  go run ./cmd/checker/ --json-out=custom.json             # Write to custom file
  go run ./cmd/checker/ --no-log                           # Don't write to log file
  go run ./cmd/checker/ --parallel=8                       # Use 8 concurrent workers
  go run ./cmd/checker/ --timeout=1h                       # Set 1 hour timeout
  go run ./cmd/checker/ --no-email                         # Disable email notifications

Exit codes:
  0 = all backups valid
  1 = some backups invalid
  2 = fatal error (config, scan, or IO failure)
*/
