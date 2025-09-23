// // scan_file.go
// package winbackupchecker

// import (
// 	"archive/zip"
// 	"fmt"
// 	"os"
// 	"path/filepath"
// 	"time"
// )

// func ScanFileBackupDir(root string, quiet bool) (*ScanReport, error) {
// 	if !quiet {
// 		fmt.Println("Scanning file backup root:", root)
// 	}

// 	report := &ScanReport{Root: root, Reports: []BackupReport{}}

// 	// Root must have MediaID.bin
// 	if !fileExists(filepath.Join(root, "MediaID.bin")) {
// 		report.Reports = append(report.Reports, BackupReport{
// 			BackupDir: root,
// 			Valid:     false,
// 			Errors:    []string{"missing MediaID.bin at root"},
// 			CheckedAt: time.Now().Format(time.RFC3339),
// 		})
// 		return report, nil
// 	}

// 	// Scan each machine directory (like RY/)
// 	entries, err := os.ReadDir(root)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to read backup root: %w", err)
// 	}

// 	for _, entry := range entries {
// 		if !entry.IsDir() {
// 			continue
// 		}
// 		machineDir := filepath.Join(root, entry.Name())
// 		backupSets, _ := os.ReadDir(machineDir)
// 		for _, set := range backupSets {
// 			if !set.IsDir() || filepath.Ext(set.Name()) != "" {
// 				continue
// 			}
// 			if !quiet {
// 				fmt.Println("Found backup set:", set.Name())
// 			}
// 			setPath := filepath.Join(machineDir, set.Name())
// 			br := validateFileBackupSet(setPath, quiet)
// 			report.Reports = append(report.Reports, br)
// 		}
// 	}

// 	return report, nil
// }

// func validateFileBackupSet(setDir string, quiet bool) BackupReport {
// 	errors := []string{}

// 	// Check for Catalogs with .wbcat
// 	catalogDir := filepath.Join(setDir, "Catalogs")
// 	if !dirExists(catalogDir) {
// 		errors = append(errors, "missing Catalogs folder")
// 	} else {
// 		hasWbcat := false
// 		filepath.Walk(catalogDir, func(path string, info os.FileInfo, err error) error {
// 			if !info.IsDir() && filepath.Ext(path) == ".wbcat" {
// 				hasWbcat = true
// 			}
// 			return nil
// 		})
// 		if !hasWbcat {
// 			errors = append(errors, "no .wbcat catalog found")
// 		}
// 	}

// 	// Look for Backup Files folder with zips
// 	foundZips := false
// 	filepath.Walk(setDir, func(path string, info os.FileInfo, err error) error {
// 		if !info.IsDir() && filepath.Ext(path) == ".zip" {
// 			foundZips = true
// 			if err := checkZipReadable(path); err != nil {
// 				errors = append(errors, fmt.Sprintf("unreadable zip %s: %v", path, err))
// 			}
// 		}
// 		return nil
// 	})
// 	if !foundZips {
// 		errors = append(errors, "no .zip backup files found")
// 	}

// 	return BackupReport{
// 		BackupDir: setDir,
// 		Valid:     len(errors) == 0,
// 		Errors:    errors,
// 		CheckedAt: time.Now().Format(time.RFC3339),
// 	}
// }

// func checkZipReadable(path string) error {
// 	r, err := zip.OpenReader(path)
// 	if err != nil {
// 		return err
// 	}
// 	defer r.Close()
// 	if len(r.File) > 0 {
// 		f, err := r.File[0].Open()
// 		if err != nil {
// 			return fmt.Errorf("failed to open first entry: %w", err)
// 		}
// 		f.Close()
// 	}
// 	return nil
// }

// // returns true if path exists and is a regular file
// func fileExists(path string) bool {
// 	info, err := os.Stat(path)
// 	if err != nil {
// 		return false
// 	}
// 	return !info.IsDir()
// }

// // returns true if path exists and is a directory
// func dirExists(path string) bool {
// 	info, err := os.Stat(path)
// 	if err != nil {
// 		return false
// 	}
// 	return info.IsDir()
// }

// scan_file.go
package winbackupchecker

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// BackupSetInfo contains metadata about a backup set
type BackupSetInfo struct {
	Path          string
	Size          int64
	FileCount     int
	ModTime       time.Time
	CatalogFiles  []string
	BackupFiles   []string
}

