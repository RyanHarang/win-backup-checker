package winbackupchecker

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type EmailConfig struct {
	Enabled        bool     `json:"enabled"`
	SMTPHost       string   `json:"smtp_host"`
	SMTPPort       int      `json:"smtp_port"`
	From           string   `json:"from"`
	To             []string `json:"to"`
	Username       string   `json:"username"`
	Password       string   `json:"password"`
	SendOnSuccess  bool     `json:"send_on_success"`
	SendOnWarnings bool     `json:"send_on_warnings"`
	SendOnErrors   bool     `json:"send_on_errors"`
	SubjectPrefix  string   `json:"subject_prefix"`
}

type Config struct {
	BackupPaths               []string     `json:"backup_paths"`
	CheckHash                 bool         `json:"check_hash"`
	DeepValidation            bool         `json:"deep_validation"`
	MaxZipSampleSize          int64        `json:"max_zip_sample_size"`
	RequiredCatalogExtensions []string     `json:"required_catalog_extensions"`
	MinBackupAge              string       `json:"min_backup_age"`
	MaxBackupAge              string       `json:"max_backup_age"`
	Email                     *EmailConfig `json:"email,omitempty"`
}

// ValidationSeverity represents severity level of validation issues
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

func (s ValidationSeverity) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

// ValidationIssue represents a specific validation problem
type ValidationIssue struct {
	Severity   ValidationSeverity `json:"severity"`
	Message    string             `json:"message"`
	Path       string             `json:"path,omitempty"`
	Suggestion string             `json:"suggestion,omitempty"`
	CheckedAt  string             `json:"checked_at"`
}

// BackupReport represents validation details for single backup folder
type BackupReport struct {
	BackupDir       string            `json:"backup_dir"`
	Valid           bool              `json:"valid"`
	Issues          []ValidationIssue `json:"issues"`
	CheckedAt       string            `json:"checked_at"`
	ValidationStats ValidationStats   `json:"validation_stats"`
}

// ValidationStats provides detailed metrics about validation process
type ValidationStats struct {
	TotalFiles       int        `json:"total_files"`
	ValidatedFiles   int        `json:"validated_files"`
	CorruptFiles     int        `json:"corrupt_files"`
	TotalSize        int64      `json:"total_size_bytes"`
	ValidationTime   string     `json:"validation_time"`
	CatalogFiles     int        `json:"catalog_files"`
	BackupFiles      int        `json:"backup_files"`
	OldestBackupTime *time.Time `json:"oldest_backup_time,omitempty"`
	NewestBackupTime *time.Time `json:"newest_backup_time,omitempty"`
	StructuralChecks int        `json:"structural_checks_passed"`
	ContentChecks    int        `json:"content_checks_passed"`
}

// ScanReport represents results for one root path
type ScanReport struct {
	Root    string         `json:"root"`
	Reports []BackupReport `json:"reports"`
}

// LoadConfig loads JSON config file from given path with defaults
func LoadConfig(path string) (*Config, error) {
	cfg := &Config{
		CheckHash:                 false,
		DeepValidation:            true,
		MaxZipSampleSize:          100 * 1024 * 1024, // 100MB
		RequiredCatalogExtensions: []string{".wbcat"},
		MinBackupAge:              "1h",
		MaxBackupAge:              "90d",
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

// LoadEmailConfig loads email configuration from a separate file
func LoadEmailConfig(path string) (*EmailConfig, error) {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to open email config file: %w", err)
	}
	defer file.Close()

	var emailCfg EmailConfig
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&emailCfg); err != nil {
		return nil, fmt.Errorf("failed to parse email config file: %w", err)
	}

	if emailCfg.Enabled {
		if err := emailCfg.Validate(); err != nil {
			return nil, fmt.Errorf("invalid email config: %w", err)
		}
	}

	return &emailCfg, nil
}

// Validate checks if configuration is valid
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

	if c.Email != nil && c.Email.Enabled {
		if err := c.Email.Validate(); err != nil {
			return fmt.Errorf("invalid email config: %w", err)
		}
	}

	return nil
}

// Validate checks if email configuration is valid
func (e *EmailConfig) Validate() error {
	if e.SMTPHost == "" {
		return fmt.Errorf("smtp_host is required")
	}
	if e.SMTPPort <= 0 || e.SMTPPort > 65535 {
		return fmt.Errorf("smtp_port must be between 1 and 65535")
	}
	if e.From == "" {
		return fmt.Errorf("from address is required")
	}
	if len(e.To) == 0 {
		return fmt.Errorf("at least one recipient (to) is required")
	}
	if e.Username == "" {
		return fmt.Errorf("username is required for SMTP authentication")
	}
	if e.Password == "" {
		return fmt.Errorf("password is required for SMTP authentication")
	}
	return nil
}

// GetMinBackupAge returns parsed minimum backup age duration
func (c *Config) GetMinBackupAge() (time.Duration, error) {
	return parseDuration(c.MinBackupAge)
}

// GetMaxBackupAge returns parsed maximum backup age duration
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
