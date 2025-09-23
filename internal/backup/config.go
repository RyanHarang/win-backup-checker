// // config.go
// package winbackupchecker

// import (
// 	"encoding/json"
// 	"fmt"
// 	"os"
// 	"time"
// )

// type Config struct {
// 	BackupPaths []string `json:"backup_paths"`
// 	CheckHash   bool     `json:"check_hash"`
// }

// // BackupReport represents validation details for single backup folder
// type BackupReport struct {
// 	BackupDir string   `json:"backup_dir"`
// 	Valid     bool     `json:"valid"`
// 	Errors    []string `json:"errors"`
// 	CheckedAt string   `json:"checked_at"`
// }

// // ScanReport represents results for one root path
// type ScanReport struct {
// 	Root    string         `json:"root"`
// 	Reports []BackupReport `json:"reports"`
// }

// // LoadConfig loads JSON config file from given path
// func LoadConfig(path string) (*Config, error) {
// 	file, err := os.Open(path)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to open config file: %w", err)
// 	}
// 	defer file.Close()

// 	decoder := json.NewDecoder(file)
// 	var cfg Config
// 	if err := decoder.Decode(&cfg); err != nil {
// 		return nil, fmt.Errorf("failed to parse config file: %w", err)
// 	}

// 	if len(cfg.BackupPaths) == 0 {
// 		return nil, fmt.Errorf("no backup paths specified in config")
// 	}

// 	return &cfg, nil
// }

// // Helper to create timestamp consistently
// func NowRFC3339() string {
// 	return time.Now().Format(time.RFC3339)
// }

// config.go
package winbackupchecker

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type Config struct {
	BackupPaths              []string `json:"backup_paths"`
	CheckHash                bool     `json:"check_hash"`
	DeepValidation           bool     `json:"deep_validation"`
	MaxZipSampleSize         int64    `json:"max_zip_sample_size"`
	RequiredCatalogExtensions []string `json:"required_catalog_extensions"`
	MinBackupAge             string   `json:"min_backup_age"` // Duration string like "24h"
	MaxBackupAge             string   `json:"max_backup_age"` // Duration string like "30d"
}

// ValidationSeverity represents the severity level of validation issues
type ValidationSeverity int

const (
	SeverityInfo ValidationSeverity = iota
	SeverityWarning
	SeverityError
	SeverityCritical
)

func (s ValidationSeverity) String() string {
	switch s {
	case SeverityInfo:
		return "info"
	case SeverityWarning:
		return "warning"
	case SeverityError:
		return "error"
	case SeverityCritical:
		return "critical"
	default:
		return "unknown"
	}
}

// ValidationIssue represents a specific validation problem
type ValidationIssue struct {
	Severity    ValidationSeverity `json:"severity"`
	Message     string             `json:"message"`
	Path        string             `json:"path,omitempty"`
	Suggestion  string             `json:"suggestion,omitempty"`
	CheckedAt   string             `json:"checked_at"`
}

// BackupReport represents validation details for single backup folder
type BackupReport struct {
	BackupDir       string             `json:"backup_dir"`
	Valid           bool               `json:"valid"`
	Errors          []string           `json:"errors"`          // Deprecated: use Issues instead
	Issues          []ValidationIssue  `json:"issues"`
	CheckedAt       string             `json:"checked_at"`
	ValidationStats ValidationStats    `json:"validation_stats"`
}

// ValidationStats provides detailed metrics about the validation process
type ValidationStats struct {
	TotalFiles        int           `json:"total_files"`
	ValidatedFiles    int           `json:"validated_files"`
	CorruptFiles      int           `json:"corrupt_files"`
	TotalSize         int64         `json:"total_size_bytes"`
	ValidationTime    string        `json:"validation_time"`
	CatalogFiles      int           `json:"catalog_files"`
	BackupFiles       int           `json:"backup_files"`
	OldestBackupTime  *time.Time    `json:"oldest_backup_time,omitempty"`
	NewestBackupTime  *time.Time    `json:"newest_backup_time,omitempty"`
	StructuralChecks  int           `json:"structural_checks_passed"`
	ContentChecks     int           `json:"content_checks_passed"`
}

// ScanReport represents results for one root path
type ScanReport struct {
	Root    string         `json:"root"`
	Reports []BackupReport `json:"reports"`
}

// LoadConfig loads JSON config file from given path with defaults
func LoadConfig(path string) (*Config, error) {
	// Set defaults
	cfg := &Config{
		CheckHash:                 false,
		DeepValidation:           true,
		MaxZipSampleSize:         100 * 1024 * 1024, // 100MB
		RequiredCatalogExtensions: []string{".wbcat"},
		MinBackupAge:             "1h",  // Backups should be at least 1 hour old
		MaxBackupAge:             "90d", // Warn if backups are older than 90 days
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return cfg, nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if len(c.BackupPaths) == 0 {
		return fmt.Errorf("no backup paths specified in config")
	}

	// Validate duration strings
	if c.MinBackupAge != "" {
		if _, err := parseDuration(c.MinBackupAge); err != nil {
			return fmt.Errorf("invalid min_backup_age duration: %w", err)
		}
	}

	if c.MaxBackupAge != "" {
		if _, err := parseDuration(c.MaxBackupAge); err != nil {
			return fmt.Errorf("invalid max_backup_age duration: %w", err)
		}
	}

	if c.MaxZipSampleSize < 0 {
		return fmt.Errorf("max_zip_sample_size cannot be negative")
	}

	return nil
}

// GetMinBackupAge returns the parsed minimum backup age duration
func (c *Config) GetMinBackupAge() (time.Duration, error) {
	return parseDuration(c.MinBackupAge)
}

// GetMaxBackupAge returns the parsed maximum backup age duration
func (c *Config) GetMaxBackupAge() (time.Duration, error) {
	return parseDuration(c.MaxBackupAge)
}

// parseDuration handles common duration suffixes
func parseDuration(s string) (time.Duration, error) {
	if s == "" {
		return 0, nil
	}

	// Handle day suffix
	if len(s) > 1 && s[len(s)-1] == 'd' {
		days := s[:len(s)-1]
		d, err := time.ParseDuration(days + "h")
		if err != nil {
			return 0, err
		}
		return d * 24, nil
	}

	return time.ParseDuration(s)
}

// Helper to create timestamp consistently
func NowRFC3339() string {
	return time.Now().Format(time.RFC3339)
}

// NewValidationIssue creates a new validation issue with timestamp
func NewValidationIssue(severity ValidationSeverity, message, path, suggestion string) ValidationIssue {
	return ValidationIssue{
		Severity:   severity,
		Message:    message,
		Path:       path,
		Suggestion: suggestion,
		CheckedAt:  NowRFC3339(),
	}
}