func ScanFileBackupDir(ctx context.Context, root string, maxWorkers int) (*ScanReport, error) {
	fmt.Printf("Scanning file backup root: %s (max workers: %d)\n", root, maxWorkers)

	report := &ScanReport{Root: root, Reports: []BackupReport{}}
	startTime := time.Now()

	// Root must have MediaID.bin
	mediaIDPath := filepath.Join(root, "MediaID.bin")
	if !fileExists(mediaIDPath) {
		issue := NewValidationIssue(SeverityCritical, 
			"missing MediaID.bin at root", 
			root, 
			"ensure the backup root directory is correct and contains MediaID.bin")
		
		report.Reports = append(report.Reports, BackupReport{
			BackupDir: root,
			Valid:     false,
			Errors:    []string{"missing MediaID.bin at root"}, // Keep for backward compatibility
			Issues:    []ValidationIssue{issue},
			CheckedAt: NowRFC3339(),
		})
		return report, nil
	}

	// Validate MediaID.bin
	if err := validateMediaID(mediaIDPath); err != nil {
		issue := NewValidationIssue(SeverityError, 
			fmt.Sprintf("invalid MediaID.bin: %v", err), 
			mediaIDPath, 
			"check if MediaID.bin is corrupted or from a different backup system")
		
		report.Reports = append(report.Reports, BackupReport{
			BackupDir: root,
			Valid:     false,
			Errors:    []string{err.Error()},
			Issues:    []ValidationIssue{issue},
			CheckedAt: NowRFC3339(),
		})
	}

	// Discover backup sets
	backupSets, err := discoverBackupSets(root)
	if err != nil {
		return nil, fmt.Errorf("failed to discover backup sets: %w", err)
	}


	fmt.Printf("Found %d backup sets to validate\n", len(backupSets))

	// Validate backup sets with controlled concurrency
	reports := validateBackupSets(ctx, backupSets, maxWorkers)
	report.Reports = append(report.Reports, reports...)

	fmt.Printf("Completed validation in %v\n", time.Since(startTime))

	return report, nil
}

func discoverBackupSets(root string) ([]BackupSetInfo, error) {
	var backupSets []BackupSetInfo

	// Scan each machine directory (like RY/)
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, fmt.Errorf("failed to read backup root: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() || entry.Name() == "." || entry.Name() == ".." {
			continue
		}

		machineDir := filepath.Join(root, entry.Name())
		backupSetDirs, err := os.ReadDir(machineDir)
		if err != nil {
			continue // Skip problematic machine directories
		}

		for _, setDir := range backupSetDirs {
			if !setDir.IsDir() || filepath.Ext(setDir.Name()) != "" {
				continue
			}

			setPath := filepath.Join(machineDir, setDir.Name())
			info, err := gatherBackupSetInfo(setPath)
			if err != nil {
				// Create a minimal info for failed discovery
				backupSets = append(backupSets, BackupSetInfo{
					Path: setPath,
				})
				continue
			}

			backupSets = append(backupSets, *info)
		}
	}

	// Sort by modification time (newest first)
	sort.Slice(backupSets, func(i, j int) bool {
		return backupSets[i].ModTime.After(backupSets[j].ModTime)
	})

	return backupSets, nil
}

func gatherBackupSetInfo(setPath string) (*BackupSetInfo, error) {
	info := &BackupSetInfo{
		Path:        setPath,
		CatalogFiles: []string{},
		BackupFiles:  []string{},
	}

	err := filepath.Walk(setPath, func(path string, fileInfo os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip problematic files
		}

		if fileInfo.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		
		// Track file counts and sizes
		info.FileCount++
		info.Size += fileInfo.Size()

		// Update modification time to newest file
		if fileInfo.ModTime().After(info.ModTime) {
			info.ModTime = fileInfo.ModTime()
		}

		// Categorize files
		switch ext {
		case ".wbcat", ".cat":
			info.CatalogFiles = append(info.CatalogFiles, path)
		case ".zip":
			info.BackupFiles = append(info.BackupFiles, path)
		}

		return nil
	})

	return info, err
}

func validateBackupSets(ctx context.Context, backupSets []BackupSetInfo, maxWorkers int) []BackupReport {
	if maxWorkers <= 0 {
		maxWorkers = 1
	}

	reports := make([]BackupReport, len(backupSets))
	
	// Create worker pool
	work := make(chan int, len(backupSets))
	var wg sync.WaitGroup

	// Start workers
	for i := 0; i < maxWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case idx, ok := <-work:
					if !ok {
						return
					}
					reports[idx] = validateFileBackupSet(ctx, backupSets[idx])
				case <-ctx.Done():
					return
				}
			}
		}()
	}

	// Queue work
	go func() {
		defer close(work)
		for i := range backupSets {
			select {
			case work <- i:
			case <-ctx.Done():
				return
			}
		}
	}()

	// Wait for completion
	wg.Wait()

	return reports
}

