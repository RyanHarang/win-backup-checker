# Setup Guide

This guide walks through setting up the Windows Backup Checker.

---

## Prerequisites

-   **Go 1.19 or later** installed ([Download Go](https://golang.org/dl/))
-   **Git** installed ([Download Git](https://git-scm.com/downloads))
-   (Optional) Email account for notifications

---

## Quick Setup

### 1. Clone the Repository

```bash
git clone https://github.com/RyanHarang/win-backup-checker.git
cd win-backup-checker
```

### 2. Run Setup Script

```bash
chmod +x setup.sh
./setup.sh
```

This will:

-   Generate `config.json` with default settings
-   Generate `email.config.json` template (disabled by default)

### 3. Configure Backup Paths

Edit `configs/config.json` and update the backup paths:

```json
{
  "backup_paths": [
    "/path/to/directory/containing/multiple/backups",
    "/or/path/to/single/backup/directly"
  ],
  ...
}
```

### 4. Test the Configuration

```bash
go run ./cmd/checker/
```

You should see output showing:

-   Discovered backup sets
-   Validation results
-   Summary statistics

---

## Detailed Setup

### Configuration File: `config.json`

Located in `configs/config.json`, this file controls the backup validation behavior:

```json
{
    "backup_paths": ["/path/to/backups"],
    "check_hash": false,
    "deep_validation": true,
    "max_zip_sample_size": 104857600,
    "required_catalog_extensions": [".wbcat", ".cat"],
    "min_backup_age": "1h",
    "max_backup_age": "90d"
}
```

#### Configuration Options

| Option                        | Description                                                                  | Default              |
| ----------------------------- | ---------------------------------------------------------------------------- | -------------------- |
| `backup_paths`                | Array of directory containing backups or backup root directories to validate | Required             |
| `check_hash`                  | Perform hash validation (not implemented yet)                                | `false`              |
| `deep_validation`             | Perform thorough content validation                                          | `true`               |
| `max_zip_sample_size`         | Maximum bytes to read when testing ZIP files                                 | `104857600` (100MB)  |
| `required_catalog_extensions` | Catalog file extensions to look for                                          | `[".wbcat", ".cat"]` |
| `min_backup_age`              | Minimum age before considering backup complete                               | `"1h"`               |
| `max_backup_age`              | Maximum age before warning about old backups                                 | `"90d"`              |

#### Duration Format

-   Hours: `"1h"`, `"24h"`
-   Days: `"7d"`, `"30d"`, `"90d"`
-   Minutes: `"30m"`

---

## Email Notifications (Optional)

Refer to [EMAIL.md](EMAIL.md) for email notification setup.

---

## Usage

### Basic Usage

```bash
# Run validation (writes to logs.json by default)
go run ./cmd/checker/

# Run without logging to file
go run ./cmd/checker/ --no-log

# Run without email notifications
go run ./cmd/checker/ --no-email

# JSON output only (no human-readable output)
go run ./cmd/checker/ --json

# Use more parallel workers (default: 4)
go run ./cmd/checker/ --parallel=8

# Set custom timeout (default: 30 minutes)
go run ./cmd/checker/ --timeout=1h

# Write logs to custom file
go run ./cmd/checker/ --json-out=backup-report.json
```

### Exit Codes

The program returns exit codes based on results:

-   `0` - All backups valid
-   `1` - Some backups invalid (warnings or errors found)
-   `2` - Fatal error (config error, scan failure, or I/O failure)

This allows for integration with scripts and monitoring systems.

---

## Building the Application

To create a standalone executable:

```bash
# Build for your current platform
go build -o backup-checker ./cmd/checker/

# Run the built executable
./backup-checker
```

### Cross-Platform Builds

```bash
# Build for Linux
GOOS=linux GOARCH=amd64 go build -o backup-checker-linux ./cmd/checker/

# Build for macOS
GOOS=darwin GOARCH=amd64 go build -o backup-checker-macos ./cmd/checker/

# Build for Windows
GOOS=windows GOARCH=amd64 go build -o backup-checker.exe ./cmd/checker/
```

---

## Scheduling Automated Checks

### Linux/macOS (cron)

Add to your crontab (`crontab -e`):

```bash
# Run daily at 2 AM
0 2 * * * cd /path/to/win-backup-checker && /usr/local/go/bin/go run ./cmd/checker/ >> /var/log/backup-checker.log 2>&1

# Run every 6 hours
0 */6 * * * cd /path/to/win-backup-checker && /usr/local/go/bin/go run ./cmd/checker/
```

### Linux (systemd timer)

Create `/etc/systemd/system/backup-checker.service`:

```ini
[Unit]
Description=Windows Backup Checker
After=network.target

[Service]
Type=oneshot
User=your-user
WorkingDirectory=/path/to/win-backup-checker
ExecStart=/usr/bin/go run ./cmd/checker/
StandardOutput=journal
StandardError=journal
```

Create `/etc/systemd/system/backup-checker.timer`:

```ini
[Unit]
Description=Run Windows Backup Checker daily
Requires=backup-checker.service

[Timer]
OnCalendar=daily
Persistent=true

[Install]
WantedBy=timers.target
```

Enable and start:

```bash
sudo systemctl daemon-reload
sudo systemctl enable backup-checker.timer
sudo systemctl start backup-checker.timer
```

### Windows (Task Scheduler)

1. Open Task Scheduler
2. Create Basic Task
3. Set trigger (e.g., Daily at 2 AM)
4. Action: Start a program
    - Program: `go.exe`
    - Arguments: `run ./cmd/checker/`
    - Start in: `C:\path\to\win-backup-checker`

---

## Updating

To update to the latest version:

```bash
cd win-backup-checker
git pull origin main
```
