
# Changelog

This file summarizes notable changes for each released tag. Trivial noise (merge commits, many dependabot lines and minor chore commits) has been removed to keep the history focused.

## v0.1 (2024-06-06)

- MVP and core features: command implementation, get/post/executable tasks, file log writer, compiler for Task.
- Schedulers: initial support for interval, cron and at schedulers.
- Config and validation: initial config structure and config validation added.

## v0.1.1 (2024-06-07)

- Improvements to concurrency and task-runner: concurrency lock, concurrency config field and sync.Locker implementation.
- Cron improvements: optional seconds field, macros support and docs updates.
- Added config schema and miscellaneous optimizations.

## v0.2 (2024-06-12)

- Docker connection: create containers from images and a fully functional docker connection implementation.
- Per-task hooks and dynamic task connections.
- Credential manager + validator and other robustness fixes.

## v0.4 (2024-06-30)

- Integrated cron parser and sanitizer.
- Webserver and HTTP event listener added.
- Notable breaking change: scheduler renamed to event.
- Various fixes and race-condition fixes in concurrent pool and test coverage improvements.

## v0.4.1 — v0.4.5 (2024-07-01 → 2024-07-02)

- Releases focused on Docker-related fixes, CI tweaks, and stability fixes (docker image/ghcr, docker API migration, CI updates).

## v0.5.0 (2024-07-06)

- Metrics/exporters: added basic exporters including command status and event counter exporter.
- Fixes for global context and concurrency behavior.

## v0.6.0-alpha → v0.6.0 (2024-07-13 → 2024-07-29)

- Docker event listener and utilities added.
- Exported/overhauled some functionalities and various dependency bumps for docker-related modules.

## v0.7.x (2024-08-17 → 2024-09-10)

- Improvements: environ key handling, logfile watcher, command event arg mode.
- Multiple dependency updates and bug fixes; numerous small improvements and tests.

## v0.7.4 → v0.7.6 (2025-02-17 → 2025-05-12)

- Tooling updates: Go version bumped to 1.24, multiple dependabot dependency bumps and CI/tooling maintenance.
- Fixes: deadlock and race-condition fixes.

## v0.8.x (2025-05-20 → 2025-07-26)

- Major refactor and system design improvements: generator-based approach, redesigned template engine, command context immutability.
- CI/CD and goreleaser/workflow fixes and simplifications. Docker and template-related fixes.

## v0.9.0 (2025-09-27)

- Template engine: switched to go-tools version and added query params support.
- Various dependency bumps and tooling updates.

## v0.9.1 (2025-11-04)

- Versioning and release tweaks; added version data to CLI command.
- Dependency/tooling bumps (CodeQL, actions, docker, etc.).

## v0.9.2 (2025-11-08)

- Bump: docker to latest version.
- Hotfix: added buffer size to zip channel (10 items per input).
- Feature: added signal handling.
- Fix: concurrent write exception fix.

## v0.10.0 (2025-11-08)

- Feature: per-task variables that can be used in hooks.
- Chore: switch logger to zap logger.
- Fix: remove redundant println in cron registration.
