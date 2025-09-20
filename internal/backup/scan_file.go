package winbackupchecker

import (
	"archive/zip"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func ScanFileBackupDir(root string, quiet bool) (*ScanReport, error) {
	if !quiet {
		fmt.Println("Scanning file backup root:", root)
	}

	report := &ScanReport{Root: root, Reports: []BackupReport{}}

	// Each backup set is usually a folder under root
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, fmt.Errorf("failed to read backup root: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		backupDir := filepath.Join(root, entry.Name())
		br := validateFileBackupSet(backupDir, quiet)
		report.Reports = append(report.Reports, br)

		if !quiet {
			if br.Valid {
				fmt.Println("✅ File backup set looks valid:", backupDir)
			} else {
				fmt.Printf("❌ File backup set invalid: %s\n   Errors: %v\n", backupDir, br.Errors)
			}
		}
	}

	return report, nil
}

func validateFileBackupSet(backupDir string, quiet bool) BackupReport {
	errors := []string{}

	// Check Catalogs folder
	catalogDir := filepath.Join(backupDir, "Catalogs")
	if !dirExists(catalogDir) {
		errors = append(errors, "missing Catalogs folder")
	} else {
		hasCatalog := false
		filepath.Walk(catalogDir, func(path string, info os.FileInfo, err error) error {
			if err == nil && !info.IsDir() && filepath.Ext(path) == ".wbcat" {
				hasCatalog = true
			}
			return nil
		})
		if !hasCatalog {
			errors = append(errors, "no .wbcat catalog found")
		}
	}

	// Look for .zip backup files
	foundZips := false
	filepath.Walk(backupDir, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() && filepath.Ext(path) == ".zip" {
			foundZips = true
			if err := checkZipReadable(path); err != nil {
				errors = append(errors, fmt.Sprintf("unreadable zip %s: %v", path, err))
			}
		}
		return nil
	})
	if !foundZips {
		errors = append(errors, "no .zip backup files found")
	}

	return BackupReport{
		BackupDir: backupDir,
		Valid:     len(errors) == 0,
		Errors:    errors,
		CheckedAt: time.Now().Format(time.RFC3339),
	}
}

func checkZipReadable(path string) error {
	r, err := zip.OpenReader(path)
	if err != nil {
		return err
	}
	defer r.Close()

	// Try listing first file
	if len(r.File) > 0 {
		f, err := r.File[0].Open()
		if err != nil {
			return fmt.Errorf("failed to open first entry: %w", err)
		}
		defer f.Close()
	}
	return nil
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