func validateFileBackupSet(ctx context.Context, setInfo BackupSetInfo) BackupReport {
	startTime := time.Now()
	issues := []ValidationIssue{}
	stats := ValidationStats{
		TotalFiles: setInfo.FileCount,
	}

	fmt.Printf("Validating backup set: %s\n", filepath.Base(setInfo.Path))

	// Structural validation
	issues = append(issues, validateBackupStructure(setInfo)...)
	stats.StructuralChecks = countPassedChecks(issues, SeverityCritical, SeverityError)

	// Content validation
	contentIssues, contentStats := validateBackupContent(ctx, setInfo)
	issues = append(issues, contentIssues...)
	stats.ContentChecks = contentStats.ContentChecks
	stats.ValidatedFiles = contentStats.ValidatedFiles
	stats.CorruptFiles = contentStats.CorruptFiles
	
	// Time-based validation
	issues = append(issues, validateBackupAge(setInfo)...)

	// Calculate final stats
	stats.ValidationTime = time.Since(startTime).String()
	stats.TotalSize = setInfo.Size
	stats.CatalogFiles = len(setInfo.CatalogFiles)
	stats.BackupFiles = len(setInfo.BackupFiles)

	if len(setInfo.CatalogFiles) > 0 || len(setInfo.BackupFiles) > 0 {
		stats.OldestBackupTime = &setInfo.ModTime
		stats.NewestBackupTime = &setInfo.ModTime
	}

	// Determine validity (no critical or error issues)
	valid := true
	errorMessages := []string{} // For backward compatibility

	for _, issue := range issues {
		if issue.Severity >= SeverityError {
			valid = false
			errorMessages = append(errorMessages, issue.Message)
		}
	}

	return BackupReport{
		BackupDir:       setInfo.Path,
		Valid:           valid,
		Errors:          errorMessages, // Keep for backward compatibility
		Issues:          issues,
		CheckedAt:       NowRFC3339(),
		ValidationStats: stats,
	}
}

func validateBackupStructure(setInfo BackupSetInfo) []ValidationIssue {
	issues := []ValidationIssue{}

	// Check for catalog directory and files
	catalogDir := filepath.Join(setInfo.Path, "Catalogs")
	if !dirExists(catalogDir) {
		issues = append(issues, NewValidationIssue(SeverityError,
			"missing Catalogs folder",
			catalogDir,
			"backup set should contain a Catalogs folder with .wbcat files"))
	} else if len(setInfo.CatalogFiles) == 0 {
		issues = append(issues, NewValidationIssue(SeverityError,
			"no catalog files found in Catalogs folder",
			catalogDir,
			"ensure the backup completed successfully and catalog files exist"))
	}

	// Check for backup files
	if len(setInfo.BackupFiles) == 0 {
		issues = append(issues, NewValidationIssue(SeverityError,
			"no backup files (.zip) found",
			setInfo.Path,
			"backup set should contain .zip files with the actual backup data"))
	}

	// Check for reasonable file count
	if setInfo.FileCount < 2 {
		issues = append(issues, NewValidationIssue(SeverityWarning,
			fmt.Sprintf("backup set contains only %d files", setInfo.FileCount),
			setInfo.Path,
			"typical backup sets should contain multiple files (catalogs + backup files)"))
	}

	// Check for reasonable size
	if setInfo.Size < 1024 { // Less than 1KB
		issues = append(issues, NewValidationIssue(SeverityWarning,
			fmt.Sprintf("backup set is very small (%d bytes)", setInfo.Size),
			setInfo.Path,
			"backup might be incomplete or corrupted"))
	}

	return issues
}

func validateBackupContent(ctx context.Context, setInfo BackupSetInfo) ([]ValidationIssue, ValidationStats) {
	issues := []ValidationIssue{}
	stats := ValidationStats{}

	// Validate ZIP files
	for _, zipPath := range setInfo.BackupFiles {
		select {
		case <-ctx.Done():
			return issues, stats
		default:
		}

		stats.ValidatedFiles++
		
		if err := validateZipFile(zipPath); err != nil {
			stats.CorruptFiles++
			issues = append(issues, NewValidationIssue(SeverityError,
				fmt.Sprintf("corrupted backup file: %v", err),
				zipPath,
				"backup file may need to be restored from another source"))
		} else {
			stats.ContentChecks++
		}
	}

	// Validate catalog files
	for _, catPath := range setInfo.CatalogFiles {
		select {
		case <-ctx.Done():
			return issues, stats
		default:
		}

		stats.ValidatedFiles++
		
		if err := validateCatalogFile(catPath); err != nil {
			stats.CorruptFiles++
			issues = append(issues, NewValidationIssue(SeverityWarning,
				fmt.Sprintf("catalog file issue: %v", err),
				catPath,
				"catalog may be corrupted but backup data might still be recoverable"))
		} else {
			stats.ContentChecks++
		}
	}

	return issues, stats
}

