# urlprobe

`urlprobe` is a small, fast, concurrent HTTP/HTTPS health checker written in Go. It probes a list of URLs, reports whether each one is healthy, and exits with a non-zero status if any endpoint fails. It is designed for automation: CI/CD gates, post-deploy smoke tests, cron-based monitoring, and quick manual checks.

## The Important DevOps Problem It Solves

Deployments, DNS changes, load balancer updates, certificate renewals, and dependency outages can silently break critical endpoints. Teams need a fast, scriptable way to answer:

> “From this machine/runner, are all my critical URLs returning successful responses within an acceptable time?”

`urlprobe` solves that by providing a single-binary, dependency-free black-box probe. It turns endpoint health into a simple pass/fail signal:

- **Exit 0** = every URL returned a `2xx` response within the timeout.
- **Exit 1** = at least one URL timed out, failed, or returned a non-`2xx` status.

This makes it useful as a **post-deployment verification step**, a **CI/CD quality gate**, or a **cron/Kubernetes CronJob synthetic check**.

## Features

- Concurrent probing with a small worker pool
- Reads URLs from command-line arguments or a text file
- Strict health check: only `2xx` responses are considered healthy
- Per-request timeout: `9s`
- Prints status and latency for each URL
- CI-friendly exit codes
- Uses only the Go standard library

## Requirements

- Go `1.27.2` or later, as declared in `go.mod`
- Network access to the target URLs

## Installation

```bash
git clone https://github.com/samuel-jinadu/urlprobe.git
cd urlprobe
go build -o urlprobe .
```

Or install directly if your Go environment is configured:

```bash
go install github.com/samuel-jinadu/urlprobe@latest
```

## Usage

```bash
urlprobe [flags] [urls...]
```

### Flags

| Flag | Description |
|------|-------------|
| `-file` | Path to a text file containing URLs, one per line. |

### Examples

Probe URLs passed as arguments:

```bash
./urlprobe https://httpbin.org/ https://gobyexample.com/maps
```

Probe URLs from a file:

```bash
./urlprobe -file url.txt
```

Example `url.txt`:

```text
https://httpbin.org/
https://httpbin.org/delay/10
https://gobyexample.com/maps
https://gobyexample.com/command-line-flags
```

## Output

```text
https://httpbin.org/ Healthy 200 OK 143.2ms
https://httpbin.org/delay/10 Unhealthy Timeout 9.000s
https://gobyexample.com/maps Healthy 200 OK 210.5ms
```

Each line contains:

- URL
- `Healthy` or `Unhealthy`
- Reason/status, such as `200 OK` or `Timeout`
- Elapsed time

## Exit Codes

| Code | Meaning |
|------|---------|
| `0` | All URLs were healthy: `2xx` within the timeout. |
| `1` | One or more URLs were unhealthy, timed out, failed, or a file error occurred. |

Example for CI:

```bash
./urlprobe -file production-urls.txt
# CI job fails automatically if any URL is unhealthy
```

## How It Works

1. `urlprobe` collects URLs from either `-file` or command-line arguments.
2. It starts 3 worker goroutines.
3. Each worker receives URLs from a channel and performs an HTTP `GET` request with a `9s` timeout.
4. Only responses with status codes in the `200–299` range are counted as healthy.
5. After all URLs are processed, the program compares the healthy count with the total count.
6. If every URL was healthy, it exits `0`; otherwise it exits `1`.

## DevOps Use Cases

- **Post-deploy smoke test**: verify critical endpoints after a release.
- **CI/CD pipeline gate**: fail the pipeline when a dependency or deployed service is unhealthy.
- **Cron-based synthetic monitoring**: run every minute or hour and alert on non-zero exit.
- **Kubernetes CronJob**: periodically probe external or internal URLs.
- **Manual incident checks**: quickly test a list of URLs from a specific network location.

## Project Layout

```text
.
├── go.mod        # Go module definition
├── main.go       # CLI, worker pool, probing, and reporting logic
├── url.txt       # Example URL list
├── scripts/      # Optional helper scripts
└── LICENSE       # MIT License
```

## License

MIT License. See [LICENSE](./LICENSE) for details.