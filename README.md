## Usage:

- go run ./cmd/checker/ # Check file backups
- go run ./cmd/checker/ --json # JSON only
- go run ./cmd/checker/ --json-out=logs.json # Append JSON report to logs.json
- go run ./cmd/checker/ --parallel=8 # Use 8 concurrent workers
- go run ./cmd/checker/ --timeout=1h # Set 1 hour timeout

## Exit codes:

- 0 = all backups valid
- 1 = some backups invalid
- 2 = fatal error (config, scan, or IO failure)
