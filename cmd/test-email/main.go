package main

import (
	"fmt"
	"log"
	"path/filepath"
	"time"

	winbackupchecker "github.com/RyanHarang/win-backup-checker/internal/backup"
)

func main() {
	emailConfigPath := filepath.Join("configs", "email.config.json")
	emailCfg, err := winbackupchecker.LoadEmailConfig(emailConfigPath)
	if err != nil {
		log.Fatalf("Error loading email config: %v", err)
	}

	// Check if email is configured
	if emailCfg == nil || !emailCfg.Enabled {
		log.Fatal("Email is not enabled in email.config.json")
	}

	fmt.Printf("Email Configuration:\n")
	fmt.Printf("  SMTP Host: %s:%d\n", emailCfg.SMTPHost, emailCfg.SMTPPort)
	fmt.Printf("  From: %s\n", emailCfg.From)
	fmt.Printf("  To: %v\n", emailCfg.To)
	fmt.Printf("  Send on Errors: %v\n", emailCfg.SendOnErrors)
	fmt.Printf("  Send on Warnings: %v\n", emailCfg.SendOnWarnings)
	fmt.Printf("  Send on Success: %v\n\n", emailCfg.SendOnSuccess)
	fmt.Println("Sending test email with mock error data...")

	mockReports := []winbackupchecker.ScanReport{
		{
			Root: "/home/harangr/projects/win-backup-checker/Backups",
			Reports: []winbackupchecker.BackupReport{
				{
					BackupDir: "/home/harangr/projects/win-backup-checker/Backups/TestMachine/Backup-2024-01-15",
					Valid:     false,
					Issues: []winbackupchecker.ValidationIssue{
						winbackupchecker.NewValidationIssue(
							winbackupchecker.SeverityError,
							"missing Catalogs folder",
							"/home/harangr/projects/win-backup-checker/Backups/TestMachine/Backup-2024-01-15/Catalogs",
							"backup set should contain a Catalogs folder with .wbcat files",
						),
						winbackupchecker.NewValidationIssue(
							winbackupchecker.SeverityWarning,
							"backup is quite old (95 days)",
							"/home/harangr/projects/win-backup-checker/Backups/TestMachine/Backup-2024-01-15",
							"consider creating more recent backups",
						),
					},
					CheckedAt: time.Now().Format(time.RFC3339),
					ValidationStats: winbackupchecker.ValidationStats{
						TotalFiles:     25,
						ValidatedFiles: 20,
						CorruptFiles:   2,
						TotalSize:      1024 * 1024 * 500, // 500MB
						CatalogFiles:   0,
						BackupFiles:    20,
					},
				},
				{
					BackupDir: "/home/harangr/projects/win-backup-checker/Backups/TestMachine/Backup-2024-02-01",
					Valid:     true,
					Issues:    []winbackupchecker.ValidationIssue{},
					CheckedAt: time.Now().Format(time.RFC3339),
					ValidationStats: winbackupchecker.ValidationStats{
						TotalFiles:     30,
						ValidatedFiles: 30,
						CorruptFiles:   0,
						TotalSize:      1024 * 1024 * 750, // 750MB
						CatalogFiles:   5,
						BackupFiles:    25,
					},
				},
			},
		},
	}

	mockSummary := winbackupchecker.ScanSummary{
		TotalBackups:   2,
		ValidBackups:   1,
		InvalidBackups: 1,
		FailedScans:    0,
	}

	err = winbackupchecker.SendEmailAlert(emailCfg, mockSummary, mockReports)
	if err != nil {
		log.Fatalf("Failed to send email: %v", err)
	}

	fmt.Println("Test email sent successfully!")
	fmt.Println("\nCheck your inbox at:", emailCfg.To[0])
	fmt.Println("\nIf you don't receive the email, check:")
	fmt.Println("  1. Spam/Junk folder")
	fmt.Println("  2. Gmail App Password is correct (not your regular password)")
	fmt.Println("  3. 'Less secure app access' or 2FA with App Password is enabled")
}

/*
Usage:
  go run ./cmd/test-email/
*/
