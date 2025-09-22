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

	// Root must have MediaID.bin
	if !fileExists(filepath.Join(root, "MediaID.bin")) {
		report.Reports = append(report.Reports, BackupReport{
			BackupDir: root,
			Valid:     false,
			Errors:    []string{"missing MediaID.bin at root"},
			CheckedAt: time.Now().Format(time.RFC3339),
		})
		return report, nil
	}

	// Scan each machine directory (like RY/)
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, fmt.Errorf("failed to read backup root: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		machineDir := filepath.Join(root, entry.Name())
		backupSets, _ := os.ReadDir(machineDir)
		for _, set := range backupSets {
			if !set.IsDir() || filepath.Ext(set.Name()) != "" {
				continue
			}
			if !quiet {
				fmt.Println("Found backup set:", set.Name())
			}
			setPath := filepath.Join(machineDir, set.Name())
			br := validateFileBackupSet(setPath, quiet)
			report.Reports = append(report.Reports, br)
		}
	}

	return report, nil
}

func validateFileBackupSet(setDir string, quiet bool) BackupReport {
	errors := []string{}

	// Check for Catalogs with .wbcat
	catalogDir := filepath.Join(setDir, "Catalogs")
	if !dirExists(catalogDir) {
		errors = append(errors, "missing Catalogs folder")
	} else {
		hasWbcat := false
		filepath.Walk(catalogDir, func(path string, info os.FileInfo, err error) error {
			if !info.IsDir() && filepath.Ext(path) == ".wbcat" {
				hasWbcat = true
			}
			return nil
		})
		if !hasWbcat {
			errors = append(errors, "no .wbcat catalog found")
		}
	}

	// Look for Backup Files folder with zips
	foundZips := false
	filepath.Walk(setDir, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() && filepath.Ext(path) == ".zip" {
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
		BackupDir: setDir,
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
	if len(r.File) > 0 {
		f, err := r.File[0].Open()
		if err != nil {
			return fmt.Errorf("failed to open first entry: %w", err)
		}
		f.Close()
	}
	return nil
}

// returns true if path exists and is a regular file
func fileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

// returns true if path exists and is a directory
func dirExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}
