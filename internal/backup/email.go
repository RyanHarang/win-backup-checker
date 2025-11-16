package winbackupchecker

import (
	"bytes"
	"fmt"
	"html/template"
	"net/smtp"
	"path/filepath"
	"strings"
	"time"
)

type EmailData struct {
	Timestamp   string
	Summary     ScanSummary
	HasErrors   bool
	HasWarnings bool
	Reports     []BackupReport
	ScanRoots   []string
}

type ScanSummary struct {
	TotalBackups   int
	ValidBackups   int
	InvalidBackups int
	FailedScans    int
}

// SendEmailAlert sends an email notification based on the scan results
func SendEmailAlert(cfg *EmailConfig, summary ScanSummary, reports []ScanReport) error {
	if cfg == nil || !cfg.Enabled {
		return nil
	}

	// Determine if we should send based on results
	shouldSend := false
	hasErrors := summary.InvalidBackups > 0 || summary.FailedScans > 0
	hasWarnings := false

	// Check for warnings in reports
	for _, scanReport := range reports {
		for _, backupReport := range scanReport.Reports {
			for _, issue := range backupReport.Issues {
				if issue.Severity == SeverityWarning {
					hasWarnings = true
					break
				}
			}
		}
	}

	if cfg.SendOnErrors && hasErrors {
		shouldSend = true
	} else if cfg.SendOnWarnings && hasWarnings {
		shouldSend = true
	} else if cfg.SendOnSuccess && !hasErrors && !hasWarnings {
		shouldSend = true
	}

	if !shouldSend {
		return nil
	}

	// Prepare email data
	emailData := EmailData{
		Timestamp:   time.Now().Format(time.RFC1123),
		Summary:     summary,
		HasErrors:   hasErrors,
		HasWarnings: hasWarnings,
		ScanRoots:   make([]string, 0),
	}

	// Flatten reports and collect roots
	for _, scanReport := range reports {
		emailData.ScanRoots = append(emailData.ScanRoots, scanReport.Root)
		emailData.Reports = append(emailData.Reports, scanReport.Reports...)
	}

	// Generate subject
	subject := generateSubject(cfg, hasErrors, hasWarnings, summary)

	// Generate email body
	body, err := generateEmailBody(emailData)
	if err != nil {
		return fmt.Errorf("failed to generate email body: %w", err)
	}

	// Send email
	return sendEmail(cfg, subject, body)
}

func generateSubject(cfg *EmailConfig, hasErrors, hasWarnings bool, summary ScanSummary) string {
	prefix := cfg.SubjectPrefix
	if prefix == "" {
		prefix = "[Backup Alert]"
	}

	status := "SUCCESS"
	if hasErrors {
		status = "ERRORS DETECTED"
	} else if hasWarnings {
		status = "WARNINGS"
	}

	return fmt.Sprintf("%s %s - %d/%d Backups Valid", prefix, status, summary.ValidBackups, summary.TotalBackups)
}