func validateBackupAge(setInfo BackupSetInfo) []ValidationIssue {
	issues := []ValidationIssue{}
	
	if setInfo.ModTime.IsZero() {
		return issues
	}

	now := time.Now()
	age := now.Sub(setInfo.ModTime)

	// Check if backup is too new (might be in progress)
	if age < time.Hour {
		issues = append(issues, NewValidationIssue(SeverityInfo,
			fmt.Sprintf("backup is very recent (%v old)", age),
			setInfo.Path,
			"backup might still be in progress"))
	}

	// Check if backup is too old
	if age > 90*24*time.Hour { // 90 days
		issues = append(issues, NewValidationIssue(SeverityWarning,
			fmt.Sprintf("backup is quite old (%v)", age),
			setInfo.Path,
			"consider creating more recent backups"))
	}

	return issues
}

func validateZipFile(zipPath string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("cannot open zip: %w", err)
	}
	defer r.Close()

	if len(r.File) == 0 {
		return fmt.Errorf("zip file is empty")
	}

	// Test reading first few files to ensure they're not corrupted
	testCount := minInt(len(r.File), 3)
	for i := 0; i < testCount; i++ {
		file := r.File[i]
		
		rc, err := file.Open()
		if err != nil {
			return fmt.Errorf("cannot open file %s in zip: %w", file.Name, err)
		}

		// Try to read some data
		buffer := make([]byte, minInt64(1024, int64(file.UncompressedSize64)))
		_, err = io.ReadFull(rc, buffer)
		rc.Close()

		if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
			return fmt.Errorf("cannot read file %s in zip: %w", file.Name, err)
		}
	}

	// Check for suspicious zip structure
	if len(r.File) == 1 && r.File[0].UncompressedSize64 == 0 {
		return fmt.Errorf("zip contains only empty file")
	}

	return nil
}

func validateCatalogFile(catPath string) error {
	info, err := os.Stat(catPath)
	if err != nil {
		return fmt.Errorf("cannot stat catalog file: %w", err)
	}

	if info.Size() == 0 {
		return fmt.Errorf("catalog file is empty")
	}

	// Basic file readability check
	file, err := os.Open(catPath)
	if err != nil {
		return fmt.Errorf("cannot open catalog file: %w", err)
	}
	defer file.Close()

	// Read first few bytes to ensure file is readable
	buffer := make([]byte, minInt64(512, info.Size()))
	_, err = file.Read(buffer)
	if err != nil {
		return fmt.Errorf("cannot read catalog file: %w", err)
	}

	return nil
}

func validateMediaID(mediaIDPath string) error {
	info, err := os.Stat(mediaIDPath)
	if err != nil {
		return fmt.Errorf("cannot stat MediaID.bin: %w", err)
	}

	if info.Size() == 0 {
		return fmt.Errorf("MediaID.bin is empty")
	}

	// MediaID.bin should be a reasonable size (typically small)
	if info.Size() > 1024*1024 { // 1MB
		return fmt.Errorf("MediaID.bin is unusually large (%d bytes)", info.Size())
	}

	// Test file readability
	file, err := os.Open(mediaIDPath)
	if err != nil {
		return fmt.Errorf("cannot open MediaID.bin: %w", err)
	}
	defer file.Close()

	buffer := make([]byte, minInt64(256, info.Size()))
	_, err = file.Read(buffer)
	if err != nil {
		return fmt.Errorf("cannot read MediaID.bin: %w", err)
	}

	return nil
}

func countPassedChecks(issues []ValidationIssue, severities ...ValidationSeverity) int {
	severitySet := make(map[ValidationSeverity]bool)
	for _, s := range severities {
		severitySet[s] = true
	}

	failedChecks := 0
	for _, issue := range issues {
		if severitySet[issue.Severity] {
			failedChecks++
		}
	}

	// Assuming we performed some baseline number of structural checks
	baselineChecks := 5 // catalog existence, backup files, file count, size, structure
	return maxInt(0, baselineChecks-failedChecks)
}

// Helper functions
func minInt64(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// fileExists returns true if path exists and is a regular file
func fileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

// dirExists returns true if path exists and is a directory
func dirExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}