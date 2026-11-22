# nginxoid

[![Go](https://img.shields.io/badge/Go-1.27.1+-00ADD8?style=flat&logo=go&logoColor=white)](https://go.dev/)
[![Linter](https://img.shields.io/badge/Linter-golangci--lint-20B2AA?style=flat)](https://golangci-lint.run/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

A fast, lightweight CLI utility written in Go for streaming NGINX access log analysis and generating structured **Markdown** and **JSON** statistical reports.

---

## Overview

`nginxoid` parses standard NGINX combined access logs from local files, glob patterns, or remote HTTP/HTTPS endpoints without loading entire files into memory. It filters traffic by date range, aggregates latency-relevant payload sizes (`average`, `max`, `p95`), identifies the most requested resources, and exports clean, deterministic reports while protecting existing files from accidental overwrites.

---

## Features

- **Flexible Log Ingestion**: Read from a single file, expand filesystem glob patterns (e.g., `logs/*.log`), or stream logs directly over `http://` and `https://`.
- **Streaming Architecture**: Processes logs line-by-line via `bufio.Scanner` (up to 1 MB per line) and `io.MultiReader` for low memory overhead.
- **Date Range Filtering**: Filter log entries using inclusive `--from` and `--to` date bounds (`YYYY-MM-DD`).
- **Traffic & Payload Analytics**:
    - Total processed lines, skipped malformed lines (with `stderr` warnings), and matched requests.
    - Response body size metrics in bytes: **Average**, **Max**, and **95th Percentile (`p95`)**.
    - **Top 10** most requested endpoints (sorted by hit count, then lexicographically).
    - HTTP response status code breakdown.
- **Dual Report Formats**: Export results as human-readable **Markdown** tables or machine-readable **JSON**.
- **Safe File Output**: Uses exclusive file creation (`O_CREATE | O_EXCL`) to prevent overwriting existing report files.

---

## Project Structure

```text
nginxoid/
├── cmd/
│   ├── nginxoid/        # Main CLI entrypoint and exit code handling
│   └── log-server/      # Mock HTTP server for local integration testing
├── internal/
│   ├── analyzer/        # Stream processing, date filtering, and metrics aggregation
│   ├── config/          # CLI flag parsing and parameter validation
│   ├── parser/          # NGINX combined log format regex parser
│   ├── report/          # JSON & Markdown rendering and exclusive file writer
│   ├── source/          # Local file, glob, and HTTP/HTTPS stream readers
│   └── timeutil/        # Shared date/time layouts and day-boundary helpers
├── sample_logs/         # Sample valid and mixed NGINX log fixtures
├── .golangci.yml        # Linter configuration
└── Makefile             # Build, test, and lint automation targets
```

---

## Installation

### Using `go install`

```bash
go install [github.com/zhkvsky/nginxoid/cmd/nginxoid@latest](https://github.com/zhkvsky/nginxoid/cmd/nginxoid@latest)
```

### Building from Source

```bash
git clone [https://github.com/zhkvsky/nginxoid.git](https://github.com/zhkvsky/nginxoid.git)
cd nginxoid
go build -o bin/nginxoid ./cmd/nginxoid
```

---

## Usage

### CLI Flags

| Flag | Shorthand | Required | Default | Description |
| :--- | :--- | :---: | :--- | :--- |
| `--path` | `-p` | Yes | — | Path to a `.log`/`.txt` file, a glob pattern, or an `http(s)://` URL |
| `--output` | `-o` | Yes | — | Destination report path (must end with `.md` or `.json` matching `--format`) |
| `--format` | `-f` | No | `markdown` | Output report format: `markdown` or `json` |
| `--from` | — | No | — | Inclusive start date filter in `YYYY-MM-DD` format |
| `--to` | — | No | — | Inclusive end date filter in `YYYY-MM-DD` format (through `23:59:59`) |

### Examples

**1. Analyze a local log file and generate a Markdown report:**

```bash
nginxoid -p sample_logs/access.txt -o report.md
```

**2. Analyze multiple files via glob pattern with date filtering and export to JSON:**

```bash
nginxoid \
  --path "logs/access-*.log" \
  --format json \
  --output report.json \
  --from 2026-11-01 \
  --to 2026-11-15
```

**3. Stream and analyze logs from a remote HTTP endpoint:**

```bash
nginxoid -p [http://127.0.0.1:8000/access.txt](http://127.0.0.1:8000/access.txt) -f markdown -o remote-report.md
```

---

## Supported Log Format

`nginxoid` expects the standard NGINX `combined` log format:

```text
$remote_addr - $remote_user [$time_local] "$request" $status $body_bytes_sent "$http_referer" "$http_user_agent"
```

Example entry:

```text
93.180.71.3 - - [17/May/2015:08:05:32 +0000] "GET /downloads/product_1 HTTP/1.1" 304 0 "-" "Debian APT-HTTP/1.3 (0.8.16~exp12ubuntu10.21)"
```

Malformed lines do not abort execution—they are counted separately, reported to `stderr` as warnings, and excluded from request metrics.

---

## Sample Output

### Markdown (`--format markdown`)

```markdown
# Log Analysis Report

- Files count: 1
- Processed lines: 10
- Malformed lines: 1
- Total requests: 9

## Processed Files

- sample_logs/mixed.txt

## Response Size (in bytes)

- Average: 412.44
- Max: 1024
- 95th percentile (p95): 1024

## Requested Resources

| Resource | Requests Count |
| --- | --- |
| /api/v1/users | 5 |
| /health | 4 |

## Response Codes

| Status Code | Count |
| --- | --- |
| 200 OK | 7 |
| 404 Not Found | 2 |
```

### JSON (`--format json`)

```json
{
  "files": [
    "sample_logs/mixed.txt"
  ],
  "filesCount": 1,
  "processedLinesCount": 10,
  "malformedLinesCount": 1,
  "totalRequestsCount": 9,
  "responseSizeInBytes": {
    "average": 412.44,
    "max": 1024,
    "p95": 1024
  },
  "resources": [
    {
      "resource": "/api/v1/users",
      "totalRequestsCount": 5
    },
    {
      "resource": "/health",
      "totalRequestsCount": 4
    }
  ],
  "responseCodes": {
    "200": 7,
    "404": 2
  }
}
```

---

## Exit Codes

| Code | Meaning |
| :---: | :--- |
| `0` | Report generated and saved successfully |
| `1` | Runtime failure (stream read error, I/O failure, or report serialization error) |
| `2` | Invalid CLI flags/configuration, log source not found, or output file already exists |

---

## Development & Testing

**Run all unit tests with race detection:**

```bash
go test -race -cover ./...
```

**Run the linter:**

```bash
golangci-lint run
```

**Test HTTP streaming locally using the built-in mock log server:**

```bash
# Terminal 1: start static file server on 127.0.0.1:8000
go run ./cmd/log-server -port 8000

# Terminal 2: run nginxoid against the mock endpoint
go run ./cmd/nginxoid -p [http://127.0.0.1:8000/access.txt](http://127.0.0.1:8000/access.txt) -o report.md
```

---

## License

This project is licensed under the terms specified in the [LICENSE](LICENSE) file.