func generateEmailBody(data EmailData) (string, error) {
	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .header { background-color: {{if .HasErrors}}#dc3545{{else if .HasWarnings}}#ffc107{{else}}#28a745{{end}}; 
                  color: white; padding: 20px; border-radius: 5px; }
        .summary { background-color: #f8f9fa; padding: 15px; margin: 20px 0; border-radius: 5px; }
        .backup-set { margin: 20px 0; padding: 15px; border: 1px solid #dee2e6; border-radius: 5px; }
        .valid { border-left: 4px solid #28a745; }
        .invalid { border-left: 4px solid #dc3545; }
        .issue { margin: 10px 0; padding: 10px; background-color: #f8f9fa; border-radius: 3px; }
        .critical { border-left: 3px solid #dc3545; }
        .error { border-left: 3px solid #fd7e14; }
        .warning { border-left: 3px solid #ffc107; }
        .info { border-left: 3px solid #17a2b8; }
        .stats { display: grid; grid-template-columns: repeat(2, 1fr); gap: 10px; margin: 10px 0; }
        .stat-item { background-color: #e9ecef; padding: 8px; border-radius: 3px; }
        h2 { color: #495057; border-bottom: 2px solid #dee2e6; padding-bottom: 10px; }
        h3 { color: #6c757d; }
        .path { font-family: monospace; background-color: #e9ecef; padding: 2px 5px; border-radius: 3px; }
    </style>
</head>
<body>
    <div class="header">
        <h1>Backup Validation Report</h1>
        <p>Scan completed at: {{.Timestamp}}</p>
    </div>

    <div class="summary">
        <h2>Summary</h2>
        <div class="stats">
            <div class="stat-item"><strong>Total Backups:</strong> {{.Summary.TotalBackups}}</div>
            <div class="stat-item"><strong>Valid Backups:</strong> {{.Summary.ValidBackups}}</div>
            <div class="stat-item"><strong>Invalid Backups:</strong> {{.Summary.InvalidBackups}}</div>
            <div class="stat-item"><strong>Failed Scans:</strong> {{.Summary.FailedScans}}</div>
        </div>
        {{if gt .Summary.TotalBackups 0}}
        <p><strong>Success Rate:</strong> {{printf "%.1f" (div (mul (float64 .Summary.ValidBackups) 100.0) (float64 .Summary.TotalBackups))}}%</p>
        {{end}}
    </div>

    <h2>Scanned Paths</h2>
    <ul>
    {{range .ScanRoots}}
        <li><span class="path">{{.}}</span></li>
    {{end}}
    </ul>

    <h2>Backup Details</h2>
    {{range .Reports}}
    <div class="backup-set {{if .Valid}}valid{{else}}invalid{{end}}">
        <h3>{{base .BackupDir}}</h3>
        <p><span class="path">{{.BackupDir}}</span></p>
        <p><strong>Status:</strong> {{if .Valid}}✅ Valid{{else}}❌ Invalid{{end}}</p>
        
        <div class="stats">
            <div class="stat-item"><strong>Total Files:</strong> {{.ValidationStats.TotalFiles}}</div>
            <div class="stat-item"><strong>Validated:</strong> {{.ValidationStats.ValidatedFiles}}</div>
            <div class="stat-item"><strong>Corrupt:</strong> {{.ValidationStats.CorruptFiles}}</div>
            <div class="stat-item"><strong>Total Size:</strong> {{formatBytes .ValidationStats.TotalSize}}</div>
            <div class="stat-item"><strong>Catalog Files:</strong> {{.ValidationStats.CatalogFiles}}</div>
            <div class="stat-item"><strong>Backup Files:</strong> {{.ValidationStats.BackupFiles}}</div>
        </div>

        {{if .Issues}}
        <h4>Issues Found ({{len .Issues}})</h4>
        {{range .Issues}}
        <div class="issue {{severityClass .Severity}}">
            <strong>{{severityString .Severity}}:</strong> {{.Message}}<br>
            {{if .Path}}<span class="path">{{.Path}}</span><br>{{end}}
            {{if .Suggestion}}<em>Suggestion: {{.Suggestion}}</em>{{end}}
        </div>
        {{end}}
        {{end}}
    </div>
    {{end}}

    <hr>
    <p style="color: #6c757d; font-size: 0.9em;">
        This is an automated message from the Windows Backup Checker system.
    </p>
</body>
</html>
`

	funcMap := template.FuncMap{
		"base":  filepath.Base,
		"upper": strings.ToUpper,
		"severityString": func(s ValidationSeverity) string {
			return strings.ToUpper(s.String())
		},
		"severityClass": func(s ValidationSeverity) string {
			return s.String()
		},
		"formatBytes": func(b int64) string {
			const unit = 1024
			if b < unit {
				return fmt.Sprintf("%d B", b)
			}
			div, exp := int64(unit), 0
			for n := b / unit; n >= unit; n /= unit {
				div *= unit
				exp++
			}
			return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
		},
		"float64": func(i int) float64 { return float64(i) },
		"mul":     func(a, b float64) float64 { return a * b },
		"div": func(a, b float64) float64 {
			if b == 0 {
				return 0
			}
			return a / b
		},
	}

	t, err := template.New("email").Funcs(funcMap).Parse(tmpl)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func sendEmail(cfg *EmailConfig, subject, htmlBody string) error {
	// Setup authentication
	auth := smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.SMTPHost)

	// Compose message
	headers := make(map[string]string)
	headers["From"] = cfg.From
	headers["To"] = strings.Join(cfg.To, ", ")
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=UTF-8"

	message := ""
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + htmlBody

	// Send email
	addr := fmt.Sprintf("%s:%d", cfg.SMTPHost, cfg.SMTPPort)
	err := smtp.SendMail(addr, auth, cfg.From, cfg.To, []byte(message))
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}
