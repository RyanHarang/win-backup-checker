## Usage:

#### Check file backups (writes to logs.json by default)

```sh
go run ./cmd/checker/
```

#### JSON only output

```sh
go run ./cmd/checker/ --json
```

#### Write to custom file

```sh
go run ./cmd/checker/ --json-out=custom.json
```

#### Don't write to log file

```sh
go run ./cmd/checker/ --no-log
```

#### Use 8 concurrent workers

```sh
go run ./cmd/checker/ --parallel=8
```

#### Set 1 hour timeout

```sh
go run ./cmd/checker/ --timeout=1h
```

#### Disable email notifications

```sh
go run ./cmd/checker/ --no-email
```

---

## Exit codes:

-   0 = all backups valid
-   1 = some backups invalid
-   2 = fatal error (config, scan, or IO failure)
