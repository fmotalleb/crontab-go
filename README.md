# Cronjob-go

[![Keep a Changelog](https://img.shields.io/badge/changelog-Keep%20a%20Changelog-%23E05735)](CHANGELOG.md)
[![GitHub Release](https://img.shields.io/github/v/release/fmotalleb/crontab-go)](https://github.com/fmotalleb/crontab-go/releases)
[![Go Reference](https://pkg.go.dev/badge/github.com/fmotalleb/crontab-go.svg)](https://pkg.go.dev/github.com/fmotalleb/crontab-go)
[![go.mod](https://img.shields.io/github/go-mod/go-version/fmotalleb/crontab-go)](go.mod)
[![LICENSE](https://img.shields.io/github/license/fmotalleb/crontab-go)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/fmotalleb/crontab-go)](https://goreportcard.com/report/github.com/fmotalleb/crontab-go)
[![Codecov](https://codecov.io/gh/fmotalleb/crontab-go/branch/main/graph/badge.svg)](https://codecov.io/gh/fmotalleb/crontab-go)

⭐ `Star` this repository if you find it valuable and worth maintaining.

👁 `Watch` this repository to get notified about new releases, issues, etc.

## Description

### Cronjob-go: A Robust Cron Scheduler for Docker Environments

Cronjob-go is a powerful, lightweight, and highly configurable Golang application designed to replace the traditional crontab in Docker environments. With its seamless integration and easy-to-use YAML configuration, Cronjob-go simplifies the process of scheduling and managing recurring tasks within your containerized applications.

## Key Features

1. **YAML Configuration**: Cronjob-go leverages YAML for its configuration, making it easy to define and manage your scheduled tasks. The YAML format provides a clean and human-readable way to specify task details, schedules, and other settings.

2. **Robust Scheduling**: The application uses a reliable and flexible scheduling engine to ensure that your tasks are executed on time. It supports a wide range of scheduling patterns, including cron-style expressions, intervals, and custom scheduling logic.

3. **Container-Friendly**: Cronjob-go is designed with Containerization in mind, making it the perfect replacement for crontab in your containerized environments.

4. **Logging and Monitoring**: The application provides comprehensive logging and monitoring capabilities, allowing you to track the execution of your scheduled tasks and quickly identify and resolve any issues that may arise.

5. **OpenTelemetry Observability**: Built-in support for OTLP export of traces, metrics, and logs. Configure via YAML (`observability:` block) or env vars (`OTEL_TRACING_URL`, `OTEL_METRICS_URL`, `OTEL_LOG_URL`). Prometheus `/metrics` endpoint remains available alongside OTLP push metrics.

**Use Cases:**

- **Automated Backups**: Schedule regular backups of your application data or logs to ensure data integrity and disaster recovery.
- **Periodic Maintenance Tasks**: Execute maintenance tasks, such as database optimizations, cache clearance, or system updates, on a scheduled basis.
- **Data Processing and Reporting**: Automate the processing and generation of reports, analytics, or other data-driven tasks.

## Configuration

This section outlines the configuration options available for the application.

**Environment Variables:**

- All environment variables and configuration samples are provided in `.env.example` and `config.example.yaml` files.

**Logging:**

- **Time Format:** The default timestamp format is `2006-01-02T15:04:05.000Z` (Golang's datetime format). You can customize this format using the `LOG_TIMESTAMP_FORMAT` environment variable.
- **Log Format:** The default log format is `ansi` (colorful). You can choose from `ansi` (colorful), `plain` (no colors), and `json` using the `LOG_FORMAT` environment variable.
- **Log File:** Logs can be saved to a file by setting the `LOG_FILE` environment variable.
- **StdOut:** Outputting logs to standard output can be disabled by setting `LOG_STDOUT=false`.
- **Log Level:** The default log level is `info`. You can adjust the level from most verbose to least verbose: `trace`, `debug`, `info`, `warn`, `fatal`, `panic` using the `LOG_LEVEL` environment variable.

**OpenTelemetry Observability:**

Crontab-go supports exporting traces, metrics, and logs to an OTLP-compatible collector (e.g. OpenTelemetry Collector). Each signal is independently configured and optional.

**Web event endpoint:**

When the web server is enabled, configure both `webserver_username` and `webserver_password` for Basic authentication. Unauthenticated access is disabled by default; set `webserver_allow_unauthenticated: true` only when the server is reachable exclusively on a trusted network or behind a TLS proxy. Event triggers accept `POST` requests at `/events/<event-name>/emit`. At most 64 event dispatch requests run concurrently; excess requests receive HTTP 429.

Environment variables:

| Variable | Purpose |
|---|---|
| `OTEL_TRACING_URL` | OTLP tracing endpoint (e.g. `http://collector:4318`) |
| `OTEL_METRICS_URL` | OTLP metrics endpoint |
| `OTEL_LOG_URL` | OTLP logging endpoint |
| `OTEL_TRACING_HEADERS` | Custom headers as JSON (e.g. `{"Authorization":"Bearer ..."}`) |
| `OTEL_METRICS_HEADERS` | Custom headers as JSON |
| `OTEL_LOG_HEADERS` | Custom headers as JSON |
| `OTEL_TRACING_INSECURE` | Skip TLS certificate verification (TLS still negotiated) |
| `OTEL_METRICS_INSECURE` | Skip TLS certificate verification (TLS still negotiated) |
| `OTEL_LOG_INSECURE` | Skip TLS certificate verification (TLS still negotiated) |
| `OTEL_SERVICE_NAME` | Service name resource attribute (default: `crontab-go`) |

Example YAML config:

```yaml
observability:
  service-name: "crontab-go"
  tracing:
    url: "http://collector:4318"
    insecure: true
  metrics:
    url: "http://collector:4318"
    insecure: true
    interval: 60s
  log:
    url: "http://collector:4318"
    insecure: true
```

All three signals are non-fatal: if a collector is unreachable at startup, the corresponding signal is disabled with a warning. The Prometheus `/metrics` endpoint continues to work independently of OTLP metrics export.

**Shell:**

- **Shell:** The application leverages your system's shell to execute commands. The default shell is `sh` for Linux and `cmd` for Windows. You can override the default using the `SHELL` environment variable, which can be set individually for each process.
- **Shell Args:** The default shell arguments are `-c` for `sh` (Linux) and `/c` for `cmd` (Windows). These can be customized using the `SHELL_ARGS` environment variable.

**Configuration File:**

- A fully documented configuration file is available at [config.example.yaml](config.example.yaml).
- You can select config file using `--config (-c)` flag. `crontab-go -c config.example.yaml`
- You can also use [schema.json](/raw/main/schema.json) as schema of config file.

> By adding this line in the `config.yaml` file you can enable the schema.
>
> `# yaml-language-server: $schema=https://github.com/fmotalleb/crontab-go/raw/main/schema.json`

**Task Output:**

Command and HTTP request (get/post) outputs are streamed to stderr (never stdout, never buffered in memory) with a per-line prefix of a unique per-execution id: `<id> | `. For commands, both stdout and stderr are piped into that single stream, and the same id is reported in the task's start and finish log messages so you can correlate log lines with a run's output. For get/post requests, set `insecure: true` on the task to skip TLS certificate verification (useful for self-signed certificates).

## Getting Started

To get started with Cronjob-go, simply download the binary for your platform and configure your scheduled tasks using the provided YAML format. The application's documentation includes detailed instructions on installation, configuration, and usage, making it easy to integrate into your existing Docker-based infrastructure.

## Thanks To

This project was possible thanks to

- [Cobra](https://github.com/spf13/cobra)
  - Cobra is a library for creating powerful modern CLI applications.
- [Viper](https://github.com/spf13/viper)
  - Go configuration with fangs!
- [Cron](https://github.com/robfig/cron)
  - Cron backend!
- [GoDotenv](https://github.com/joho/godotenv)
  - A Go (golang) port of the Ruby dotenv project (which loads env vars from a .env file).
- [Signals](github.com/maniartech/signals)
  - High-Performance, Production-Ready Event System for Go.
