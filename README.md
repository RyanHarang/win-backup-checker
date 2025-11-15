## Usage:

-   `# Check file backups (writes to logs.json by default) \n go run ./cmd/checker/`
-   `# JSON only output \n go run ./cmd/checker/ --json`
-   `# Write to custom file \n go run ./cmd/checker/ --json-out=custom.json`
-   `# Don't write to log file \n go run ./cmd/checker/ --no-log`
-   `# Use 8 concurrent workers \n go run ./cmd/checker/ --parallel=8`
-   `# Set 1 hour timeout \n go run ./cmd/checker/ --timeout=1h`
-   `# Disable email notifications \n go run ./cmd/checker/ --no-email`

## Exit codes:

-   0 = all backups valid
-   1 = some backups invalid
-   2 = fatal error (config, scan, or IO failure)

## Email notifications:

### Gmail:

-   [Set up an app password](https://myaccount.google.com/apppasswords)
-   Use the app password in config
