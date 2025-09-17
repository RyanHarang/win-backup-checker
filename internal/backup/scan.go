package winbackupchecker

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type BackupReport struct {
	BackupDir string   `json:"backup_dir"`
	Valid     bool     `json:"valid"`
	Errors    []string `json:"errors"`
	CheckedAt string   `json:"checked_at"`
}

type ScanReport struct {
	Root    string         `json:"root"`
	Reports []BackupReport `json:"reports"`
}

// ScanBackupDir scans a root directory and returns a JSON-friendly report.
// If quiet = true, all console printing is suppressed.
func ScanBackupDir(root string, quiet bool) (*ScanReport, error) {
	if !quiet {
		fmt.Println("Scanning backup root:", root)
	}

	report := &ScanReport{Root: root, Reports: []BackupReport{}}

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if filepath.Base(path) == "BackupSpecs.xml" {
			backupSetDir := filepath.Dir(path)
			if !quiet {
				fmt.Printf("Found backup set: %s\n", backupSetDir)
			}

			rep := validateBackupSet(backupSetDir)
			report.Reports = append(report.Reports, rep)

			if !quiet {
				if rep.Valid {
					fmt.Println("✅ Backup set looks valid.")
				} else {
					fmt.Printf("❌ Backup set invalid: %v\n", rep.Errors)
				}
			}
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("scan failed: %w", err)
	}

	return report, nil
}

func validateBackupSet(backupDir string) BackupReport {
	errors := []string{}

	// 1. BackupSpecs.xml
	if !fileExists(filepath.Join(backupDir, "BackupSpecs.xml")) {
		errors = append(errors, "missing BackupSpecs.xml")
	}

	// 2. MediaId
	if !fileExists(filepath.Join(backupDir, "MediaId")) {
		errors = append(errors, "missing MediaId file")
	}

	// 3. Catalog folder with files
	catalogDir := filepath.Join(backupDir, "Catalog")
	if !dirExists(catalogDir) {
		errors = append(errors, "missing Catalog folder")
	} else {
		hasCatalogFile := false
		filepath.Walk(catalogDir, func(path string, info os.FileInfo, err error) error {
			if !info.IsDir() && (filepath.Ext(path) == ".wbcat" || filepath.Ext(path) == ".cat") {
				hasCatalogFile = true
			}
			return nil
		})
		if !hasCatalogFile {
			errors = append(errors, "no catalog files (*.wbcat or *.cat) found in Catalog folder")
		}
	}

	// 4. Disk images
	hasDiskImage := false
	filepath.Walk(backupDir, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() && (filepath.Ext(path) == ".vhd" || filepath.Ext(path) == ".vhdx") {
			hasDiskImage = true
		}
		return nil
	})
	if !hasDiskImage {
		errors = append(errors, "no virtual disk files (*.vhd or *.vhdx) found")
	}

	return BackupReport{
		BackupDir: backupDir,
		Valid:     len(errors) == 0,
		Errors:    errors,
		CheckedAt: time.Now().Format(time.RFC3339),
	}
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
