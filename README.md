# Cronjob-go

[![Keep a Changelog](https://img.shields.io/badge/changelog-Keep%20a%20Changelog-%23E05735)](CHANGELOG.md)
[![GitHub Release](https://img.shields.io/github/v/release/FMotalleb/crontab-go)](https://github.com/fmotalleb/crontab-go/releases)
[![Go Reference](https://pkg.go.dev/badge/github.com/fmotalleb/crontab-go.svg)](https://pkg.go.dev/github.com/fmotalleb/crontab-go)
[![go.mod](https://img.shields.io/github/go-mod/go-version/FMotalleb/crontab-go)](go.mod)
[![LICENSE](https://img.shields.io/github/license/FMotalleb/crontab-go)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/fmotalleb/crontab-go)](https://goreportcard.com/report/github.com/fmotalleb/crontab-go)
[![Codecov](https://codecov.io/gh/FMotalleb/crontab-go/branch/main/graph/badge.svg)](https://codecov.io/gh/FMotalleb/crontab-go)

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
